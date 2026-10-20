local fs = require("@lune/fs")
local process = require("@lune/process")
local serde = require("@lune/serde")

local function join(...)
	return table.concat({ ... }, "/")
end

local function posix(path: string): string
	return (string.gsub(path, "\\", "/"))
end

local function isRepoRoot(dir: string): boolean
	return fs.isDir(join(dir, "proj-spec", "chapters"))
		and fs.isDir(join(dir, "sync-protocol"))
		and fs.isDir(join(dir, "studio-plugin"))
end

local function repoRoot(): string
	local dir = posix(process.cwd)
	for _ = 1, 8 do
		if isRepoRoot(dir) then
			return dir
		end
		local parent = string.match(dir, "^(.*)/[^/]+$")
		if parent == nil or parent == dir then
			break
		end
		dir = parent
	end
	error("could not locate repository root from " .. process.cwd)
end

local function mappingFixture(name: string): any
	local path = join(repoRoot(), "sync-protocol", "fixtures", "v1", "mapping", name)
	return serde.decode("json", fs.readFile(path))
end

local mappingModule
local mappingLoadErr
do
	local ok, result = pcall(function()
		return require("../src/mapping")
	end)
	if ok then
		mappingModule = result
	else
		mappingLoadErr = result
	end
end

local function mapping()
	if mappingLoadErr ~= nil then
		error("studio-plugin/src/mapping.luau: " .. tostring(mappingLoadErr))
	end
	return mappingModule
end

local function fail(msg: string)
	error(msg, 2)
end

local function assertEq(got: any, want: any, msg: string)
	if got ~= want then
		fail(string.format("%s: got %s, want %s", msg, tostring(got), tostring(want)))
	end
end

local function nodeById(nodes: { any }, id: string): any?
	for _, node in nodes do
		if node.nodeId == id then
			return node
		end
	end
	return nil
end

local tests = {}

local function add(name: string, fn: () -> ())
	table.insert(tests, { name = name, fn = fn })
