-- Pure Luau project mapping (§6.4). No HttpService / Lune APIs.

local initFiles = {
	["init.luau"] = "ModuleScript",
	["init.lua"] = "ModuleScript",
	["init.server.luau"] = "Script",
	["init.server.lua"] = "Script",
	["init.client.luau"] = "LocalScript",
	["init.client.lua"] = "LocalScript",
}

local leafSuffixes = {
	{ suffix = ".server.luau", className = "Script" },
	{ suffix = ".server.lua", className = "Script" },
	{ suffix = ".client.luau", className = "LocalScript" },
	{ suffix = ".client.lua", className = "LocalScript" },
	{ suffix = ".luau", className = "ModuleScript" },
	{ suffix = ".lua", className = "ModuleScript" },
}

local forbiddenCharSet = {
	[":"] = true,
	["*"] = true,
	["?"] = true,
	['"'] = true,
	["<"] = true,
	[">"] = true,
	["|"] = true,
	["/"] = true,
	["\\"] = true,
}

local function projectInvalid(message: string)
	return { code = "PROJECT_INVALID", message = message }
end

local function hasForbiddenChar(stem: string): boolean
	for i = 1, #stem do
		if forbiddenCharSet[string.sub(stem, i, i)] then
			return true
		end
	end
	return false
end

local function validateStem(stem: string): any?
	if stem == "" or stem == "." or stem == ".." then
		return projectInvalid("forbidden file stem")
	end
	if hasForbiddenChar(stem) then
		return projectInvalid("forbidden file stem")
	end
	return nil
end

