-- DataModel + ScriptEditorService adapter (§6.3, §6.4).

local SOURCE_CLASSES = {
	Script = true,
	LocalScript = true,
	ModuleScript = true,
}

local MAPPED_CLASSES = {
	Script = true,
	LocalScript = true,
	ModuleScript = true,
	Folder = true,
}

local function splitDot(path: string): { string }
	local parts = {}
	for part in string.gmatch(path, "[^.]+") do
		table.insert(parts, part)
	end
	return parts
end

local function splitSlash(path: string): { string }
	local parts = {}
	for part in string.gmatch(path, "[^/]+") do
		table.insert(parts, part)
	end
	return parts
end

local function resolveStudioPath(dataModel: any, studioPath: string): any?
	if dataModel == nil or studioPath == nil or studioPath == "" then
		return nil
	end
	local parts = splitDot(studioPath)
	if #parts == 0 then
		return nil
	end
	local inst = nil
	local ok, svc = pcall(function()
		return dataModel:GetService(parts[1])
	end)
	if ok and svc ~= nil then
		inst = svc
	else
		inst = dataModel:FindFirstChild(parts[1])
	end
	if inst == nil then
		return nil
	end
	for i = 2, #parts do
		inst = inst:FindFirstChild(parts[i])
		if inst == nil then
			return nil
		end
	end
	return inst
end

local function nodeIdFor(rootId: string, rootInst: any, inst: any): string?
	if inst == rootInst then
		return nil
	end
	local names = {}
	local cur = inst
	while cur ~= nil and cur ~= rootInst do
		table.insert(names, 1, cur.Name)
		cur = cur.Parent
	end
	if cur ~= rootInst then
		return nil
	end
	return rootId .. "/" .. table.concat(names, "/")
end

local function defaultNewInstance(className: string): any
	return Instance.new(className)
end

local Studio = {}