end

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
-- Spec-Anchor: proj-spec/chapters/09-design-decisions.md#d9--independent-mapping-names-server--client--init
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
add("PluginMappingClassifySourceNames_WhenFixtureExists", function()
	local doc = mappingFixture("source-names.json")
	local parent = "Bar"
	for _, row in doc.rows do
		local classified, err = mapping().classifySourceFile(row.fileName, parent)
		if err ~= nil then
			fail(string.format("classifySourceFile(%q): %s", row.fileName, tostring(err.message or err)))
		end
		if classified == nil then
			fail("classifySourceFile(" .. row.fileName .. ") returned nil")
		end
		assertEq(classified.className, row.className, row.fileName .. " className")
		if row.nameIsParentDirectory then
			assertEq(classified.instanceName, parent, row.fileName .. " instanceName")
			assertEq(classified.init, true, row.fileName .. " init")
		else
			assertEq(classified.instanceName, row.instanceName, row.fileName .. " instanceName")
			assertEq(classified.init, false, row.fileName .. " init")
		end
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
-- Spec-Anchor: proj-spec/chapters/14-acceptance-criteria.md#protocol-and-mapping
add("PluginMappingNodeIdsMatchSpecExamples_WhenFixtureExists", function()
	local doc = mappingFixture("nodeid-examples.json")
	local paths = {}
	for _, row in doc.rows do
		table.insert(paths, row.diskPath)
	end
	local nodes, err = mapping().nodesFromTree(doc.project, paths)
	if err ~= nil then
		fail("nodesFromTree: " .. tostring(err.message or err))
	end
	for _, row in doc.rows do
		local node = nodeById(nodes, row.nodeId)
		if node == nil then
			fail("missing nodeId " .. row.nodeId)
		end
		assertEq(node.className, row.className, row.nodeId .. " className")
		local wantDisk = string.gsub(row.diskPath, "/$", "")
		assertEq(node.diskPath, wantDisk, row.nodeId .. " diskPath")
		if string.sub(row.nodeId, 1, 1) == "/" then
			fail(row.nodeId .. " has a leading slash")
		end
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
add("PluginMappingStudioPathWalksFromService_WhenNodeIdExamplesExist", function()
	local doc = mappingFixture("nodeid-examples.json")
	local paths = {}
	for _, row in doc.rows do
		table.insert(paths, row.diskPath)
	end
	local nodes, err = mapping().nodesFromTree(doc.project, paths)
	if err ~= nil then
		fail("nodesFromTree: " .. tostring(err.message or err))
	end
	local child = nodeById(nodes, "server/Bar/Child")
	if child == nil then
		fail("missing server/Bar/Child")
	end
	assertEq(child.studioPath, "ServerScriptService.Bar.Child", "server/Bar/Child studioPath")
	local shared = nodeById(nodes, "shared/Config")
	if shared == nil then
		fail("missing shared/Config")
	end
	assertEq(shared.studioPath, "ReplicatedStorage.Shared.Config", "shared/Config studioPath")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#inverse-mapping-roblox-studio--disk
add("PluginMappingInversePrefersLuau_WhenCreatingFiles", function()
	local doc = mappingFixture("inverse-examples.json")
	for _, row in doc.rows do
		local got = mapping().inverseDiskRelative(row.className, "Name", row.mappedChildren, row.existingLua)
		assertEq(
			got,
			row.diskRelative,
			string.format(
				"inverse %s children=%s lua=%s",
				row.className,
				tostring(row.mappedChildren),
				tostring(row.existingLua)
			)
		)
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
add("PluginMappingDuplicateLuaLuauRefuses_WhenSameDirectory", function()
	local doc = mappingFixture("invalid-duplicate-lua-luau.json")
	local paths = {}
	for _, fileName in doc.files do
		table.insert(paths, doc.directory .. "/" .. fileName)
	end
	local _, err = mapping().nodesFromTree(doc.project, paths)
	if err == nil or err.code ~= doc.expectedError then
		fail(string.format("duplicate lua/luau code = %s, want %s", tostring(err and err.code), doc.expectedError))
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
add("PluginMappingAtMostOneInit_WhenDirectoryHasTwoInits", function()
	local doc = mappingFixture("invalid-multiple-init.json")
	local paths = {}
	for _, fileName in doc.files do
		table.insert(paths, doc.directory .. "/" .. fileName)
	end
	local _, err = mapping().nodesFromTree(doc.project, paths)
	if err == nil or err.code ~= doc.expectedError then
		fail(string.format("two inits code = %s, want %s", tostring(err and err.code), doc.expectedError))
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
add("PluginMappingForbiddenStemsRefuse_WhenStemIllegal", function()
	local doc = mappingFixture("invalid-forbidden-stems.json")
	for _, stem in doc.stems do
		local fileName = stem .. ".luau"
		local _, err = mapping().classifySourceFile(fileName, "parent")
		if err == nil or err.code ~= doc.expectedError then
			fail(
				string.format(
					"stem %q file %q code = %s, want %s",
					stem,
					fileName,
					tostring(err and err.code),
					doc.expectedError
				)
			)
		end
	end
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#source-file-names
-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#nodeid
add("PluginMappingFolderVsInitParent_WhenDirectoryHasOrLacksInit", function()
	local doc = mappingFixture("folder-vs-init-parent.json")
	local paths = {}
	for _, row in doc.rows do
		table.insert(paths, row.diskPath)
	end
	local nodes, err = mapping().nodesFromTree(doc.project, paths)
	if err ~= nil then
		fail("nodesFromTree: " .. tostring(err.message or err))
	end
	local bar = nodeById(nodes, "server/Bar")
	if bar == nil then
		fail("missing server/Bar")
	end
	assertEq(bar.className, "ModuleScript", "Bar className")
	local child = nodeById(nodes, "server/Bar/Child")
	if child == nil then
		fail("missing server/Bar/Child")
	end
	assertEq(child.parentNodeId, "server/Bar", "Bar/Child parentNodeId")
	local qux = nodeById(nodes, "server/Qux")
	if qux == nil then
		fail("missing server/Qux")
	end
	assertEq(qux.className, "Folder", "Qux className")
	local qchild = nodeById(nodes, "server/Qux/Child")
	if qchild == nil then
		fail("missing server/Qux/Child")
	end
	assertEq(qchild.parentNodeId, "server/Qux", "Qux/Child parentNodeId")
end)

return tests
