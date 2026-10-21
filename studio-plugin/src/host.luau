-- Studio plugin host: toolbar, dock widget, HttpService adapter (§6.3).

local function siblingRequire(name: string, fsPath: string): any
	local ok, mod = pcall(function()
		return require(script.Parent:WaitForChild(name))
	end)
	if ok then
		return mod
	end
	return require(fsPath)
end

local Session = siblingRequire("session", "./session")
local Studio = siblingRequire("studio", "./studio")

local LABELS = {
	"Connect",
	"Disconnect",
	"Disk wins",
	"Roblox Studio wins",
	"Cancel",
	"Keep disk",
	"Keep Roblox Studio",
}

local function isLoopback(host: string): boolean
	return host == "127.0.0.1" or host == "::1"
end

local function hostHeader(host: string): string
	if host == "::1" then
		return "[::1]"
	end
	return host
end

local function urlFor(host: string, port: number, path: string, token: string?, since: number?): string
	local parts = {}
	if since ~= nil then
		table.insert(parts, "since=" .. tostring(since))
	end
	if token ~= nil and token ~= "" then
		table.insert(parts, "accessToken=" .. tostring(token))
	end
	local q = ""
	if #parts > 0 then
		q = "?" .. table.concat(parts, "&")
	end
	return "http://" .. hostHeader(host) .. ":" .. tostring(port) .. path .. q
end

local function httpFromService(httpService: any, host: string, port: number, token: string)
	return {
		request = function(req: any)
			local authedToken = req.queryToken
			local url = urlFor(host, port, req.path, authedToken, req.since)
			local body = nil
			if req.body ~= nil then
				body = httpService:JSONEncode(req.body)
			end
			local requestOptions = {
				Url = url,
				Method = req.method,
				Headers = { ["Content-Type"] = "application/json" },
			}
			if body ~= nil then
				requestOptions.Body = body
			end
			local result = httpService:RequestAsync(requestOptions)
			local decoded = {}
			if result ~= nil and result.Body ~= nil and result.Body ~= "" then
				local decOk, dec = pcall(function()
					return httpService:JSONDecode(result.Body)
				end)
				if decOk then
					decoded = dec
				end
			end
			return { status = result.StatusCode, body = decoded }
		end,
		openStream = function(path: string, onMessage: ((any) -> ())?): boolean
			local url = urlFor(host, port, path, token)
			local streamType = nil
			pcall(function()
				streamType = (Enum :: any).WebStreamClientType.SSE
			end)
			local ok, client = pcall(function()
				return httpService:CreateWebStreamClient(streamType, {
					Url = url,
					Method = "GET",
				})
			end)
			if not ok or client == nil then
				return false
			end
			if onMessage ~= nil then
				pcall(function()
					client.MessageReceived:Connect(function(message: string)
						local event = message
						if type(message) == "string" then
							local json = message
							local brace = string.find(message, "{")
							if brace ~= nil then
								json = string.sub(message, brace)
							end
							local decOk, decoded = pcall(function()
								return httpService:JSONDecode(json)
							end)
							if decOk then
								event = decoded
							end
						end
						onMessage(event)
					end)
				end)
			end
			return true
		end,
	}
end

local function tryGlobalGame(): any?
	local ok, result = pcall(function()
		return game
	end)
	if ok then
		return result
	end
	return nil
end

local function tryGetService(dataModel: any, name: string): any?
	if dataModel == nil then
		return nil
	end
	local ok, svc = pcall(function()
		return dataModel:GetService(name)
	end)
	if ok then
		return svc
	end
	return nil
end

local function defaultStudio(deps: any): any
	if deps.studio ~= nil then
		return deps.studio
	end
	local dataModel = deps.dataModel or tryGlobalGame()
	return Studio.new({
		dataModel = dataModel,
		scriptEditorService = deps.scriptEditorService or tryGetService(dataModel, "ScriptEditorService"),
		newInstance = deps.newInstance,
	})
end

local function defaultClock()
	return {
		nowMs = function()
			return math.floor(os.clock() * 1000)
		end,
	}
end