function Studio.new(deps: any)
	deps = deps or {}
	local dataModel = deps.dataModel
	local scriptEditor = deps.scriptEditorService
	local newInstance = deps.newInstance or defaultNewInstance
	local roots = {}
	local instanceNodeId: { [any]: string } = {}

	local function rootById(rootId: string): any?
		for _, root in roots do
			if root.id == rootId then
				return root
			end
		end
		return nil
	end

	local function rootInstFor(root: any): any?
		return resolveStudioPath(dataModel, root.studio)
	end

	local function findInstance(nodeId: string): any?
		local parts = splitSlash(nodeId)
		if #parts < 2 then
			return nil
		end
		local root = rootById(parts[1])
		if root == nil then
			return nil
		end
		local inst = rootInstFor(root)
		if inst == nil then
			return nil
		end
		for i = 2, #parts do
			inst = inst:FindFirstChild(parts[i])
			if inst == nil then
				return nil
			end
		end
		return inst
	end

	local function isEditorOpenInst(inst: any): boolean
		if scriptEditor == nil or inst == nil then
			return false
		end
		local ok, doc = pcall(function()
			return scriptEditor:FindScriptDocument(inst)
		end)
		return ok and doc ~= nil
	end

	local function readSourceInst(inst: any): string
		if inst == nil or not SOURCE_CLASSES[inst.ClassName] then
			return ""
		end
		if isEditorOpenInst(inst) then
			local ok, src = pcall(function()
				return scriptEditor:GetEditorSource(inst)
			end)
			if ok and type(src) == "string" then
				return src
			end
		end
		return inst.Source or ""
	end

	local function writeSourceInst(inst: any, source: string)
		if inst == nil or not SOURCE_CLASSES[inst.ClassName] then
			return
		end
		if isEditorOpenInst(inst) then
			pcall(function()
				scriptEditor:UpdateSourceAsync(inst, function(_old: string)
					return source
				end)
			end)
			return
		end
		inst.Source = source
	end

	local function ensureChild(parent: any, name: string, className: string): any
		local existing = parent:FindFirstChild(name)
		if existing ~= nil then
			if existing.ClassName == className then
				return existing
			end
			if MAPPED_CLASSES[existing.ClassName] then
				existing:Destroy()
			else
				return existing
			end
		end
		local inst = newInstance(className)
		inst.Name = name
		inst.Parent = parent
		return inst
	end

	local function ensureStudioPath(studioPath: string): any?
		if dataModel == nil or studioPath == nil or studioPath == "" then
			return nil
		end
		local parts = splitDot(studioPath)
		if #parts == 0 then
			return nil
		end
		local inst = nil
		local ok, svc = pcall(function()
			return dataModel:GetService(parts[1])
		end)
		if ok and svc ~= nil then
			inst = svc
		else
			inst = dataModel:FindFirstChild(parts[1])
		end
		if inst == nil then
			return nil
		end
		for i = 2, #parts do
			inst = ensureChild(inst, parts[i], "Folder")
		end
		return inst
	end

	local function ensureNode(nodeId: string, className: string): any?
		local parts = splitSlash(nodeId)
		if #parts < 2 then
			return nil
		end
		local root = rootById(parts[1])
		if root == nil then
			return nil
		end
		local inst = ensureStudioPath(root.studio)
		if inst == nil then
			return nil
		end
		for i = 2, #parts do
			local classFor = if i == #parts then className else "Folder"
			inst = ensureChild(inst, parts[i], classFor)
		end
		return inst
	end

	local adapter = {}

	function adapter.resolveRoot(studioPath: string): boolean
		return resolveStudioPath(dataModel, studioPath) ~= nil
	end

	function adapter.setRoots(nextRoots: { any })
		roots = nextRoots or {}
	end

	function adapter.mappedNodesForRoot(rootId: string): { any }
		local nodes = {}
		local root = rootById(rootId)
		if root == nil then
			return nodes
		end
		local rootInst = rootInstFor(root)
		if rootInst == nil then
			return nodes
		end
		for _, inst in rootInst:GetDescendants() do
			if MAPPED_CLASSES[inst.ClassName] then
				local id = nodeIdFor(rootId, rootInst, inst)
				if id ~= nil then
					local node = {
						nodeId = id,
						className = inst.ClassName,
					}
					if SOURCE_CLASSES[inst.ClassName] then
						node.source = readSourceInst(inst)
					end
					table.insert(nodes, node)
				end
			end
		end
		return nodes
	end

	function adapter.mappedNodes(): { any }
		local nodes = {}
		for _, root in roots do
			for _, n in adapter.mappedNodesForRoot(root.id) do
				table.insert(nodes, n)
			end
		end
		return nodes
	end

	function adapter.isEditorOpen(nodeId: string): boolean
		return isEditorOpenInst(findInstance(nodeId))
	end

	function adapter.getEditorSource(nodeId: string): string
		return readSourceInst(findInstance(nodeId))
	end

	function adapter.getInstanceSource(nodeId: string): string
		local inst = findInstance(nodeId)
		if inst == nil then
			return ""
		end
		return inst.Source or ""
	end

	function adapter.updateEditorSource(nodeId: string, source: string)
		writeSourceInst(findInstance(nodeId), source)
	end

	function adapter.setInstanceSource(nodeId: string, source: string)
		local inst = findInstance(nodeId)
		if inst ~= nil then
			inst.Source = source
		end
	end

	function adapter.createInstance(nodeId: string, className: string)
		ensureNode(nodeId, className)
	end

	function adapter.destroyInstance(nodeId: string)
		local inst = findInstance(nodeId)
		if inst ~= nil and MAPPED_CLASSES[inst.ClassName] then
			inst:Destroy()
		end
	end

	function adapter.applyNodes(nodes: { any })
		local keep: { [string]: boolean } = {}
		for _, n in nodes or {} do
			keep[n.nodeId] = true
			local inst = ensureNode(n.nodeId, n.className)
			if inst ~= nil and SOURCE_CLASSES[n.className] then
				writeSourceInst(inst, n.source or "")
			end
		end
		for _, root in roots do
			local rootInst = rootInstFor(root)
			if rootInst == nil then
				continue
			end
			local mapped = {}
			for _, inst in rootInst:GetDescendants() do
				if MAPPED_CLASSES[inst.ClassName] then
					table.insert(mapped, inst)
				end
			end
			for i = #mapped, 1, -1 do
				local inst = mapped[i]
				local id = nodeIdFor(root.id, rootInst, inst)
				if id ~= nil and not keep[id] then
					inst:Destroy()
				end
			end
		end
	end

	function adapter.startListeners(onMutation: (any) -> ())
		local function emit(change: any)
			if onMutation then
				onMutation(change)
			end
		end

		local function bindInst(inst: any, nodeId: string)
			instanceNodeId[inst] = nodeId
			pcall(function()
				inst:GetPropertyChangedSignal("Name"):Connect(function()
					local oldId = instanceNodeId[inst]
					local newId = nil
					for _, root in roots do
						local rootInst = rootInstFor(root)
						if rootInst ~= nil then
							newId = nodeIdFor(root.id, rootInst, inst)
							if newId ~= nil then
								break
							end
						end
					end
					if newId ~= nil then
						instanceNodeId[inst] = newId
					end
					emit({
						op = "rename",
						nodeId = oldId,
						newNodeId = newId,
						className = inst.ClassName,
						source = readSourceInst(inst),
					})
				end)
			end)
			if SOURCE_CLASSES[inst.ClassName] then
				pcall(function()
					inst:GetPropertyChangedSignal("Source"):Connect(function()
						local id = instanceNodeId[inst]
						if id == nil then
							return
						end
						emit({
							op = "upsert",
							nodeId = id,
							className = inst.ClassName,
							source = readSourceInst(inst),
						})
					end)
				end)
			end
		end

		local function onAdded(inst: any)
			if not MAPPED_CLASSES[inst.ClassName] then
				return
			end
			for _, root in roots do
				local rootInst = rootInstFor(root)
				if rootInst == nil then
					continue
				end
				local id = nodeIdFor(root.id, rootInst, inst)
				if id ~= nil then
					bindInst(inst, id)
					emit({
						op = "upsert",
						nodeId = id,
						className = inst.ClassName,
						source = readSourceInst(inst),
					})
					return
				end
			end
		end

		local function onRemoving(inst: any)
			if not MAPPED_CLASSES[inst.ClassName] then
				return
			end
			local id = instanceNodeId[inst]
			if id == nil then
				for _, root in roots do
					local rootInst = rootInstFor(root)
					if rootInst ~= nil then
						id = nodeIdFor(root.id, rootInst, inst)
						if id ~= nil then
							break
						end
					end
				end
			end
			instanceNodeId[inst] = nil
			if id ~= nil then
				emit({
					op = "delete",
					nodeId = id,
					className = inst.ClassName,
				})
			end
		end

		for _, root in roots do
			local rootInst = rootInstFor(root)
			if rootInst == nil then
				continue
			end
			for _, inst in rootInst:GetDescendants() do
				if MAPPED_CLASSES[inst.ClassName] then
					local id = nodeIdFor(root.id, rootInst, inst)
					if id ~= nil then
						bindInst(inst, id)
					end
				end
			end
			pcall(function()
				rootInst.DescendantAdded:Connect(onAdded)
			end)
			pcall(function()
				rootInst.DescendantRemoving:Connect(onRemoving)
			end)
		end

		if scriptEditor ~= nil and scriptEditor.TextDocumentDidChange then
			pcall(function()
				scriptEditor.TextDocumentDidChange:Connect(function(document: any, _changes: any)
					local ok, inst = pcall(function()
						return document:GetScript()
					end)
					if not ok or inst == nil then
						return
					end
					local id = instanceNodeId[inst]
					if id == nil then
						for _, root in roots do
							local rootInst = rootInstFor(root)
							if rootInst ~= nil then
								id = nodeIdFor(root.id, rootInst, inst)
								if id ~= nil then
									break
								end
							end
						end
					end
					if id == nil then
						return
					end
					emit({
						op = "upsert",
						nodeId = id,
						className = inst.ClassName,
						source = readSourceInst(inst),
					})
				end)
			end)
		end
	end

	return adapter
end

return Studio
