-- Plugin connect / session state machine (§6.3, §6.5). HTTP and DataModel injected.

local STATUS_NOT_REACHABLE = "rbx-sync not reachable"
local STATUS_REPLACED = "replaced by another session"
local STATUS_DISCONNECTED = "disconnected"
local STATUS_CONNECTED = "connected"
local STATUS_BOOTSTRAP = "choose Disk wins, Roblox Studio wins, or Cancel"

local function isLoopback(host: string): boolean
	return host == "127.0.0.1" or host == "::1"
end

local function splitNodeId(nodeId: string): { string }
	local parts = {}
	for part in string.gmatch(nodeId, "[^/]+") do
		table.insert(parts, part)
	end
	return parts
end

local function ancestorId(parts: { string }, through: number): string
	return table.concat(parts, "/", 1, through)
end

local function belongsToRoot(nodeId: string, rootId: string): boolean
	return nodeId == rootId or string.sub(nodeId, 1, #rootId + 1) == rootId .. "/"
end

local Session = {}
Session.__index = Session

function Session:_request(method: string, path: string, body: any?, authed: boolean): (any?, string?)
	local req = {
		method = method,
		path = path,
		body = body,
		headers = {},
		queryToken = if authed then self.token else nil,
	}
	local ok, result = pcall(function()
		return self.http.request(req)
	end)
	if not ok then
		return nil, "pcall"
	end
	return result, nil
end

function Session:_collectSnapshot(): { any }
	local nodes = {}
	for _, root in self.roots do
		if not self.studio.resolveRoot(root.studio) then
			table.insert(self.warnings, "unresolved studio path: " .. root.studio)
			continue
		end
		local fromRoot
		if self.studio.mappedNodesForRoot then
			fromRoot = self.studio.mappedNodesForRoot(root.id)
		else
			fromRoot = {}
			for _, n in self.studio.mappedNodes() do
				if belongsToRoot(n.nodeId, root.id) then
					table.insert(fromRoot, n)
				end
			end
		end
		for _, n in fromRoot do
			table.insert(nodes, n)
		end
	end
	return nodes
end

function Session:_enterLive()
	self.live = true
	self.awaitingBootstrap = false
	self.status = STATUS_CONNECTED
	self.studio.startListeners(function(change)
		self:noteStudioMutation(change)
	end)
	local streamOk = false
	if self.http.openStream then
		local ok = self.http.openStream("/v1/events")
		streamOk = ok == true
	end
	self.streamOpen = streamOk
	if not streamOk then
		self.polling = true
		self.nextStreamRetryMs = self.clock.nowMs() + 2000
	end
end

function Session:_storeRevisionsFromNodes(nodes: { any })
	for _, n in nodes do
		if n.revision ~= nil then
			self.revisions[n.nodeId] = n.revision
		end
	end
end

function Session:_storeRevisionsFromApplied(applied: { any }?)
	if applied == nil then
		return
	end
	for _, row in applied do
		if row.revision ~= nil then
			local id = row.nodeId
			self.revisions[id] = row.revision
		end
	end
end

function Session:_handshakeAndCompare(): ()
	local health, healthErr = self:_request("GET", "/v1/health", nil, false)
	if healthErr ~= nil or health == nil or health.status ~= 200 then
		self.status = STATUS_NOT_REACHABLE
		return
	end

	local sess, sessErr = self:_request("POST", "/v1/session", {
		protocolVersion = 1,
		pluginVersion = self.pluginVersion,
	}, true)
	if sessErr ~= nil then
		self.status = STATUS_DISCONNECTED
		return
	end
	if sess.status ~= 200 then
		local err = sess.body and sess.body.error
		if err and (err.code == "UNAUTHORIZED" or err.code == "PROTOCOL_MISMATCH") then
			self.status = err.message
		else
			self.status = STATUS_DISCONNECTED
		end
		return
	end
	self.roots = sess.body.roots or {}
	if self.studio.setRoots then
		self.studio.setRoots(self.roots)
	end

	local snapshot = self:_collectSnapshot()
	self.lastCompareNodes = snapshot
	local compare, compareErr = self:_request("POST", "/v1/bootstrap/compare", { nodes = snapshot }, true)
	if compareErr ~= nil or compare == nil then
		self.status = STATUS_DISCONNECTED
		return
	end
	if compare.body and compare.body.equal == true then
		local snap, snapErr = self:_request("GET", "/v1/snapshot", nil, true)
		if snapErr ~= nil or snap == nil or snap.status ~= 200 then
			self.status = STATUS_DISCONNECTED
			return
		end
		self:_storeRevisionsFromNodes(snap.body.nodes or {})
		self:_enterLive()
		return
	end
	self.awaitingBootstrap = true
	self.status = STATUS_BOOTSTRAP
end

function Session.connect(opts: any): any
	local self = setmetatable({
		host = opts.host,
		port = opts.port,
		token = opts.token,
		pluginVersion = opts.pluginVersion or "0.0.0",
		http = opts.http,
		studio = opts.studio,
		clock = opts.clock,
		status = "disconnected",
		live = false,
		awaitingBootstrap = false,
		revisions = {},
		warnings = {},
		roots = {},
		lastCompareNodes = {},
		pending = {},
		conflicts = {},
		pendingByNode = {},
		debounceDeadline = nil,
		polling = false,
		streamOpen = false,
		nextStreamRetryMs = nil,
		lastPollMs = nil,
		liveStartMs = opts.clock.nowMs(),
		stoppedApplying = false,
		ignoring = {},
	}, Session)

	if not isLoopback(opts.host) then
		self.status = "host must be loopback"
		return self
	end

	self:_handshakeAndCompare()
	return self
end

function Session:chooseBootstrap(choice: string)
	if not self.awaitingBootstrap then
		return
	end
	if choice == "cancel" then
		self.awaitingBootstrap = false
		self.live = false
		self.status = STATUS_DISCONNECTED
		return
	end
	if choice == "disk" then
		local result, err = self:_request("POST", "/v1/bootstrap/apply", { direction = "disk-to-studio" }, true)
		if err ~= nil or result == nil or result.status ~= 200 then
			self.status = STATUS_DISCONNECTED
			return
		end
		local nodes = result.body.nodes or {}
		self.studio.applyNodes(nodes)
		self:_storeRevisionsFromNodes(nodes)
		self:_enterLive()
		return
	end
	if choice == "studio" then
		local result, err = self:_request("POST", "/v1/bootstrap/apply", {
			direction = "studio-to-disk",
			nodes = self.lastCompareNodes,
		}, true)
		if err ~= nil or result == nil or result.status ~= 200 then
			self.status = STATUS_DISCONNECTED
			return
		end
		self:_storeRevisionsFromApplied(result.body.applied)
		self:_enterLive()
	end
end

function Session:readSource(nodeId: string): string
	if self.studio.isEditorOpen(nodeId) then
		return self.studio.getEditorSource(nodeId)
	end
	return self.studio.getInstanceSource(nodeId)
end

function Session:writeSource(nodeId: string, source: string)
	if self.studio.isEditorOpen(nodeId) then
		self.studio.updateEditorSource(nodeId, source)
		return
	end
	self.studio.setInstanceSource(nodeId, source)
end

function Session:noteStudioMutation(change: any)
	if not self.live or self.stoppedApplying then
		return
	end
	if self.ignoring[change.nodeId] then
		return
	end
	self.pending[change.nodeId] = change
	self.pendingByNode[change.nodeId] = change
	if self.debounceDeadline == nil then
		self.debounceDeadline = self.clock.nowMs() + 50
	end
end

function Session:_flushMutations()
	self.debounceDeadline = nil
	if self.stoppedApplying then
		self.pending = {}
		return
	end
	if self.studio.hasDuplicateLuaLuau and self.studio.hasDuplicateLuaLuau() then
		self.pending = {}
		return
	end
	local changes = {}
	for _, ch in self.pending do
		local entry = {
			op = ch.op,
			nodeId = ch.nodeId,
			className = ch.className,
			source = ch.source,
			baseRevision = self.revisions[ch.nodeId] or 0,
			force = false,
		}
		if ch.newNodeId ~= nil then
			entry.newNodeId = ch.newNodeId
		end
		table.insert(changes, entry)
	end
	if #changes == 0 then
		return
	end
	local result, err = self:_request("POST", "/v1/apply", { changes = changes }, true)
	self.pending = {}
	if err ~= nil or result == nil then
		self.status = STATUS_DISCONNECTED
		return
	end
	if result.status == 409 then
		for _, c in result.body.conflicts or {} do
			self.conflicts[c.nodeId] = c.currentRevision
		end
		return
	end
	if result.status == 200 then
		self:_storeRevisionsFromApplied(result.body.applied)
	end
end

function Session:tick()
	local now = self.clock.nowMs()
	if self.debounceDeadline ~= nil and now >= self.debounceDeadline then
		self:_flushMutations()
	end
	if not self.live or self.stoppedApplying then
		return
	end
	if self.polling then
		if self.lastPollMs == nil then
			if now - self.liveStartMs >= 200 then
				self:_request("GET", "/v1/changes", nil, true)
				self.lastPollMs = now
			end
		elseif now - self.lastPollMs >= 200 then
			self:_request("GET", "/v1/changes", nil, true)
			self.lastPollMs = now
		end
		if self.nextStreamRetryMs ~= nil and now >= self.nextStreamRetryMs then
			local ok = self.http.openStream("/v1/events")
			if ok == true then
				self.streamOpen = true
				self.polling = false
			else
				self.nextStreamRetryMs = now + 2000
			end
		end
	end
end

function Session:resolveConflict(nodeId: string, choice: string)
	local current = self.conflicts[nodeId]
	if current == nil then
		return
	end
	if choice == "skip" then
		self.conflicts[nodeId] = nil
		return
	end
	if choice == "keep-studio" then
		local ch = self.pendingByNode[nodeId]
		if ch == nil then
			return
		end
		local result, err = self:_request("POST", "/v1/apply", {
			changes = {
				{
					op = ch.op,
					nodeId = ch.nodeId,
					className = ch.className,
					source = ch.source,
					baseRevision = current,
					force = true,
				},
			},
		}, true)
		if err == nil and result and result.status == 200 then
			self:_storeRevisionsFromApplied(result.body.applied)
			self.conflicts[nodeId] = nil
		end
		return
	end
	if choice == "keep-disk" then
		local snap, err = self:_request("GET", "/v1/snapshot", nil, true)
		if err ~= nil or snap == nil or snap.status ~= 200 then
			self.status = STATUS_DISCONNECTED
			return
		end
		local found = false
		for _, n in snap.body.nodes or {} do
			if n.nodeId == nodeId then
				found = true
				for _, m in self.studio.mappedNodes() do
					if m.nodeId == nodeId then
						m.source = n.source
						m.className = n.className
						m.revision = n.revision
					end
				end
				if n.className ~= nil and n.className ~= "Folder" then
					self:writeSource(nodeId, n.source or "")
				end
				if n.revision ~= nil then
					self.revisions[nodeId] = n.revision
				end
			end
		end
		if not found and self.studio.destroyInstance then
			self.studio.destroyInstance(nodeId)
		end
		self.conflicts[nodeId] = nil
	end
end

function Session:onStreamEnd(reason: string)
	if reason == "replaced" then
		self.status = STATUS_REPLACED
		self.stoppedApplying = true
		self.live = false
	end
end

function Session:_createAncestors(nodeId: string)
	local parts = splitNodeId(nodeId)
	if #parts < 2 then
		return
	end
	for i = 2, #parts - 1 do
		local id = ancestorId(parts, i)
		self.studio.createInstance(id, "Folder")
	end
end

function Session:applyDiskEvent(event: any, during: (() -> ())?)
	if self.stoppedApplying then
		return
	end
	self.ignoring[event.nodeId] = true
	if event.op == "upsert" then
		self:_createAncestors(event.nodeId)
		if self.studio.createInstance then
			self.studio.createInstance(event.nodeId, event.className)
		end
		self:writeSource(event.nodeId, event.source or "")
	elseif event.op == "delete" then
		local inst = self.studio.instances and self.studio.instances[event.nodeId]
		if inst == nil or inst.mapped ~= false then
			self.studio.destroyInstance(event.nodeId)
		end
	end
	if during then
		during()
	end
	self.ignoring[event.nodeId] = nil
	if event.revision ~= nil then
		self.revisions[event.nodeId] = event.revision
	end
end

return {
	connect = Session.connect,
}