local function tryInstance(className: string, factory: ((string) -> any)?): any?
	local ok, inst = pcall(function()
		if factory ~= nil then
			return factory(className)
		end
		return Instance.new(className)
	end)
	if ok then
		return inst
	end
	return nil
end

local function createDockWidget(pluginApi: any)
	local info = nil
	pcall(function()
		info = DockWidgetPluginGuiInfo.new(Enum.InitialDockState.Float, false, false, 320, 480, 250, 250)
	end)
	local widget = nil
	if pluginApi.CreateDockWidgetPluginGuiAsync then
		local ok, result = pcall(function()
			return pluginApi:CreateDockWidgetPluginGuiAsync("rbx-sync", info)
		end)
		if ok then
			widget = result
		end
	end
	if widget == nil and pluginApi.CreateDockWidgetPluginGui then
		local ok, result = pcall(function()
			return pluginApi:CreateDockWidgetPluginGui("rbx-sync", info)
		end)
		if ok then
			widget = result
		end
	end
	if widget ~= nil then
		widget.Title = "rbx-sync"
	end
	return widget
end

local function addTextBox(parent: any, placeholder: string, text: string, layoutOrder: number, factory: ((string) -> any)?): any?
	local box = tryInstance("TextBox", factory)
	if box == nil then
		return nil
	end
	box.PlaceholderText = placeholder
	box.Text = text
	box.ClearTextOnFocus = false
	pcall(function()
		box.Size = UDim2.new(1, -8, 0, 24)
	end)
	box.LayoutOrder = layoutOrder
	box.Parent = parent
	return box
end

local function addButton(parent: any, text: string, layoutOrder: number, onClick: () -> (), factory: ((string) -> any)?): any?
	local button = tryInstance("TextButton", factory)
	if button == nil then
		return nil
	end
	button.Text = text
	pcall(function()
		button.Size = UDim2.new(1, -8, 0, 24)
	end)
	button.LayoutOrder = layoutOrder
	button.Parent = parent
	button.MouseButton1Click:Connect(onClick)
	return button
end

local function addStatus(parent: any, layoutOrder: number, factory: ((string) -> any)?): any?
	local label = tryInstance("TextLabel", factory)
	if label == nil then
		return nil
	end
	label.Text = "disconnected"
	pcall(function()
		label.Size = UDim2.new(1, -8, 0, 36)
	end)
	label.TextWrapped = true
	label.BackgroundTransparency = 1
	local okColor, color = pcall(function()
		return Color3.fromRGB(245, 245, 245)
	end)
	if okColor then
		label.TextColor3 = color
	else
		label.TextColor3 = { R = 245 / 255, G = 245 / 255, B = 245 / 255 }
	end
	label.LayoutOrder = layoutOrder
	label.Parent = parent
	return label
end

local Host = {}