local function endsWith(value: string, suffix: string): boolean
	return string.sub(value, -#suffix) == suffix
end

local function classifySourceFile(fileName: string, parentDirName: string): (any?, any?)
	local initClass = initFiles[fileName]
	if initClass ~= nil then
		local err = validateStem(parentDirName)
		if err ~= nil then
			return nil, err
		end
		return { instanceName = parentDirName, className = initClass, init = true }, nil
	end
	for _, spec in leafSuffixes do
		if endsWith(fileName, spec.suffix) then
			local stem = string.sub(fileName, 1, #fileName - #spec.suffix)
			local err = validateStem(stem)
			if err ~= nil then
				return nil, err
			end
			return { instanceName = stem, className = spec.className, init = false }, nil
		end
	end
	return nil, nil
end

local function inverseDiskRelative(className: string, instanceName: string, mappedChildren: boolean, existingLua: boolean): string
	if className == "Folder" then
		return instanceName .. "/"
	end
	local ext = if existingLua then ".lua" else ".luau"
	if mappedChildren then
		if className == "Script" then
			return instanceName .. "/init.server" .. ext
		elseif className == "LocalScript" then
			return instanceName .. "/init.client" .. ext
		end
		return instanceName .. "/init" .. ext
	end
	if className == "Script" then
		return instanceName .. ".server" .. ext
	elseif className == "LocalScript" then
		return instanceName .. ".client" .. ext
	end
	return instanceName .. ext
end

local function trimSlash(path: string): string
	if string.sub(path, -1) == "/" then
		return string.sub(path, 1, -2)
	end
	return path
end

local function isDirPath(path: string): boolean
	return string.sub(path, -1) == "/"
end

local function dirname(path: string): string
	local parent = string.match(path, "^(.*)/[^/]+$")
	return parent or ""
end

local function basename(path: string): string
	return string.match(path, "([^/]+)$") or path
end

local function startsWith(value: string, prefix: string): boolean
	return string.sub(value, 1, #prefix) == prefix
end

local function findRoot(project: any, diskPath: string): any?
	local normalized = trimSlash(diskPath)
	local best = nil
	local bestLen = -1
	for _, root in project.roots do
		if normalized == root.disk or startsWith(normalized, root.disk .. "/") then
			if #root.disk > bestLen then
				best = root
				bestLen = #root.disk
			end
		end
	end
	return best
end

local function parentNodeId(nodeId: string): string
	return (string.match(nodeId, "^(.*)/[^/]+$"))
end

local function studioPathFor(root: any, nodeId: string): string
	local rest = string.match(nodeId, "^[^/]+/(.*)$")
	if rest == nil or rest == "" then
		return root.studio
	end
	return root.studio .. "." .. string.gsub(rest, "/", ".")
end

local function relativeToRoot(root: any, diskPath: string): string
	local normalized = trimSlash(diskPath)
	if normalized == root.disk then
		return ""
	end
	return string.sub(normalized, #root.disk + 2)
end

local function nodesFromTree(project: any, paths: { string }): (any?, any?)
	local filesByDir: { [string]: { string } } = {}
	local dirs: { [string]: boolean } = {}

	local function noteDir(dirPath: string)
		if dirPath ~= "" then
			dirs[dirPath] = true
			local parent = dirname(dirPath)
			if parent ~= "" then
				noteDir(parent)
			end
		end
	end

	for _, path in paths do
		local root = findRoot(project, path)
		if root == nil then
			continue
		end
		if isDirPath(path) then
			noteDir(trimSlash(path))
		else
			local parent = dirname(path)
			filesByDir[parent] = filesByDir[parent] or {}
			table.insert(filesByDir[parent], basename(path))
			noteDir(parent)
		end
	end

	for dirPath, names in filesByDir do
		local parentName = basename(dirPath)
		local initCount = 0
		local seen: { [string]: number } = {}
		for _, name in names do
			local classified, err = classifySourceFile(name, parentName)
			if err ~= nil then
				return nil, err
			end
			if classified == nil then
				continue
			end
			if classified.init then
				initCount += 1
			end
			local key = classified.instanceName .. "\0" .. classified.className
			seen[key] = (seen[key] or 0) + 1
		end
		if initCount > 1 then
			return nil, projectInvalid("directory has more than one init source file")
		end
		for _, n in seen do
			if n > 1 then
				return nil, projectInvalid("duplicate lua and luau sources")
			end
		end
	end

	local function dirHasInit(dirPath: string): boolean
		local names = filesByDir[dirPath]
		if names == nil then
			return false
		end
		for _, name in names do
			if initFiles[name] ~= nil then
				return true
			end
		end
		return false
	end

	local nodes = {}
	local emitted: { [string]: boolean } = {}

	local function emit(node: any)
		if emitted[node.nodeId] then
			return
		end
		emitted[node.nodeId] = true
		table.insert(nodes, node)
	end

	for dirPath, names in filesByDir do
		local root = findRoot(project, dirPath)
		if root == nil then
			continue
		end
		local parentName = basename(dirPath)
		for _, name in names do
			local classified, err = classifySourceFile(name, parentName)
			if err ~= nil then
				return nil, err
			end
			if classified == nil then
				continue
			end
			local filePath = dirPath .. "/" .. name
			local rel = relativeToRoot(root, filePath)
			local instanceRel
			if classified.init then
				local dirRel = dirname(rel)
				if dirRel == "" then
					instanceRel = classified.instanceName
				else
					instanceRel = dirRel
				end
			else
				local dirRel = dirname(rel)
				if dirRel == "" then
					instanceRel = classified.instanceName
				else
					instanceRel = dirRel .. "/" .. classified.instanceName
				end
			end
			local nodeId = root.id .. "/" .. instanceRel
			emit({
				nodeId = nodeId,
				className = classified.className,
				diskPath = filePath,
				studioPath = studioPathFor(root, nodeId),
				parentNodeId = parentNodeId(nodeId),
			})
		end
	end

	for dirPath in dirs do
		local root = findRoot(project, dirPath)
		if root == nil then
			continue
		end
		if dirPath == root.disk then
			continue
		end
		if dirHasInit(dirPath) then
			continue
		end
		local rel = relativeToRoot(root, dirPath)
		if rel == "" then
			continue
		end
		local nodeId = root.id .. "/" .. rel
		emit({
			nodeId = nodeId,
			className = "Folder",
			diskPath = dirPath,
			studioPath = studioPathFor(root, nodeId),
			parentNodeId = parentNodeId(nodeId),
		})
	end

	return nodes, nil
end

return {
	classifySourceFile = classifySourceFile,
	inverseDiskRelative = inverseDiskRelative,
	nodesFromTree = nodesFromTree,
}
