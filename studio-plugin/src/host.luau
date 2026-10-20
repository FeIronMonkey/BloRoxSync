-- Studio plugin host: toolbar, dock widget, HttpService adapter (§6.3).

local Session = require("./session")

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

local function urlFor(host: string, port: number, path: string, token: string?): string
	local q = ""
	if token ~= nil and token ~= "" then
		q = "?accessToken=" .. tostring(token)
	end
	return "http://" .. hostHeader(host) .. ":" .. tostring(port) .. path .. q
end

local function httpFromService(httpService: any, host: string, port: number, token: string)
	return {
		request = function(req: any)
			local authedToken = req.queryToken
			local url = urlFor(host, port, req.path, authedToken)
			local body = nil
			if req.body ~= nil then
				body = httpService:JSONEncode(req.body)
			end
			local result = httpService:RequestAsync({
				Url = url,
				Method = req.method,
				Headers = { ["Content-Type"] = "application/json" },
				Body = body,
			})
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
		openStream = function(path: string): boolean
			local url = urlFor(host, port, path, token)
			local ok, client = pcall(function()
				return httpService:CreateWebStreamClient(nil, {
					Url = url,
					Method = "GET",
				})
			end)
			return ok and client ~= nil
		end,
	}
end

local function stubStudio()
	return {
		resolveRoot = function(): boolean
			return false
		end,
		mappedNodes = function()
			return {}
		end,
		startListeners = function() end,
		isEditorOpen = function(): boolean
			return false
		end,
		getEditorSource = function(): string
			return ""
		end,
		getInstanceSource = function(): string
			return ""
		end,
		updateEditorSource = function() end,
		setInstanceSource = function() end,
		applyNodes = function() end,
		createInstance = function() end,
		destroyInstance = function() end,
	}
end

local function defaultClock()
	return {
		nowMs = function()
			return math.floor(os.clock() * 1000)
		end,
	}
end

local function tryInstance(className: string): any?
	local ok, inst = pcall(function()
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

local function addTextBox(parent: any, placeholder: string, text: string, layoutOrder: number): any?
	local box = tryInstance("TextBox")
	if box == nil then
		return nil
	end
	box.PlaceholderText = placeholder
	box.Text = text
	box.ClearTextOnFocus = false
	box.Size = UDim2.new(1, -8, 0, 24)
	box.LayoutOrder = layoutOrder
	box.Parent = parent
	return box
end

local function addButton(parent: any, text: string, layoutOrder: number, onClick: () -> ()): any?
	local button = tryInstance("TextButton")
	if button == nil then
		return nil
	end
	button.Text = text
	button.Size = UDim2.new(1, -8, 0, 24)
	button.LayoutOrder = layoutOrder
	button.Parent = parent
	button.MouseButton1Click:Connect(onClick)
	return button
end

local function addStatus(parent: any, layoutOrder: number): any?
	local label = tryInstance("TextLabel")
	if label == nil then
		return nil
	end
	label.Text = "disconnected"
	label.Size = UDim2.new(1, -8, 0, 36)
	label.TextWrapped = true
	label.BackgroundTransparency = 1
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
		studio = deps.studio or stubStudio(),
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

	local hostBox, portBox, tokenBox, statusLabel
	if widget then
		local list = tryInstance("UIListLayout")
		if list then
			pcall(function()
				list.SortOrder = Enum.SortOrder.LayoutOrder
				list.Padding = UDim.new(0, 4)
			end)
			list.Parent = widget
		end
		hostBox = addTextBox(widget, "host", app.hostValue, 1)
		portBox = addTextBox(widget, "port", tostring(app.portValue), 2)
		tokenBox = addTextBox(widget, "token", app.tokenValue, 3)
		statusLabel = addStatus(widget, 4)
		addButton(widget, "Connect", 5, function()
			app:clickConnect()
		end)
		addButton(widget, "Disconnect", 6, function()
			app:clickDisconnect()
		end)
		addButton(widget, "Disk wins", 7, function()
			if app.session then
				app.session:chooseBootstrap("disk")
			end
		end)
		addButton(widget, "Roblox Studio wins", 8, function()
			if app.session then
				app.session:chooseBootstrap("studio")
			end
		end)
		addButton(widget, "Cancel", 9, function()
			if app.session then
				app.session:chooseBootstrap("cancel")
			end
		end)
		addButton(widget, "Keep disk", 10, function() end)
		addButton(widget, "Keep Roblox Studio", 11, function() end)
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
			self.hostValue = hostBox.Text
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
		self.session = Session.connect({
			host = tostring(self.hostValue),
			port = port,
			token = token,
			pluginVersion = self.pluginVersion,
			http = http,
			studio = self.studio,
			clock = self.clock,
		})
		self.status = self.session.status
		refreshTexts()
	end

	function app:clickDisconnect()
		if self.session then
			self.session:onStreamEnd("shutdown")
		end
		self.status = "disconnected"
		refreshTexts()
	end

	return app
end

return Host