function Host.start(deps: any)
	local pluginApi = deps.plugin
	local httpService = deps.httpService
	local app = {
		hostValue = "127.0.0.1",
		portValue = 29850,
		tokenValue = "",
		visibleTexts = LABELS,
		status = "disconnected",
		httpService = httpService,
		studio = defaultStudio(deps),
		clock = deps.clock or defaultClock(),
		pluginVersion = deps.pluginVersion or "0.0.0",
		session = nil,
	}

	local toolbar = pluginApi:CreateToolbar("rbx-sync")
	local button = toolbar:CreateButton("rbx-sync", "Open rbx-sync", "", "rbx-sync")
	button.ClickableWhenViewportHidden = true

	local widget = createDockWidget(pluginApi)
	if widget then
		app.widget = widget
	end

	local guiFactory = deps.newGuiInstance
	local hostBox, portBox, tokenBox, statusLabel
	if widget then
		local list = tryInstance("UIListLayout", guiFactory)
		if list then
			pcall(function()
				list.SortOrder = Enum.SortOrder.LayoutOrder
				list.Padding = UDim.new(0, 4)
			end)
			list.Parent = widget
		end
		hostBox = addTextBox(widget, "host", app.hostValue, 1, guiFactory)
		portBox = addTextBox(widget, "port", tostring(app.portValue), 2, guiFactory)
		tokenBox = addTextBox(widget, "token", app.tokenValue, 3, guiFactory)
		statusLabel = addStatus(widget, 4, guiFactory)
		addButton(widget, "Connect", 5, function()
			app:clickConnect()
		end, guiFactory)
		addButton(widget, "Disconnect", 6, function()
			app:clickDisconnect()
		end, guiFactory)
		addButton(widget, "Disk wins", 7, function()
			app:clickDiskWins()
		end, guiFactory)
		addButton(widget, "Roblox Studio wins", 8, function()
			app:clickStudioWins()
		end, guiFactory)
		addButton(widget, "Cancel", 9, function()
			app:clickCancel()
		end, guiFactory)
		addButton(widget, "Keep disk", 10, function()
			app:clickKeepDisk()
		end, guiFactory)
		addButton(widget, "Keep Roblox Studio", 11, function()
			app:clickKeepRobloxStudio()
		end, guiFactory)
	end

	local function refreshTexts()
		local texts = {}
		for _, label in LABELS do
			table.insert(texts, label)
		end
		table.insert(texts, app.status)
		app.visibleTexts = texts
		if statusLabel then
			statusLabel.Text = app.status
		end
	end

	button.Click:Connect(function()
		if widget then
			widget.Enabled = not widget.Enabled
		end
	end)

	function app:clickConnect()
		if hostBox then
			self.hostValue = (hostBox.Text or ""):gsub("^%s+", ""):gsub("%s+$", "")
		end
		if portBox then
			self.portValue = tonumber(portBox.Text) or self.portValue
		end
		if tokenBox then
			self.tokenValue = tokenBox.Text
		end
		if not isLoopback(tostring(self.hostValue)) then
			self.status = "host must be loopback"
			refreshTexts()
			return
		end
		local port = tonumber(self.portValue) or 29850
		local token = tostring(self.tokenValue or "")
		local http = httpFromService(self.httpService, tostring(self.hostValue), port, token)
		local ok, result = pcall(function()
			return Session.connect({
				host = tostring(self.hostValue),
				port = port,
				token = token,
				pluginVersion = self.pluginVersion,
				http = http,
				studio = self.studio,
				clock = self.clock,
			})
		end)
		if not ok then
			self.session = nil
			self.status = "disconnected"
			refreshTexts()
			return
		end
		self.session = result
		self.status = self.session.status
		refreshTexts()
	end

	function app:clickCancel()
		if self.session then
			self.session:chooseBootstrap("cancel")
			self.status = self.session.status
		end
		refreshTexts()
	end

	function app:clickDiskWins()
		if self.session then
			self.session:chooseBootstrap("disk")
			self.status = self.session.status
		end
		refreshTexts()
	end

	function app:clickStudioWins()
		if self.session then
			self.session:chooseBootstrap("studio")
			self.status = self.session.status
		end
		refreshTexts()
	end

	function app:clickDisconnect()
		if self.session then
			self.session:onStreamEnd("shutdown")
		end
		self.status = "disconnected"
		refreshTexts()
	end

	local function firstConflictNode(): string?
		if app.session == nil then
			return nil
		end
		for nodeId in app.session.conflicts do
			return nodeId
		end
		return nil
	end

	function app:firstConflictNodeId()
		return firstConflictNode()
	end

	function app:clickKeepDisk()
		local nodeId = firstConflictNode()
		if self.session and nodeId then
			self.session:resolveConflict(nodeId, "keep-disk")
		end
	end

	function app:clickKeepRobloxStudio()
		local nodeId = firstConflictNode()
		if self.session and nodeId then
			self.session:resolveConflict(nodeId, "keep-studio")
		end
	end

	local function onHeartbeat()
		if app.session and app.session.tick then
			app.session:tick()
		end
	end
	if deps.heartbeatConnect then
		deps.heartbeatConnect(onHeartbeat)
	else
		local runService = tryGetService(tryGlobalGame(), "RunService")
		if runService and runService.Heartbeat then
			pcall(function()
				runService.Heartbeat:Connect(onHeartbeat)
			end)
		end
	end

	return app
end

return Host
