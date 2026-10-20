--> L4 plugin Studio host tests (Lune). Fake Plugin / HttpService; no Roblox Studio.

local serde = require("@lune/serde")
local FakeDataModel = require("./fake_datamodel")

local hostModule
local hostLoadErr
do
	local ok, result = pcall(function()
		return require("../src/host")
	end)
	if ok then
		hostModule = result
	else
		hostLoadErr = result
	end
end

local function host()
	if hostLoadErr ~= nil then
		error("studio-plugin/src/host.luau: " .. tostring(hostLoadErr))
	end
	return hostModule
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

local function contains(hay: string, needle: string): boolean
	return string.find(hay, needle, 1, true) ~= nil
end

local function joinedTexts(app: any): string
	local parts = app.visibleTexts
	if type(parts) ~= "table" then
		return tostring(parts or "")
	end
	return table.concat(parts, "\n")
end

local function newPlugin()
	local log = {
		toolbars = {},
		buttons = {},
		widgets = {},
	}
	local pluginApi = {}
	function pluginApi:CreateToolbar(name: string)
		table.insert(log.toolbars, name)
		local toolbar = {}
		function toolbar:CreateButton(buttonId: string, tooltip: string, iconname: string, text: string)
			local button = {
				id = buttonId,
				tooltip = tooltip,
				iconname = iconname,
				text = text,
				ClickableWhenViewportHidden = false,
				onClick = nil :: (() -> ())?,
			}
			button.Click = {
				Connect = function(_, fn: () -> ())
					button.onClick = fn
					return {}
				end,
			}
			table.insert(log.buttons, button)
			return button
		end
		return toolbar
	end
	local function createWidget(pluginGuiId: string, info: any)
		local widget = {
			id = pluginGuiId,
			info = info,
			Title = "",
			Enabled = false,
		}
		table.insert(log.widgets, widget)
		return widget
	end
	function pluginApi:CreateDockWidgetPluginGui(pluginGuiId: string, info: any)
		return createWidget(pluginGuiId, info)
	end
	function pluginApi:CreateDockWidgetPluginGuiAsync(pluginGuiId: string, info: any)
		return createWidget(pluginGuiId, info)
	end
	pluginApi.log = log
	return pluginApi
end

local function newStudio()
	local mapped = { { nodeId = "server/Foo", className = "Script", source = "print(1)\n" } }
	return {
		resolveRoot = function(_path: string): boolean
			return true
		end,
		mappedNodes = function()
			return mapped
		end,
		mappedNodesForRoot = function(_rootId: string)
			return mapped
		end,
		startListeners = function() end,
		isEditorOpen = function(_nodeId: string): boolean
			return false
		end,
		getEditorSource = function(_nodeId: string): string
			return ""
		end,
		getInstanceSource = function(_nodeId: string): string
			for _, n in mapped do
				if n.nodeId == _nodeId then
					return n.source or ""
				end
			end
			return ""
		end,
		updateEditorSource = function(nodeId: string, source: string)
			for _, n in mapped do
				if n.nodeId == nodeId then
					n.source = source
				end
			end
		end,
		setInstanceSource = function(nodeId: string, source: string)
			for _, n in mapped do
				if n.nodeId == nodeId then
					n.source = source
				end
			end
		end,
		applyNodes = function() end,
		createInstance = function() end,
		destroyInstance = function() end,
	}
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

