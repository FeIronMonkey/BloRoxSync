--> L4 DataModel adapter tests (Lune). Fake instance tree; no Roblox Studio.

local FakeDataModel = require("./fake_datamodel")

local studioModule
local studioLoadErr
do
	local ok, result = pcall(function()
		return require("../src/studio")
	end)
	if ok then
		studioModule = result
	else
		studioLoadErr = result
	end
end

local function studioApi()
	if studioLoadErr ~= nil then
		error("studio-plugin/src/studio.luau: " .. tostring(studioLoadErr))
	end
	return studioModule
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

local function newAdapter(world: any)
	return studioApi().new({
		dataModel = world.game,
		scriptEditorService = world.scriptEditor,
		newInstance = world.newInstance,
	})
end

local tests = {}

local function add(name: string, fn: () -> ())
	table.insert(tests, { name = name, fn = fn })
end

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution
add("PluginStudioResolvesRoot_WhenServiceExists", function()
	local world = FakeDataModel.new()
	local adapter = newAdapter(world)
	assertTrue(adapter.resolveRoot("ServerScriptService") == true, "service root")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution
add("PluginStudioSkipsRoot_WhenStudioPathMissing", function()
	local world = FakeDataModel.new()
	local adapter = newAdapter(world)
	assertTrue(adapter.resolveRoot("ReplicatedStorage.Shared") == false, "missing intermediate")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginStudioApplyNodesCreatesNestedRootFolder_WhenStudioPathMissing", function()
	local world = FakeDataModel.new()
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService.RbxSyncL5" } })
	adapter.applyNodes({
		{
			nodeId = "server/Hello",
			className = "Script",
			source = "print(l5)\n",
		},
	})
	local folder = world.sss:FindFirstChild("RbxSyncL5")
	assertTrue(folder ~= nil, "disk-to-studio creates missing studio Folder")
	assertEq(folder.ClassName, "Folder", "intermediate is Folder")
	local hello = folder:FindFirstChild("Hello")
	assertTrue(hello ~= nil, "Script under nested root")
	assertEq(hello.ClassName, "Script", "class")
	assertEq(hello.Source, "print(l5)\n", "source")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#roblox-studio-path-resolution
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginStudioCreateInstanceCreatesNestedRootFolder_WhenStudioPathMissing", function()
	local world = FakeDataModel.new()
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService.RbxSyncL5" } })
	adapter.createInstance("server/Hello", "Script")
	local folder = world.sss:FindFirstChild("RbxSyncL5")
	assertTrue(folder ~= nil, "live create creates missing studio Folder")
	assertEq(folder.ClassName, "Folder", "intermediate is Folder")
	local hello = folder:FindFirstChild("Hello")
	assertTrue(hello ~= nil, "Script under nested root")
	assertEq(hello.ClassName, "Script", "class")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#what-is-synced
add("PluginStudioMappedNodesIncludeScript_WhenDescendantOfRoot", function()
	local world = FakeDataModel.new()
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Source = "print(1)\n"
	scriptInst.Parent = world.sss
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	local nodes = adapter.mappedNodesForRoot("server")
	assertEq(#nodes, 1, "one mapped script")
	assertEq(nodes[1].nodeId, "server/Foo", "nodeId")
	assertEq(nodes[1].className, "Script", "class")
	assertEq(nodes[1].source, "print(1)\n", "source")
end)

-- Spec-Anchor: proj-spec/chapters/06.4-project-mapping.md#what-is-synced
add("PluginStudioMappedNodesOmitUnmappedPart_WhenUnderRoot", function()
	local world = FakeDataModel.new()
	local part = world.newInstance("Part", "Block")
	part.Parent = world.sss
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Source = "print(1)\n"
	scriptInst.Parent = world.sss
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	local nodes = adapter.mappedNodesForRoot("server")
	assertEq(#nodes, 1, "Part is not mapped")
	assertEq(nodes[1].nodeId, "server/Foo", "only script")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
add("PluginStudioApplyNodesCreatesScript_WhenUpsertNewNodeId", function()
	local world = FakeDataModel.new()
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	adapter.applyNodes({
		{
			nodeId = "server/Foo",
			className = "Script",
			source = "print(disk)\n",
		},
	})
	local foo = world.sss:FindFirstChild("Foo")
	assertTrue(foo ~= nil, "created Script")
	assertEq(foo.ClassName, "Script", "class")
	assertEq(foo.Source, "print(disk)\n", "closed-tab Source")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
add("PluginStudioApplyNodesWritesEditor_WhenTabOpen", function()
	local world = FakeDataModel.new()
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Source = "print(old)\n"
	scriptInst.Parent = world.sss
	world.scriptEditor:Open(scriptInst, "print(old)\n")
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	adapter.applyNodes({
		{
			nodeId = "server/Foo",
			className = "Script",
			source = "print(new)\n",
		},
	})
	assertEq(world.scriptEditor:GetEditorSource(scriptInst), "print(new)\n", "UpdateSourceAsync path")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginStudioApplyNodesDestroysMapped_WhenDelete", function()
	local world = FakeDataModel.new()
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Parent = world.sss
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	adapter.applyNodes({})
	assertTrue(world.sss:FindFirstChild("Foo") == nil, "mapped Script deleted when omitted")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#creating-and-deleting-instances
add("PluginStudioDoesNotDestroyUnmapped_WhenDeleteMappedSibling", function()
	local world = FakeDataModel.new()
	local part = world.newInstance("Part", "Block")
	part.Parent = world.sss
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Parent = world.sss
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	adapter.destroyInstance("server/Foo")
	assertTrue(world.sss:FindFirstChild("Foo") == nil, "mapped gone")
	assertTrue(world.sss:FindFirstChild("Block") ~= nil, "unmapped Part remains")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#roblox-studio-change-detection
add("PluginStudioStartListenersEmitsMutation_WhenNameChanges", function()
	local world = FakeDataModel.new()
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Source = "print(1)\n"
	scriptInst.Parent = world.sss
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	local seen = {}
	adapter.startListeners(function(change)
		table.insert(seen, change)
	end)
	scriptInst.Name = "Bar"
	assertTrue(#seen >= 1, "name change observed")
	local last = seen[#seen]
	assertEq(last.op, "rename", "rename op")
	assertEq(last.nodeId, "server/Foo", "old nodeId")
	assertEq(last.newNodeId, "server/Bar", "new nodeId")
end)

-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#reading-and-writing-source
-- Spec-Anchor: proj-spec/chapters/06.3-studio-plugin.md#roblox-studio-change-detection
add("PluginStudioStartListenersEmitsMutation_WhenEditorChanges", function()
	local world = FakeDataModel.new()
	local scriptInst = world.newInstance("Script", "Foo")
	scriptInst.Source = "print(old)\n"
	scriptInst.Parent = world.sss
	world.scriptEditor:Open(scriptInst, "print(old)\n")
	local adapter = newAdapter(world)
	adapter.setRoots({ { id = "server", studio = "ServerScriptService" } })
	local seen = {}
	adapter.startListeners(function(change)
		table.insert(seen, change)
	end)
	world.scriptEditor:Open(scriptInst, "print(new)\n")
	world.scriptEditor:FireTextChanged(scriptInst)
	assertTrue(#seen >= 1, "editor change observed")
	assertEq(seen[#seen].op, "upsert", "upsert")
	assertEq(seen[#seen].source, "print(new)\n", "GetEditorSource")
end)

return tests
