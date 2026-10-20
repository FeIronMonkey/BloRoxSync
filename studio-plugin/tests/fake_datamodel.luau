--> L4 fake DataModel / ScriptEditorService. No Roblox Studio.

local M = {}

local function connectable()
	local listeners = {}
	return {
		Connect = function(_, fn: (...any) -> ())
			table.insert(listeners, fn)
			return {
				Disconnect = function()
					for i, existing in listeners do
						if existing == fn then
							table.remove(listeners, i)
							return
						end
					end
				end,
			}
		end,
		Fire = function(_, ...: any)
			for _, fn in listeners do
				fn(...)
			end
		end,
	}
end

function M.new()
	local byRef = {}

	local function newInstance(className: string, name: string?)
		local childAdded = connectable()
		local childRemoved = connectable()
		local descendantAdded = connectable()
		local descendantRemoving = connectable()
		local nameChanged = connectable()
		local sourceChanged = connectable()
		local inst
		inst = {
			ClassName = className,
			_name = name or className,
			_source = "",
			_parent = nil,
			_children = {},
			_destroyed = false,
			ChildAdded = childAdded,
			ChildRemoved = childRemoved,
			DescendantAdded = descendantAdded,
			DescendantRemoving = descendantRemoving,
		}
		function inst:GetChildren()
			local out = {}
			for _, c in self._children do
				table.insert(out, c)
			end
			return out
		end
		function inst:GetDescendants()
			local out = {}
			local function walk(node: any)
				for _, c in node._children do
					table.insert(out, c)
					walk(c)
				end
			end
			walk(self)
			return out
		end
		function inst:FindFirstChild(childName: string)
			for _, c in self._children do
				if c.Name == childName then
					return c
				end
			end
			return nil
		end
		function inst:GetPropertyChangedSignal(prop: string)
			if prop == "Name" then
				return nameChanged
			end
			if prop == "Source" then
				return sourceChanged
			end
			return connectable()
		end
		function inst:Destroy()
			self._destroyed = true
			self.Parent = nil
		end
		setmetatable(inst, {
			__index = function(t, k)
				if k == "Parent" then
					return rawget(t, "_parent")
				end
				if k == "Name" then
					return rawget(t, "_name")
				end
				if k == "Source" then
					return rawget(t, "_source")
				end
				return rawget(t, k)
			end,
			__newindex = function(t, k, v)
				if k == "Parent" then
					local old = rawget(t, "_parent")
					if old == v then
						return
					end
					if old ~= nil then
						local kept = {}
						for _, c in old._children do
							if c ~= t then
								table.insert(kept, c)
							end
						end
						old._children = kept
						old.ChildRemoved:Fire(t)
						local climb = old
						while climb do
							climb.DescendantRemoving:Fire(t)
							climb = climb.Parent
						end
					end
					rawset(t, "_parent", v)
					if v ~= nil then
						table.insert(v._children, t)
						v.ChildAdded:Fire(t)
						local climb = v
						while climb do
							climb.DescendantAdded:Fire(t)
							climb = climb.Parent
						end
					end
					return
				end
				if k == "Name" then
					rawset(t, "_name", v)
					nameChanged:Fire()
					return
				end
				if k == "Source" then
					rawset(t, "_source", v)
					sourceChanged:Fire()
					return
				end
				rawset(t, k, v)
			end,
		})
		table.insert(byRef, inst)
		return inst
	end

	local game = newInstance("DataModel", "game")
	function game:GetService(serviceName: string)
		local existing = self:FindFirstChild(serviceName)
		if existing then
			return existing
		end
		local created = newInstance(serviceName, serviceName)
		created.Parent = self
		return created
	end

	local sss = newInstance("ServerScriptService", "ServerScriptService")
	sss.Parent = game

	local openDocs: { [any]: any } = {}
	local textChanged = connectable()
	local scriptEditor = {
		TextDocumentDidChange = textChanged,
	}
	function scriptEditor:FindScriptDocument(scriptInst: any)
		return openDocs[scriptInst]
	end
	function scriptEditor:GetEditorSource(scriptInst: any)
		local doc = openDocs[scriptInst]
		if doc then
			return doc._text
		end
		return scriptInst.Source
	end
	function scriptEditor:UpdateSourceAsync(scriptInst: any, callback: (string) -> string)
		local doc = openDocs[scriptInst]
		local old = if doc then doc._text else scriptInst.Source
		local nextText = callback(old)
		if doc then
			doc._text = nextText
		end
		scriptInst.Source = nextText
	end
	function scriptEditor:Open(scriptInst: any, text: string?)
		local doc = {
			_text = text or scriptInst.Source,
			IsCommandBar = function()
				return false
			end,
			GetScript = function()
				return scriptInst
			end,
		}
		openDocs[scriptInst] = doc
		return doc
	end
	function scriptEditor:Close(scriptInst: any)
		openDocs[scriptInst] = nil
	end
	function scriptEditor:FireTextChanged(scriptInst: any)
		local doc = openDocs[scriptInst]
		textChanged:Fire(doc, {})
	end

	return {
		game = game,
		sss = sss,
		scriptEditor = scriptEditor,
		newInstance = newInstance,
	}
end

return M