local function newHttpService(opts: any)
	opts = opts or {}
	local log = { requests = {}, streams = {}, pcallWrapped = false }
	local svc = { log = log }
	function svc:JSONEncode(value: any): string
		return serde.encode("json", value)
	end
	function svc:JSONDecode(text: string): any
		return serde.decode("json", text)
	end
	function svc:RequestAsync(requestOptions: any)
		table.insert(log.requests, requestOptions)
		if opts.throwOnRequest then
			error("HttpService failed")
		end
		local url = requestOptions.Url or ""
		if contains(url, "/v1/health") then
			if opts.healthFail then
				error("connection refused")
			end
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({ ok = true, protocolVersion = 1 }),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/session") then
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({
					protocolVersion = 1,
					agentVersion = "0.0.0",
					projectName = "example",
					roots = { { id = "server", studio = "ServerScriptService" } },
				}),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/bootstrap/compare") then
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({ equal = opts.equal ~= false }),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/snapshot") then
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({
					nodes = {
						{
							nodeId = "server/Foo",
							className = "Script",
							source = "print(disk)\n",
							revision = 1,
						},
					},
				}),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/bootstrap/apply") then
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({
					nodes = {
						{
							nodeId = "server/Foo",
							className = "Script",
							source = "print(disk)\n",
							revision = 1,
						},
					},
					applied = { { nodeId = "server/Foo", revision = 1 } },
				}),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/apply") then
			if opts.conflictOnApply and not log.applyForced then
				local decoded = {}
				if requestOptions.Body ~= nil and requestOptions.Body ~= "" then
					decoded = svc:JSONDecode(requestOptions.Body)
				end
				local force = decoded.changes and decoded.changes[1] and decoded.changes[1].force
				if force == true then
					log.applyForced = true
					return {
						Success = true,
						StatusCode = 200,
						StatusMessage = "OK",
						Body = svc:JSONEncode({ applied = { { nodeId = "server/Foo", revision = 5 } } }),
						Headers = { ["Content-Type"] = "application/json" },
					}
				end
				return {
					Success = false,
					StatusCode = 409,
					StatusMessage = "Conflict",
					Body = svc:JSONEncode({
						error = { code = "CONFLICT", message = "stale" },
						applied = {},
						conflicts = { { nodeId = "server/Foo", currentRevision = 4 } },
					}),
					Headers = { ["Content-Type"] = "application/json" },
				}
			end
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = svc:JSONEncode({ applied = { { nodeId = "server/Foo", revision = 2 } } }),
				Headers = { ["Content-Type"] = "application/json" },
			}
		end
		if contains(url, "/v1/events") then
			return {
				Success = true,
				StatusCode = 200,
				StatusMessage = "OK",
				Body = "",
				Headers = {},
			}
		end
		return {
			Success = false,
			StatusCode = 404,
			StatusMessage = "Not Found",
			Body = svc:JSONEncode({ error = { code = "NOT_FOUND", message = "not found" } }),
			Headers = { ["Content-Type"] = "application/json" },
		}
	end
	function svc:CreateWebStreamClient(streamClientType: any, requestOptions: any)
		table.insert(log.streams, { streamClientType = streamClientType, requestOptions = requestOptions })
		return {
			Opened = {
				Connect = function()
					return {}
				end,
			},
			MessageReceived = {
				Connect = function()
					return {}
				end,
			},
			Closed = {
				Connect = function()
					return {}
				end,
			},
			Error = {
				Connect = function()
					return {}
				end,
			},
			close = function() end,
		}
	end
	return svc
end

local function startHost(overrides: any)
	overrides = overrides or {}
	local pluginApi = overrides.plugin or newPlugin()
	local httpService = overrides.httpService or newHttpService(overrides.httpOpts)
	local clock = overrides.clock or newClock()
	local studio = overrides.studio or newStudio()
	local app = host().start({
		plugin = pluginApi,
		httpService = httpService,
		studio = studio,
		clock = clock,
		pluginVersion = "0.0.0",
		heartbeatConnect = overrides.heartbeatConnect,
	})
	return app, pluginApi, httpService, clock
end

local tests = {}

local function add(name: string, fn: () -> ())
	table.insert(tests, { name = name, fn = fn })
