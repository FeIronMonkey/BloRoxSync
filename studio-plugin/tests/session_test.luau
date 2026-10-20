--> L4 plugin session tests (Lune). Mock HTTP; no Roblox Studio.

local sessionModule
local sessionLoadErr
do
	local ok, result = pcall(function()
		return require("../src/session")
	end)
	if ok then
		sessionModule = result
	else
		sessionLoadErr = result
	end
end

local function session()
	if sessionLoadErr ~= nil then
		error("studio-plugin/src/session.luau: " .. tostring(sessionLoadErr))
	end
	return sessionModule
end

local function fail(msg: string)
	error(msg, 2)
end

local function assertEq(got: any, want: any, msg: string)
	if got ~= want then
		fail(string.format("%s: got %s, want %s", msg, tostring(got), tostring(want)))
	end
end

local function assertTrue(cond: boolean, msg: string)
	if not cond then
		fail(msg)
	end
end

local function pathOf(call: any): string
	return call.path or ""
end

local function countPath(calls: { any }, path: string): number
	local n = 0
	for _, call in calls do
		if pathOf(call) == path then
			n += 1
		end
	end
	return n
end

local function lastCall(calls: { any }, path: string): any?
	local found = nil
	for _, call in calls do
		if pathOf(call) == path then
			found = call
		end
	end
	return found
end

local function newClock()
	local t = 0
	return {
		nowMs = function()
			return t
		end,
		advance = function(ms: number)
			t += ms
		end,
	}
end