end

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
add("PluginHostCreatesToolbarNamedRbxSync_WhenStarted", function()
	local _app, pluginApi = startHost()
	assertTrue(#pluginApi.log.toolbars >= 1, "CreateToolbar must run")
	assertEq(pluginApi.log.toolbars[1], "rbx-sync", "toolbar name")
	assertTrue(#pluginApi.log.buttons >= 1, "toolbar needs a button to appear on the Plugins ribbon")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
add("PluginHostCreatesDockableWidget_WhenStarted", function()
	local _app, pluginApi = startHost()
	assertTrue(#pluginApi.log.widgets >= 1, "dockable widget")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
add("PluginHostShowsConnectDisconnectAndChoices_WhenWidgetCreated", function()
	local app = startHost()
	local texts = joinedTexts(app)
	assertTrue(contains(texts, "Connect"), "Connect")
	assertTrue(contains(texts, "Disconnect"), "Disconnect")
	assertTrue(contains(texts, "Disk wins"), "Disk wins")
	assertTrue(contains(texts, "Roblox Studio wins"), "Roblox Studio wins")
	assertTrue(contains(texts, "Cancel"), "Cancel")
	assertTrue(contains(texts, "Keep disk"), "Keep disk")
	assertTrue(contains(texts, "Keep Roblox Studio"), "Keep Roblox Studio")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
add("PluginHostDefaultsLoopbackAndPort_WhenWidgetCreated", function()
	local app = startHost()
	assertEq(app.hostValue, "127.0.0.1", "default host")
	assertEq(tonumber(app.portValue), 29850, "default port")
	assertEq(app.tokenValue, "", "default token empty")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#http
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
add("PluginHostCallsRequestAsync_WhenConnectSubmitted", function()
	local app, _pluginApi, httpService = startHost()
	app.tokenValue = "tok"
	app:clickConnect()
	assertTrue(#httpService.log.requests >= 1, "RequestAsync")
	local health = httpService.log.requests[1]
	assertTrue(contains(health.Url or "", "/v1/health"), "health first")
	assertTrue(contains(health.Url or "", "127.0.0.1"), "loopback")
	assertTrue(contains(health.Url or "", "29850"), "default port")
	assertTrue(not contains(health.Url or "", "accessToken="), "health has no token")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#http
add("PluginHostPutsTokenInQuery_WhenAuthenticatedRequestAsync", function()
	local app, _pluginApi, httpService = startHost()
	app.tokenValue = "tok"
	app:clickConnect()
	local found = false
	for _, req in httpService.log.requests do
		local url = req.Url or ""
		if contains(url, "/v1/session") then
			found = true
			assertTrue(contains(url, "accessToken=tok"), "token query")
			local headers = req.Headers or {}
			for key in headers do
				local lower = string.lower(tostring(key))
				assertTrue(lower ~= "authorization" and not contains(lower, "token"), "no custom auth header")
			end
		end
	end
	assertTrue(found, "session RequestAsync")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#http
-- Spec-Anchor: proj-spec/chapters/07.2-event-stream.md#get-v1events-preferred
add("PluginHostCallsCreateWebStreamClient_WhenEnteringLive", function()
	local app, _pluginApi, httpService = startHost({ httpOpts = { equal = true } })
	app.tokenValue = "tok"
	app:clickConnect()
	assertTrue(#httpService.log.streams >= 1, "CreateWebStreamClient")
	local stream = httpService.log.streams[1]
	local url = stream.requestOptions and stream.requestOptions.Url or ""
	assertTrue(contains(url, "/v1/events"), "events url")
	assertTrue(contains(url, "accessToken=tok"), "token on stream")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#error-handling
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#http
add("PluginHostWrapsHttpServiceInPcall_WhenRequestAsyncThrows", function()
	local app, _pluginApi, httpService = startHost({ httpOpts = { throwOnRequest = true } })
	app.tokenValue = "tok"
	local ok = pcall(function()
		app:clickConnect()
	end)
	assertTrue(ok, "HttpService errors must be pcall-wrapped")
	assertTrue(#httpService.log.requests >= 1, "attempted RequestAsync")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#error-handling
add("PluginHostShowsNotReachable_WhenHealthFails", function()
	local app = startHost({ httpOpts = { healthFail = true } })
	app.tokenValue = "tok"
	app:clickConnect()
	local texts = joinedTexts(app)
	assertTrue(contains(texts, "rbx-sync not reachable") or contains(tostring(app.status or ""), "rbx-sync not reachable"), "not reachable")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#construction--configuration
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#connect-flow
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginHostAppliesDiskBootstrapToDataModel_WhenStudioNotInjected", function()
	local world = FakeDataModel.new()
	local pluginApi = newPlugin()
	local httpService = newHttpService({ equal = false })
	local app = host().start({
		plugin = pluginApi,
		httpService = httpService,
		clock = newClock(),
		dataModel = world.game,
		scriptEditorService = world.scriptEditor,
		newInstance = world.newInstance,
		pluginVersion = "0.0.0",
	})
	app.tokenValue = "tok"
	app:clickConnect()
	assertTrue(app.session ~= nil, "session started")
	app.session:chooseBootstrap("disk")
	local foo = world.sss:FindFirstChild("Foo")
	assertTrue(foo ~= nil, "disk-to-studio must create mapped Script on DataModel")
	assertEq(foo.ClassName, "Script", "class")
	assertEq(foo.Source, "print(disk)\n", "source")
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
add("PluginHostKeepDiskMatchesSnapshot_WhenClicked", function()
	local app, _pluginApi, httpService, clock = startHost({ httpOpts = { equal = true, conflictOnApply = true } })
	app.tokenValue = "tok"
	app:clickConnect()
	app.session:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	clock.advance(50)
	app.session:tick()
	app:clickKeepDisk()
	local mapped = app.session.studio.mappedNodes()[1]
	assertEq(mapped.source, "print(disk)\n", "Keep disk uses snapshot source")
	assertTrue(httpService.log.applyForced ~= true, "Keep disk must not force apply")
end)

-- Spec-Anchor: proj-spec/chapters/06.5-sync-session.md#conflict-ui
add("PluginHostKeepRobloxStudioForcesApply_WhenClicked", function()
	local app, _pluginApi, httpService, clock = startHost({ httpOpts = { equal = true, conflictOnApply = true } })
	app.tokenValue = "tok"
	app:clickConnect()
	app.session:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	clock.advance(50)
	app.session:tick()
	app:clickKeepRobloxStudio()
	assertTrue(httpService.log.applyForced == true, "Keep Roblox Studio posts force apply")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#roblox-studio-change-detection
add("PluginHostTicksSession_WhenHeartbeatFires", function()
	local heartbeats = {}
	local app, _pluginApi, httpService = startHost({
		httpOpts = { equal = true },
		heartbeatConnect = function(fn)
			table.insert(heartbeats, fn)
		end,
	})
	app.tokenValue = "tok"
	app:clickConnect()
	assertTrue(#heartbeats >= 1, "host must bind a heartbeat to flush the 50 ms batch")
	app.session:noteStudioMutation({
		op = "upsert",
		nodeId = "server/Foo",
		className = "Script",
		source = "print(studio)\n",
	})
	app.session.clock.advance(50)
	heartbeats[1]()
	local found = false
	for _, req in httpService.log.requests do
		if contains(req.Url or "", "/v1/apply") then
			found = true
		end
	end
	assertTrue(found, "heartbeat tick must POST /v1/apply")
end)

add("PluginHostRejectsHost_WhenAddressIsNotLoopback", function()
	local app, _pluginApi, httpService = startHost()
	app.hostValue = "10.0.0.1"
	app.portValue = 29850
	app.tokenValue = "tok"
	app:clickConnect()
	assertEq(#httpService.log.requests, 0, "LAN host must not call HttpService")
end)

return tests