local function newStudio(opts: any)
	opts = opts or {}
	local mapped = opts.mapped
		or {
			{ nodeId = "server/Foo", className = "Script", source = "print(1)\n" },
		}
	local unresolved = opts.unresolved or {}
	local openTabs = opts.openTabs or {}
	local instances = opts.instances or {}
	local log = {
		applied = {},
		destroyed = {},
		created = {},
		writes = {},
		reads = {},
		listeners = 0,
		warningsFromStudio = {},
	}
	return {
		log = log,
		mapped = mapped,
		unmapped = opts.unmapped or { { name = "Part", className = "Part" } },
		duplicateLuaLuau = opts.duplicateLuaLuau == true,
		instances = instances,
		resolveRoot = function(path: string): boolean
			return unresolved[path] ~= true
		end,
		mappedNodes = function(): { any }
			return mapped
		end,
		mappedNodesForRoot = function(rootId: string): { any }
			local out = {}
			for _, n in mapped do
				if n.nodeId == rootId or string.sub(n.nodeId, 1, #rootId + 1) == rootId .. "/" then
					table.insert(out, n)
				end
			end
			return out
		end,
		applyNodes = function(nodes: { any })
			table.insert(log.applied, nodes)
			mapped = nodes
		end,
		startListeners = function()
			log.listeners += 1
		end,
		isEditorOpen = function(nodeId: string): boolean
			return openTabs[nodeId] == true
		end,
		getEditorSource = function(nodeId: string): string
			table.insert(log.reads, { nodeId = nodeId, via = "editor" })
			for _, n in mapped do
				if n.nodeId == nodeId then
					return n.source or ""
				end
			end
			return ""
		end,
		updateEditorSource = function(nodeId: string, source: string)
			table.insert(log.writes, { nodeId = nodeId, via = "editor", source = source })
			for _, n in mapped do
				if n.nodeId == nodeId then
					n.source = source
				end
			end
		end,
		getInstanceSource = function(nodeId: string): string
			table.insert(log.reads, { nodeId = nodeId, via = "source" })
			for _, n in mapped do
				if n.nodeId == nodeId then
					return n.source or ""
				end
			end
			return ""
		end,
		setInstanceSource = function(nodeId: string, source: string)
			table.insert(log.writes, { nodeId = nodeId, via = "source", source = source })
			for _, n in mapped do
				if n.nodeId == nodeId then
					n.source = source
				end
			end
		end,
		getMapped = function(nodeId: string): any?
			for _, n in mapped do
				if n.nodeId == nodeId then
					return n
				end
			end
			return nil
		end,
		createInstance = function(nodeId: string, className: string)
			table.insert(log.created, { nodeId = nodeId, className = className })
			instances[nodeId] = { className = className, mapped = true }
		end,
		destroyInstance = function(nodeId: string)
			table.insert(log.destroyed, nodeId)
			instances[nodeId] = nil
		end,
		hasDuplicateLuaLuau = function(): boolean
			return opts.duplicateLuaLuau == true
		end,
	}
end

local function newHttp(opts: any)
	opts = opts or {}
	local calls = {}
	local streamAttempts = { 0 }
	local streamShouldFail = opts.streamShouldFail == true
	local applyHandler = opts.applyHandler
	return {
		calls = calls,
		streamAttempts = streamAttempts,
		request = function(req: any)
			table.insert(calls, req)
			local path = req.path
			if path == "/v1/health" then
				if opts.healthFail then
					error("connection refused")
				end
				if opts.healthStatus then
					return { status = opts.healthStatus, body = opts.healthBody or {} }
				end
				return { status = 200, body = { ok = true, protocolVersion = 1 } }
			end
			if path == "/v1/session" then
				if opts.sessionStatus then
					return {
						status = opts.sessionStatus,
						body = opts.sessionBody or {
							error = { code = opts.sessionCode, message = opts.sessionMessage },
						},
					}
				end
				return {
					status = 200,
					body = opts.sessionOkBody or {
						protocolVersion = 1,
						agentVersion = "0.0.0",
						projectName = "example",
						roots = { { id = "server", studio = "ServerScriptService" } },
					},
				}
			end
			if path == "/v1/bootstrap/compare" then
				return { status = 200, body = { equal = opts.equal ~= false } }
			end
			if path == "/v1/bootstrap/apply" then
				if opts.applyBootstrap then
					return { status = 200, body = opts.applyBootstrap(req.body) }
				end
				return {
					status = 200,
					body = {
						nodes = {
							{
								nodeId = "server/Foo",
								className = "Script",
								source = "print(1)\n",
								revision = 1,
							},
						},
						applied = { { nodeId = "server/Foo", revision = 1 } },
					},
				}
			end
			if path == "/v1/snapshot" then
				return {
					status = 200,
					body = opts.snapshotBody or {
						nodes = {
							{
								nodeId = "server/Foo",
								className = "Script",
								source = "print(disk)\n",
								revision = 1,
							},
						},
					},
				}
			end
			if path == "/v1/apply" then
				if applyHandler then
					return applyHandler(req.body)
				end
				return { status = 200, body = { applied = { { nodeId = "server/Foo", revision = 2 } } } }
			end
			if path == "/v1/changes" then
				return { status = 200, body = { cursor = opts.changeCursor or 0, changes = opts.changes or {} } }
			end
			return { status = 404, body = { error = { code = "NOT_FOUND", message = "not found" } } }
		end,
		openStream = function(_urlPath: string): (boolean, string?)
			streamAttempts[1] += 1
			if streamShouldFail then
				return false, "stream unavailable"
			end
			return true, nil
		end,
	}
end

local function connect(overrides: any)
	overrides = overrides or {}
	local clock = overrides.clock or newClock()
	local studio = overrides.studio or newStudio(overrides.studioOpts)
	local http = overrides.http or newHttp(overrides.httpOpts)
	local result = session().connect({
		host = overrides.host or "127.0.0.1",
		port = overrides.port or 29850,
		token = overrides.token or "tok",
		pluginVersion = overrides.pluginVersion or "0.0.0",
		http = http,
		studio = studio,
		clock = clock,
	})
	return result, http, studio, clock
end

local tests = {}

local function add(name: string, fn: () -> ())
	table.insert(tests, { name = name, fn = fn })
end

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
add("PluginSessionShowsNotReachable_WhenHealthFails", function()
	local sess, http = connect({ httpOpts = { healthFail = true } })
	assertEq(sess.status, "rbx-sync not reachable", "status")
	assertEq(sess.live, false, "live")
	assertEq(#http.calls, 1, "only health")
	assertEq(pathOf(http.calls[1]), "/v1/health", "health path")
	assertEq(http.calls[1].queryToken, nil, "health has no token")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#platform
add("PluginSessionMakesNoFurtherCalls_WhenHealthFails", function()
	local sess, http = connect({ httpOpts = { healthFail = true } })
	assertEq(sess.live, false, "live")
	assertEq(countPath(http.calls, "/v1/session"), 0, "no session")
	assertEq(countPath(http.calls, "/v1/bootstrap/compare"), 0, "no compare")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#error-handling
add("PluginSessionShowsErrorMessage_WhenSessionUnauthorized", function()
	local sess, http = connect({
		httpOpts = {
			sessionStatus = 401,
			sessionCode = "UNAUTHORIZED",
			sessionMessage = "token rejected",
		},
	})
	assertEq(sess.status, "token rejected", "status shows message")
	assertEq(sess.live, false, "live")
	assertEq(countPath(http.calls, "/v1/bootstrap/compare"), 0, "stopped")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
add("PluginSessionShowsErrorMessage_WhenProtocolMismatch", function()
	local sess, http = connect({
		httpOpts = {
			sessionStatus = 409,
			sessionCode = "PROTOCOL_MISMATCH",
			sessionMessage = "protocol 1 required",
		},
	})
	assertEq(sess.status, "protocol 1 required", "status shows message")
	assertEq(sess.live, false, "live")
	assertEq(countPath(http.calls, "/v1/bootstrap/compare"), 0, "stopped")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#http
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d4--token-in-the-query-string
add("PluginSessionPutsTokenInQuery_WhenAuthenticatedRoute", function()
	local _, http = connect({ httpOpts = { equal = true } })
	local sessCall = lastCall(http.calls, "/v1/session")
	assertTrue(sessCall ~= nil, "session call")
	assertEq(sessCall.queryToken, "tok", "token in query")
	assertEq(sessCall.headers["accessToken"], nil, "no custom token header")
	assertEq(sessCall.body.protocolVersion, 1, "protocolVersion")
	assertEq(sessCall.body.pluginVersion, "0.0.0", "pluginVersion")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#revision-initialization
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
add("PluginSessionSkipsBootstrapAndStoresRevisions_WhenCompareEqual", function()
	local sess, http = connect({
		httpOpts = {
			equal = true,
			snapshotBody = {
				nodes = {
					{ nodeId = "server/Foo", className = "Script", source = "print(1)\n", revision = 1 },
				},
			},
		},
	})
	assertEq(sess.live, true, "live")
	assertEq(sess.awaitingBootstrap, false, "no widget")
	assertEq(countPath(http.calls, "/v1/bootstrap/apply"), 0, "no apply")
	assertEq(countPath(http.calls, "/v1/snapshot"), 1, "snapshot")
	assertEq(sess.revisions["server/Foo"], 1, "revision stored")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d6--explicit-bootstrap-when-trees-differ
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
add("PluginSessionRequiresChoiceAndBlocksLive_WhenCompareUnequal", function()
	local sess, http, studio = connect({ httpOpts = { equal = false } })
	assertEq(sess.awaitingBootstrap, true, "widget required")
	assertEq(sess.live, false, "not live")
	assertEq(studio.log.listeners, 0, "no listeners")
	assertEq(countPath(http.calls, "/v1/events"), 0, "no sse path")
	assertEq(countPath(http.calls, "/v1/changes"), 0, "no poll")
	assertEq(http.streamAttempts[1], 0, "no stream yet")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#bootstrap
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#session-safety
add("PluginSessionCancelWritesNothing_WhenUnequal", function()
	local sess, http, studio = connect({ httpOpts = { equal = false } })
	sess:chooseBootstrap("cancel")
	assertEq(sess.live, false, "not live")
	assertEq(countPath(http.calls, "/v1/bootstrap/apply"), 0, "no apply")
	assertEq(#studio.log.applied, 0, "no studio writes")
	assertEq(studio.log.listeners, 0, "no listeners")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
-- Spec-Anchor: proj-spec/chapters/11-examples.md#example-first-connect-trees-differ
add("PluginSessionDiskWinsOmitsNodes_WhenApplyingBootstrap", function()
	local diskNodes = {
		{ nodeId = "server/Foo", className = "Script", source = "print(disk)\n", revision = 1 },
	}
	local sess, http, studio = connect({
		httpOpts = {
			equal = false,
			applyBootstrap = function(body)
				assertEq(body.direction, "disk-to-studio", "direction")
				assertEq(body.nodes, nil, "nodes omitted")
				return { nodes = diskNodes }
			end,
		},
	})
	sess:chooseBootstrap("disk")
	assertEq(sess.live, true, "live after disk wins")
	assertEq(#studio.log.applied, 1, "applied returned nodes")
	assertEq(sess.revisions["server/Foo"], 1, "revision")
	local applyCall = lastCall(http.calls, "/v1/bootstrap/apply")
	assertTrue(applyCall ~= nil, "apply called")
	assertEq(applyCall.body.direction, "disk-to-studio", "direction")
	assertEq(applyCall.body.nodes, nil, "no nodes field")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/07.1-http-api.md#post-v1bootstrapapply
add("PluginSessionStudioWinsResendsCompareSnapshot_WhenApplyingBootstrap", function()
	local studioNodes = {
		{ nodeId = "server/Foo", className = "Script", source = "print(1)\n" },
	}
	local compareBody
	local http = newHttp({
		equal = false,
		applyBootstrap = function(body)
			assertEq(body.direction, "studio-to-disk", "direction")
			assertTrue(body.nodes ~= nil, "nodes required")
			assertEq(body.nodes[1].nodeId, studioNodes[1].nodeId, "same nodeId")
			assertEq(body.nodes[1].source, studioNodes[1].source, "same source")
			return { applied = { { nodeId = "server/Foo", revision = 1 } } }
		end,
	})
	local origRequest = http.request
	http.request = function(req)
		if req.path == "/v1/bootstrap/compare" then
			compareBody = req.body
		end
		return origRequest(req)
	end
	local studio = newStudio({ mapped = studioNodes })
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = studio,
		clock = newClock(),
	})
	sess:chooseBootstrap("studio")
	assertEq(sess.live, true, "live")
	assertEq(#studio.log.applied, 0, "no DataModel mutate")
	assertEq(sess.revisions["server/Foo"], 1, "stored applied revision")
	assertTrue(compareBody ~= nil, "compare sent")
	local applyCall = lastCall(http.calls, "/v1/bootstrap/apply")
	assertEq(applyCall.body.nodes[1].nodeId, compareBody.nodes[1].nodeId, "same snapshot nodeId")
	assertEq(applyCall.body.nodes[1].source, compareBody.nodes[1].source, "same snapshot source")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d5--loopback-only-unprivileged-port-29850
add("PluginSessionRejectsHost_WhenAddressIsNotLoopback", function()
	local http = newHttp({})
	local sess = session().connect({
		host = "192.168.1.10",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = newStudio({}),
		clock = newClock(),
	})
	assertEq(sess.live, false, "not live")
	assertEq(#http.calls, 0, "no HTTP")
	assertTrue(sess.status ~= "rbx-sync not reachable" or #http.calls == 0, "rejected before health")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
add("PluginSessionAcceptsHost_WhenAddressIsIPv6Loopback", function()
	local sess, http = connect({ host = "::1", httpOpts = { equal = true } })
	assertEq(pathOf(http.calls[1]), "/v1/health", "health")
	assertTrue(sess.live == true or sess.awaitingBootstrap == true or sess.status ~= nil, "proceeded")
	assertEq(sess.live, true, "live on loopback")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution
add("PluginSessionSkipsUnresolvedRoot_WhenStudioToDiskSnapshot", function()
	local http = newHttp({
		equal = false,
		sessionOkBody = {
			protocolVersion = 1,
			agentVersion = "0.0.0",
			projectName = "example",
			roots = {
				{ id = "server", studio = "ServerScriptService" },
				{ id = "shared", studio = "ReplicatedStorage" },
			},
		},
	})
	local studio = newStudio({
		mapped = {
			{ nodeId = "server/Foo", className = "Script", source = "print(1)\n" },
			{ nodeId = "shared/Config", className = "ModuleScript", source = "return {}\n" },
		},
		unresolved = { ReplicatedStorage = true },
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = studio,
		clock = newClock(),
	})
	assertTrue(#sess.warnings > 0, "plugin warning")
	local compare = lastCall(http.calls, "/v1/bootstrap/compare")
	assertTrue(compare ~= nil, "compare")
	for _, n in compare.body.nodes do
		assertTrue(string.sub(n.nodeId, 1, 7) ~= "shared/", "skipped unresolved shared root")
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
-- Spec-Anchor: proj-spec/chapters/11-examples.md#example-conflict
add("PluginSessionKeepStudioForcesApply_WhenConflict", function()
	local applyCount = 0
	local http = newHttp({
		equal = true,
		applyHandler = function(body)
			applyCount += 1
			if applyCount == 1 then
				return {
					status = 409,
					body = {
						error = { code = "CONFLICT", message = "stale" },
						applied = {},
						conflicts = { { nodeId = "server/Foo", currentRevision = 4 } },
					},
				}
			end
			assertEq(body.changes[1].force, true, "force")
			assertEq(body.changes[1].baseRevision, 4, "current revision")
			return { status = 200, body = { applied = { { nodeId = "server/Foo", revision = 5 } } } }
		end,
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = newStudio({}),
		clock = clock,
	})
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	clock.advance(50)
	sess:tick()
	sess:resolveConflict("server/Foo", "keep-studio")
	assertEq(applyCount, 2, "second apply with force")
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
add("PluginSessionKeepDiskMatchesSnapshot_WhenConflict", function()
	local applyCount = 0
	local snapshotNode = {
		nodeId = "server/Foo",
		className = "Script",
		source = "print(disk)\n",
		revision = 4,
	}
	local http = newHttp({
		equal = true,
		snapshotBody = { nodes = { snapshotNode } },
		applyHandler = function(_body)
			applyCount += 1
			return {
				status = 409,
				body = {
					error = { code = "CONFLICT", message = "stale" },
					applied = {},
					conflicts = { { nodeId = "server/Foo", currentRevision = 4 } },
				},
			}
		end,
	})
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(studio)\n" } },
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = studio,
		clock = clock,
	})
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	clock.advance(50)
	sess:tick()
	sess:resolveConflict("server/Foo", "keep-disk")
	local mapped = studio.getMapped("server/Foo")
	assertTrue(mapped ~= nil, "mapped exists")
	assertEq(mapped.source, snapshotNode.source, "matches snapshot source")
	assertEq(mapped.className, snapshotNode.className, "matches snapshot class")
	assertEq(sess.revisions["server/Foo"], 4, "revision from snapshot")
	assertEq(applyCount, 1, "no extra apply route")
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
add("PluginSessionSkipLeavesBothUnchanged_WhenConflict", function()
	local applyCount = 0
	local http = newHttp({
		equal = true,
		applyHandler = function(_body)
			applyCount += 1
			return {
				status = 409,
				body = {
					error = { code = "CONFLICT", message = "stale" },
					applied = {},
					conflicts = { { nodeId = "server/Foo", currentRevision = 4 } },
				},
			}
		end,
	})
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(studio)\n" } },
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = studio,
		clock = clock,
	})
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	clock.advance(50)
	sess:tick()
	local before = studio.getMapped("server/Foo").source
	sess:resolveConflict("server/Foo", "skip")
	assertEq(studio.getMapped("server/Foo").source, before, "studio unchanged")
	assertEq(applyCount, 1, "no retry apply")
end)

-- Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1changes-fallback
-- Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#error-handling
-- Spec-Anchor: proj-spec/chapters/13-performance.md#limits
add("PluginSessionPollsChanges_WhenStreamFails", function()
	local clock = newClock()
	local sess, http = connect({
		clock = clock,
		httpOpts = { equal = true, streamShouldFail = true },
	})
	assertEq(sess.live, true, "live")
	clock.advance(200)
	sess:tick()
	assertTrue(countPath(http.calls, "/v1/changes") >= 1, "poll fallback")
	local n = countPath(http.calls, "/v1/changes")
	clock.advance(199)
	sess:tick()
	assertEq(countPath(http.calls, "/v1/changes"), n, "no poll before 200ms")
	clock.advance(1)
	sess:tick()
	assertEq(countPath(http.calls, "/v1/changes"), n + 1, "poll interval 200ms")
end)

-- Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#error-handling
add("PluginSessionRetriesStream_WhenTwoSecondsElapsed", function()
	local clock = newClock()
	local sess, http = connect({
		clock = clock,
		httpOpts = { equal = true, streamShouldFail = true },
	})
	local first = http.streamAttempts[1]
	clock.advance(1999)
	sess:tick()
	assertEq(http.streamAttempts[1], first, "no retry yet")
	clock.advance(1)
	sess:tick()
	assertTrue(http.streamAttempts[1] > first, "retry stream at 2s")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#error-handling
-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#session-lifecycle
add("PluginSessionStopsApplying_WhenSessionReplaced", function()
	local applyCount = 0
	local http = newHttp({
		equal = true,
		applyHandler = function(_body)
			applyCount += 1
			return { status = 200, body = { applied = { { nodeId = "server/Foo", revision = 2 } } } }
		end,
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = newStudio({}),
		clock = clock,
	})
	sess:onStreamEnd("replaced")
	assertEq(sess.status, "replaced by another session", "status")
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(x)\n",
	})
	clock.advance(50)
	sess:tick()
	assertEq(applyCount, 0, "stopped applying")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
add("PluginSessionRefusesApply_WhenDuplicateLuaLuau", function()
	local applyCount = 0
	local http = newHttp({
		equal = true,
		applyHandler = function(_body)
			applyCount += 1
			return { status = 200, body = { applied = {} } }
		end,
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = newStudio({ duplicateLuaLuau = true }),
		clock = clock,
	})
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "ModuleScript",
		source = "return 1\n",
	})
	clock.advance(50)
	sess:tick()
	assertEq(applyCount, 0, "refused apply")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d7--prefer-scripteditorservice-for-open-tabs
add("PluginSessionReadsEditorSource_WhenTabOpen", function()
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(open)\n" } },
		openTabs = { ["server/Foo"] = true },
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true }),
		studio = studio,
		clock = newClock(),
	})
	local src = sess:readSource("server/Foo")
	assertEq(src, "print(open)\n", "source")
	assertEq(studio.log.reads[1].via, "editor", "ScriptEditorService")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d7--prefer-scripteditorservice-for-open-tabs
add("PluginSessionReadsInstanceSource_WhenTabClosed", function()
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(closed)\n" } },
		openTabs = {},
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true }),
		studio = studio,
		clock = newClock(),
	})
	local src = sess:readSource("server/Foo")
	assertEq(src, "print(closed)\n", "source")
	assertEq(studio.log.reads[1].via, "source", "LuaSourceContainer.Source")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
add("PluginSessionWritesEditorSource_WhenTabOpen", function()
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(1)\n" } },
		openTabs = { ["server/Foo"] = true },
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true }),
		studio = studio,
		clock = newClock(),
	})
	sess:writeSource("server/Foo", "print(2)\n")
	assertEq(studio.log.writes[1].via, "editor", "UpdateSourceAsync")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
add("PluginSessionWritesInstanceSource_WhenTabClosed", function()
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(1)\n" } },
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true }),
		studio = studio,
		clock = newClock(),
	})
	sess:writeSource("server/Foo", "print(2)\n")
	assertEq(studio.log.writes[1].via, "source", "assign Source")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#roblox-studio-change-detection
-- Spec-Anchor: proj-spec/chapters/13-performance.md#limits
add("PluginSessionCoalescesMutations_WhenWithin50ms", function()
	local applyCount = 0
	local lastBody
	local http = newHttp({
		equal = true,
		applyHandler = function(body)
			applyCount += 1
			lastBody = body
			return { status = 200, body = { applied = { { nodeId = "server/Foo", revision = 2 } } } }
		end,
	})
	local clock = newClock()
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = newStudio({}),
		clock = clock,
	})
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(a)\n",
	})
	clock.advance(40)
	sess:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(b)\n",
	})
	clock.advance(10)
	sess:tick()
	assertEq(applyCount, 1, "one batched apply")
	assertEq(lastBody.changes[1].source, "print(b)\n", "latest source")
	assertEq(lastBody.changes[1].baseRevision, 1, "baseRevision")
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#echo-suppression
add("PluginSessionIgnoresStudioSignal_WhenApplyingDiskEvent", function()
	local applyCount = 0
	local http = newHttp({
		equal = true,
		applyHandler = function(_body)
			applyCount += 1
			return { status = 200, body = { applied = { { nodeId = "server/Foo", revision = 3 } } } }
		end,
	})
	local clock = newClock()
	local studio = newStudio({})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = http,
		studio = studio,
		clock = clock,
	})
	sess:applyDiskEvent({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(disk)\n",
		revision = 2,
		origin = "disk",
	}, function()
		sess:noteStudioMutation({
			op = "upsert",
			nodeId = "server/Foo",
			className = "Script",
			source = "print(disk)\n",
		})
	end)
	clock.advance(50)
	sess:tick()
	assertEq(applyCount, 0, "echo suppressed")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginSessionCreatesFolderAncestors_WhenUpsertNewNodeId", function()
	local studio = newStudio({ mapped = {}, instances = {} })
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true, snapshotBody = { nodes = {} } }),
		studio = studio,
		clock = newClock(),
	})
	sess:applyDiskEvent({
		op = "upsert",
		nodeId = "server/Bar/Child",
		className = "ModuleScript",
		source = "return 1\n",
		revision = 1,
		origin = "disk",
	})
	local created = studio.log.created
	assertTrue(#created >= 2, "ancestor + leaf")
	local classes = {}
	for _, c in created do
		classes[c.nodeId] = c.className
	end
	assertEq(classes["server/Bar"], "Folder", "missing intermediate Folder")
	assertEq(classes["server/Bar/Child"], "ModuleScript", "leaf class")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginSessionDoesNotDestroyUnmapped_WhenDeleteMappedSibling", function()
	local studio = newStudio({
		mapped = { { nodeId = "server/Foo", className = "Script", source = "print(1)\n" } },
		instances = {
			["server/Foo"] = { className = "Script", mapped = true },
			["server/Part"] = { className = "Part", mapped = false },
		},
		unmapped = { { name = "Part", className = "Part" } },
	})
	local sess = session().connect({
		host = "127.0.0.1",
		port = 29850,
		token = "tok",
		pluginVersion = "0.0.0",
		http = newHttp({ equal = true }),
		studio = studio,
		clock = newClock(),
	})
	sess:applyDiskEvent({
		op = "delete",
		nodeId = "server/Foo",
		revision = 2,
		origin = "disk",
	})
	local destroyedPart = false
	for _, id in studio.log.destroyed do
		if id == "server/Part" then
			destroyedPart = true
		end
	end
	assertEq(destroyedPart, false, "unmapped Part remains")
	local fooGone = false
	for _, id in studio.log.destroyed do
		if id == "server/Foo" then
			fooGone = true
		end
	end
	assertEq(fooGone, true, "mapped Foo destroyed")
end)

return tests
