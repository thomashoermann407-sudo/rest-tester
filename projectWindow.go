package main

import (
	"fmt"
	"strings"

	"hoermi.com/rest-test/win32"
)

// Global control handles - Request Panel
type ProjectWindow struct {
	mainWindow     *win32.Window
	tabs           *win32.TabManager
	methodCombo    win32.HWND
	urlInput       win32.HWND
	headersInput   win32.HWND
	queryInput     win32.HWND
	bodyInput      win32.HWND
	responseOutput win32.HWND
	statusLabel    win32.HWND
	sendBtn        win32.HWND
	// Labels for request panel
	methodLabel   win32.HWND
	urlLabel      win32.HWND
	headersLabel  win32.HWND
	queryLabel    win32.HWND
	bodyLabel     win32.HWND
	responseLabel win32.HWND

	// Global control handles - Project View Panel
	projectListBox win32.HWND
	openReqBtn     win32.HWND
	deleteReqBtn   win32.HWND
	projectInfo    win32.HWND
	saveBtn        win32.HWND

	// Global control handles - New Tab Panel
	newTabTitle   win32.HWND
	newTabNewBtn  win32.HWND
	newTabOpenBtn win32.HWND
	recentLabel   win32.HWND
	recentListBox win32.HWND

	// Global control handles - Settings Panel
	certInput     win32.HWND
	keyInput      win32.HWND
	caInput       win32.HWND
	skipVerifyChk win32.HWND
	certBtn       win32.HWND
	keyBtn        win32.HWND
	caBtn         win32.HWND
	certLabel     win32.HWND
	keyLabel      win32.HWND
	caLabel       win32.HWND
	settingsTitle win32.HWND

	// Layout state for resizing
	windowWidth       int32
	windowHeight      int32
	paramsHeightRatio float32
	bodyHeightRatio   float32

	panels *Panels

	// Current loaded project
	currentProject *Project
}

func NewProjectWindow() *ProjectWindow {
	mainWindow := win32.NewWindow("REST Tester", 1100, 850)
	tabs := mainWindow.EnableTabs()
	return &ProjectWindow{
		mainWindow:        mainWindow,
		tabs:              tabs,
		windowWidth:       1100,
		windowHeight:      850,
		paramsHeightRatio: 0.10, // 10% of available height for params/headers
		bodyHeightRatio:   0.15, // 15% for body
	}
}

func (pw *ProjectWindow) createRequestPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(12) // Start right after tab bar (tab bar height handled by control offset)

	// === Request Row ===
	pw.methodLabel = pw.mainWindow.CreateLabel("Method", padding, y+3, 50, labelHeight)
	pw.methodCombo = pw.mainWindow.CreateComboBox(padding+55, y, 90, 200, ID_METHOD_COMBO)
	win32.ComboBoxAddString(pw.methodCombo, "GET")
	win32.ComboBoxAddString(pw.methodCombo, "POST")
	win32.ComboBoxAddString(pw.methodCombo, "PUT")
	win32.ComboBoxAddString(pw.methodCombo, "PATCH")
	win32.ComboBoxAddString(pw.methodCombo, "DELETE")
	win32.ComboBoxAddString(pw.methodCombo, "HEAD")
	win32.ComboBoxAddString(pw.methodCombo, "OPTIONS")
	win32.ComboBoxSetCurSel(pw.methodCombo, 0)

	// URL input
	pw.urlLabel = pw.mainWindow.CreateLabel("URL", padding+155, y+3, 30, labelHeight)
	pw.urlInput = pw.mainWindow.CreateInput(padding+190, y, 680, inputHeight)

	// Send button
	pw.sendBtn = pw.mainWindow.CreateButton("  Send  ", padding+885, y, 90, inputHeight+2, ID_SEND_BTN)

	// === Query Parameters Section ===
	y += inputHeight + padding
	pw.queryLabel = pw.mainWindow.CreateLabel("Query Parameters (one per line: key=value)", padding, y, 300, labelHeight)
	y += labelHeight + 3
	pw.queryInput = pw.mainWindow.CreateCodeEdit(padding, y, 520, 70, false)

	// === Headers Section ===
	pw.headersLabel = pw.mainWindow.CreateLabel("Headers (one per line: Header: value)", padding+540, y-labelHeight-3, 300, labelHeight)
	pw.headersInput = pw.mainWindow.CreateCodeEdit(padding+540, y, 520, 70, false)

	// === Body Section ===
	y += 70 + padding
	pw.bodyLabel = pw.mainWindow.CreateLabel("Request Body", padding, y, 150, labelHeight)
	y += labelHeight + 3
	pw.bodyInput = pw.mainWindow.CreateCodeEdit(padding, y, 1050, 120, false)

	// === Response Section ===
	y += 120 + padding
	pw.responseLabel = pw.mainWindow.CreateLabel("Response", padding, y, 80, labelHeight)
	pw.statusLabel = pw.mainWindow.CreateLabel("Ready", padding+90, y, 400, labelHeight)
	y += labelHeight + 5
	pw.responseOutput = pw.mainWindow.CreateCodeEdit(padding, y, 1050, 350, true)
}

func (pw *ProjectWindow) createProjectViewPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(50) // After toolbar

	// Project info
	pw.projectInfo = pw.mainWindow.CreateLabel("Double-click a request to open it in a new tab", padding, y, 500, labelHeight)
	y += labelHeight + padding

	// List of requests
	pw.projectListBox = pw.mainWindow.CreateListBox(padding, y, 500, 500, ID_PROJECT_LIST)

	// Action buttons
	pw.openReqBtn = pw.mainWindow.CreateButton("Open in Tab", padding+520, y, 120, inputHeight, ID_OPEN_REQ_BTN)
	pw.deleteReqBtn = pw.mainWindow.CreateButton("Delete", padding+520, y+inputHeight+5, 120, inputHeight, ID_DELETE_REQ_BTN)
	pw.saveBtn = pw.mainWindow.CreateButton("Save Project", padding+520, y+inputHeight*2+15, 120, inputHeight, ID_SAVE_BTN)

	// Initially hide this panel
	win32.ShowWindow(pw.projectInfo, win32.SW_HIDE)
	win32.ShowWindow(pw.projectListBox, win32.SW_HIDE)
	win32.ShowWindow(pw.openReqBtn, win32.SW_HIDE)
	win32.ShowWindow(pw.deleteReqBtn, win32.SW_HIDE)
	win32.ShowWindow(pw.saveBtn, win32.SW_HIDE)
}

func (pw *ProjectWindow) createSettingsPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(50) // After toolbar

	// === Client Certificate Section ===
	pw.settingsTitle = pw.mainWindow.CreateLabel("Global Settings", padding, y, 200, labelHeight+4)
	y += labelHeight + padding

	pw.certLabel = pw.mainWindow.CreateLabel("Client Certificate (PEM)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	pw.certInput = pw.mainWindow.CreateInput(padding, y, 400, inputHeight)
	pw.certBtn = pw.mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_CERT_BTN)
	y += inputHeight + padding

	pw.keyLabel = pw.mainWindow.CreateLabel("Private Key (PEM)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	pw.keyInput = pw.mainWindow.CreateInput(padding, y, 400, inputHeight)
	pw.keyBtn = pw.mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_KEY_BTN)
	y += inputHeight + padding

	pw.caLabel = pw.mainWindow.CreateLabel("CA Bundle (optional)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	pw.caInput = pw.mainWindow.CreateInput(padding, y, 400, inputHeight)
	pw.caBtn = pw.mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_CA_BTN)
	y += inputHeight + padding

	pw.skipVerifyChk = pw.mainWindow.CreateCheckbox("Skip TLS Verification (insecure)", padding, y, 280, inputHeight, ID_SKIP_VERIFY)

	// Initially hide this panel
	win32.ShowWindow(pw.settingsTitle, win32.SW_HIDE)
	win32.ShowWindow(pw.certLabel, win32.SW_HIDE)
	win32.ShowWindow(pw.certInput, win32.SW_HIDE)
	win32.ShowWindow(pw.certBtn, win32.SW_HIDE)
	win32.ShowWindow(pw.keyLabel, win32.SW_HIDE)
	win32.ShowWindow(pw.keyInput, win32.SW_HIDE)
	win32.ShowWindow(pw.keyBtn, win32.SW_HIDE)
	win32.ShowWindow(pw.caLabel, win32.SW_HIDE)
	win32.ShowWindow(pw.caInput, win32.SW_HIDE)
	win32.ShowWindow(pw.caBtn, win32.SW_HIDE)
	win32.ShowWindow(pw.skipVerifyChk, win32.SW_HIDE)
}

func (pw *ProjectWindow) createNewTabPanel() {
	padding := int32(12)
	labelHeight := int32(24)
	inputHeight := int32(32)
	btnWidth := int32(150)
	y := int32(80) // After toolbar, centered area

	// Welcome title
	pw.newTabTitle = pw.mainWindow.CreateLabel("REST Tester - Start", padding, y, 400, labelHeight+10)
	y += labelHeight + 30

	// New and Open buttons
	pw.newTabNewBtn = pw.mainWindow.CreateButton("📄 New Project", padding, y, btnWidth, inputHeight, ID_NEW_BTN)
	pw.newTabOpenBtn = pw.mainWindow.CreateButton("📂 Open Project", padding+btnWidth+10, y, btnWidth, inputHeight, ID_OPEN_BTN)
	y += inputHeight + 30

	// Recent projects section
	pw.recentLabel = pw.mainWindow.CreateLabel("Recent Projects:", padding, y, 200, labelHeight)
	y += labelHeight + 5
	pw.recentListBox = pw.mainWindow.CreateListBox(padding, y, 400, 300, ID_RECENT_LIST)
}

// handleResize is called when the window is resized
func (pw *ProjectWindow) handleResize(width, height int32) {
	pw.windowWidth = width
	pw.windowHeight = height
	pw.tabs.SetWidth(width)

	// Get current active tab
	if pw.tabs != nil {
		pw.resizeRequestPanel()
		// TODO: Add resize handlers for other panel types if needed
	}
}

// resizeRequestPanel adjusts all controls in the request panel based on window size
func (pw *ProjectWindow) resizeRequestPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	tabHeight := int32(46) // Tab bar height (matches TabManager.titleBarHeight)

	// Calculate available height (excluding tab bar and padding)
	availableHeight := pw.windowHeight - tabHeight - padding*5 - labelHeight*5 - inputHeight

	// Calculate panel heights based on ratios
	paramsHeight := max(int32(float32(availableHeight)*pw.paramsHeightRatio), 60)

	bodyHeight := max(int32(float32(availableHeight)*pw.bodyHeightRatio), 80)

	responseHeight := max(availableHeight-paramsHeight-bodyHeight, 150)

	// Calculate available width
	availableWidth := pw.windowWidth - padding*2

	// Start position (relative to client area including tab bar offset)
	y := tabHeight + padding

	// === Request Row (fixed height) ===
	// Method combo
	win32.MoveWindow(pw.methodCombo, padding+55, y, 90, 200, true)

	// URL input - expand with window width
	urlWidth := availableWidth - 330 // Leave space for method, label, and send button
	win32.MoveWindow(pw.urlInput, padding+190, y, urlWidth, inputHeight, true)

	// Send button - move to right edge
	win32.MoveWindow(pw.sendBtn, padding+190+urlWidth+15, y, 90, inputHeight+2, true)

	// === Query Parameters & Headers Section ===
	y += inputHeight + padding
	y += labelHeight + 3

	// Split width 50/50 for params and headers
	halfWidth := (availableWidth - padding) / 2

	// Query parameters (left half)
	win32.MoveWindow(pw.queryInput, padding, y, halfWidth, paramsHeight, true)

	// Headers (right half)
	win32.MoveWindow(pw.headersInput, padding+halfWidth+padding, y, halfWidth, paramsHeight, true)

	// === Body Section ===
	y += paramsHeight + padding
	y += labelHeight + 3
	win32.MoveWindow(pw.bodyInput, padding, y, availableWidth, bodyHeight, true)

	// === Response Section ===
	y += bodyHeight + padding
	y += labelHeight + 5
	win32.MoveWindow(pw.responseOutput, padding, y, availableWidth, responseHeight, true)
}

func (pw *ProjectWindow) handleCommand(id int) {
	switch id {
	case ID_SEND_BTN:
		// Get the bound request from the current tab
		request := pw.getBoundRequest()
		if request == nil {
			win32.SetWindowText(pw.statusLabel, "❌ No request")
			win32.SetWindowText(pw.responseOutput, "Error: No request bound to this tab")
			return
		}

		// Sync UI values to the bound request before sending
		pw.syncUIToRequest(request)

		pw.saveCertificateConfig()
		win32.SetWindowText(pw.statusLabel, "⏳ Sending...")
		win32.SetWindowText(pw.responseOutput, "")
		go sendRequest(request, func(response string, err error) {
			if err != nil {
				win32.SetWindowText(pw.statusLabel, "❌ Error")
				win32.SetWindowText(pw.responseOutput, fmt.Sprintf("Error sending request:\r\n%v", err))
				return
			}
			win32.SetWindowText(pw.statusLabel, "✅ Success")
			win32.SetWindowText(pw.responseOutput, response)
		})
	case ID_NEW_BTN:
		pw.newProject()
	case ID_OPEN_BTN:
		pw.openProject()
	case ID_SAVE_BTN:
		pw.saveProject()
	case ID_CERT_BTN:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(pw.certInput, path)
		}
	case ID_KEY_BTN:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(pw.keyInput, path)
		}
	case ID_CA_BTN:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(pw.caInput, path)
		}
	case ID_OPEN_REQ_BTN:
		pw.openSelectedRequest()
	case ID_DELETE_REQ_BTN:
		pw.deleteSelectedRequest()
	case ID_RECENT_LIST:
		pw.openRecentProject()
	}
}

// showContextMenu displays the main context menu
func (pw *ProjectWindow) showContextMenu() {
	menu := win32.CreatePopupMenu()
	if menu == nil {
		return
	}
	defer menu.Destroy()

	menu.AddItem(MENU_PROJECT, "📁 Project View")
	menu.AddItem(MENU_NEW_REQUEST, "➕ New Request")
	menu.AddSeparator()
	menu.AddItem(MENU_SETTINGS, "⚙ Settings")
	menu.AddSeparator()
	menu.AddItem(MENU_ABOUT, "About REST Tester")

	// Show menu at cursor position (since menu button is in tab bar)
	selected := menu.Show(pw.mainWindow.Hwnd)

	switch selected {
	case MENU_SETTINGS:
		pw.CreateSettingsTab()
	case MENU_PROJECT:
		pw.CreateProjectViewTab()
	case MENU_NEW_REQUEST:
		pw.addNewRequestTab()
	case MENU_ABOUT:
		win32.MessageBox(pw.mainWindow.Hwnd, "REST Tester v1.0\nA modern REST API testing tool", "About", win32.MB_OK)
	}
}

// openRecentProject opens a project from the recent list
func (pw *ProjectWindow) openRecentProject() {
	idx := win32.ListBoxGetCurSel(pw.recentListBox)
	if idx < 0 || idx >= len(globalSettings.RecentProjects) {
		return
	}
	pw.openProjectFromPath(globalSettings.RecentProjects[idx])
}

// addNewRequestTab creates a new request tab
func (pw *ProjectWindow) addNewRequestTab() {
	// First ensure we have a project
	if pw.currentProject == nil {
		pw.currentProject = NewProject("Untitled Project")
	}

	// Create a new request
	req := NewRequest(fmt.Sprintf("Request %d", len(pw.currentProject.Requests)+1))
	req.Headers["Content-Type"] = "application/json"
	req.Headers["Accept"] = "application/json"
	pw.currentProject.AddRequest(req)

	// Create tab bound to the request
	pw.CreateRequestTab(req)
}

// openSelectedRequest opens the selected request from project list in a new tab
func (pw *ProjectWindow) openSelectedRequest() {
	idx := pw.getSelectedRequestIndex()
	if idx < 0 || idx >= len(pw.currentProject.Requests) {
		return
	}

	req := pw.currentProject.Requests[idx]

	// Open in new tab bound to the request
	pw.CreateRequestTab(req)
}

// deleteSelectedRequest removes the selected request from project
func (pw *ProjectWindow) deleteSelectedRequest() {
	idx := pw.getSelectedRequestIndex()
	if idx < 0 || idx >= len(pw.currentProject.Requests) {
		return
	}

	req := pw.currentProject.Requests[idx]

	// Remove from project
	pw.currentProject.RemoveRequest(req.ID)
	pw.updateProjectList()
}

// newProject creates a new empty project
func (pw *ProjectWindow) newProject() {
	pw.currentProject = NewProject("Untitled Project")
	// Open the project view tab
	pw.CreateProjectViewTab()
}

// openProject opens a project from file dialog
func (pw *ProjectWindow) openProject() {
	filePath, ok := win32.OpenFileDialog(
		pw.mainWindow.Hwnd,
		"Open Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
	)
	if !ok {
		return
	}
	pw.openProjectFromPath(filePath)
}

// openProjectFromPath opens a project from a specific file path
func (pw *ProjectWindow) openProjectFromPath(filePath string) {
	project, err := LoadProject(filePath)
	if err != nil {
		win32.MessageBox(pw.mainWindow.Hwnd, fmt.Sprintf("Error loading project: %v", err), "Error", win32.MB_OK)
		return
	}

	// Add to recent projects
	globalSettings.addRecentProject(filePath)

	pw.currentProject = project

	// Open the project view tab
	pw.CreateProjectViewTab()
	pw.loadCertificateUI()
}

// saveProject saves the current project to file
func (pw *ProjectWindow) saveProject() {
	pw.saveCertificateConfig()

	defaultName := pw.currentProject.Name + ".rtp"
	filePath, ok := win32.SaveFileDialog(
		pw.mainWindow.Hwnd,
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := pw.currentProject.Save(filePath); err != nil {
		win32.MessageBox(pw.mainWindow.Hwnd, fmt.Sprintf("Error saving project: %v", err), "Error", win32.MB_OK)
		return
	}

	// Update project name from filename
	name := filePath
	if idx := strings.LastIndex(name, "\\"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".rtp")
	pw.currentProject.Name = name
}

// saveCertificateConfig saves the certificate settings from UI to project
func (pw *ProjectWindow) saveCertificateConfig() {
	if globalSettings.Certificate == nil {
		globalSettings.Certificate = &CertificateConfig{}
	}
	globalSettings.Certificate.CertFile = win32.GetWindowText(pw.certInput)
	globalSettings.Certificate.KeyFile = win32.GetWindowText(pw.keyInput)
	globalSettings.Certificate.CACertFile = win32.GetWindowText(pw.caInput)
	globalSettings.Certificate.SkipVerify = win32.CheckboxIsChecked(pw.skipVerifyChk)
}

// loadCertificateUI loads certificate settings into the UI
func (pw *ProjectWindow) loadCertificateUI() {
	if globalSettings.Certificate == nil {
		pw.clearCertificateUI()
		return
	}
	win32.SetWindowText(pw.certInput, globalSettings.Certificate.CertFile)
	win32.SetWindowText(pw.keyInput, globalSettings.Certificate.KeyFile)
	win32.SetWindowText(pw.caInput, globalSettings.Certificate.CACertFile)
	win32.CheckboxSetChecked(pw.skipVerifyChk, globalSettings.Certificate.SkipVerify)
}

// clearCertificateUI clears all certificate input fields
func (pw *ProjectWindow) clearCertificateUI() {
	win32.SetWindowText(pw.certInput, "")
	win32.SetWindowText(pw.keyInput, "")
	win32.SetWindowText(pw.caInput, "")
	win32.CheckboxSetChecked(pw.skipVerifyChk, false)
}

// CreateRequestTab creates a new request tab bound to a Request object
func (pw *ProjectWindow) CreateRequestTab(req *Request) int {
	// Build headers string from map
	var headersText string
	for name, value := range req.Headers {
		if headersText != "" {
			headersText += "\r\n"
		}
		headersText += name + ": " + value
	}

	// Build query params string from map
	var queryText string
	for name, value := range req.QueryParams {
		if queryText != "" {
			queryText += "\r\n"
		}
		queryText += name + "=" + value
	}

	content := &RequestTabContent{
		BoundRequest: req, // Direct binding to the Request object
		Method:       req.Method,
		URL:          req.URL,
		Headers:      headersText,
		QueryParams:  queryText,
		Body:         req.Body,
		Status:       "Ready",
	}
	name := req.Method + " " + req.Name
	tabID := pw.tabs.AddTab(name, content)
	pw.tabs.SetActiveTab(tabID)
	return tabID
}

// CreateProjectViewTab creates the project structure view tab
func (pw *ProjectWindow) CreateProjectViewTab() int {
	content := &ProjectViewTabContent{
		BoundProject:  pw.currentProject, // Bind to current project
		SelectedIndex: -1,
	}
	tabID := pw.tabs.AddTab("📁 Project", content)
	pw.tabs.SetActiveTab(tabID)
	pw.panels.showProjectViewPanel()
	pw.updateProjectList()
	return tabID
}

// CreateSettingsTab creates the global settings tab
func (pw *ProjectWindow) CreateSettingsTab() int {
	content := &SettingsTabContent{
		SkipVerify: globalSettings.Certificate != nil && globalSettings.Certificate.SkipVerify,
	}
	tabID := pw.tabs.AddTab("⚙ Settings", content)
	pw.tabs.SetActiveTab(tabID)
	pw.panels.showSettingsPanel()
	pw.loadCertificateUI()
	return tabID
}

// CreateNewTabTab creates the "New Tab" start tab
func (pw *ProjectWindow) CreateNewTabTab() int {
	content := &NewTabTabContent{
		SelectedRecentIndex: -1,
	}
	tabID := pw.tabs.AddTab("Welcome", content)
	pw.tabs.SetActiveTab(tabID)
	return tabID
}

// SaveCurrentTabState saves the current UI state to the active tab's TabContent
func (pw *ProjectWindow) SaveCurrentTabState() {
	activeTab := pw.tabs.GetActiveTab()
	if activeTab == nil {
		return
	}

	content, ok := activeTab.Data.(TabContent)
	if !ok || content == nil {
		return
	}

	// Save state based on tab type using type assertion
	switch c := content.(type) {
	case *RequestTabContent:
		// Capture current UI values
		c.Method = win32.ComboBoxGetText(pw.methodCombo)
		c.URL = win32.GetWindowText(pw.urlInput)
		c.Headers = win32.GetWindowText(pw.headersInput)
		c.QueryParams = win32.GetWindowText(pw.queryInput)
		c.Body = win32.GetWindowText(pw.bodyInput)
		c.Response = win32.GetWindowText(pw.responseOutput)
		c.Status = win32.GetWindowText(pw.statusLabel)

		// Also sync to the bound Request object for persistence
		if c.BoundRequest != nil {
			pw.syncUIToRequest(c.BoundRequest)
		}

	case *ProjectViewTabContent:
		// Capture selected request index
		c.SelectedIndex = win32.ListBoxGetCurSel(pw.projectListBox)

	case *SettingsTabContent:
		// Capture certificate settings from UI
		c.CertFile = win32.GetWindowText(pw.certInput)
		c.KeyFile = win32.GetWindowText(pw.keyInput)
		c.CACertFile = win32.GetWindowText(pw.caInput)
		c.SkipVerify = win32.CheckboxIsChecked(pw.skipVerifyChk)

	case *NewTabTabContent:
		// Capture selected recent project index
		c.SelectedRecentIndex = win32.ListBoxGetCurSel(pw.recentListBox)
	}
}

// RestoreTabState restores UI state from a tab's saved content
func (pw *ProjectWindow) RestoreTabState(content TabContent) {
	if content == nil {
		return
	}

	switch c := content.(type) {
	case *RequestTabContent:
		pw.panels.showRequestPanel()
		pw.restoreRequestState(c)
		pw.resizeRequestPanel() // Apply current window size to controls
	case *ProjectViewTabContent:
		pw.panels.showProjectViewPanel()
		pw.updateProjectList()
		pw.restoreProjectViewState(c)
	case *SettingsTabContent:
		pw.panels.showSettingsPanel()
		pw.loadCertificateUI()
		pw.restoreSettingsState(c)
	case *NewTabTabContent:
		pw.panels.showWelcomePanel()
		pw.updateRecentList()
	}
}

// restoreRequestState restores request tab state to UI
func (pw *ProjectWindow) restoreRequestState(content *RequestTabContent) {
	if content == nil {
		return
	}

	// Set method
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for i, m := range methods {
		if m == content.Method {
			win32.ComboBoxSetCurSel(pw.methodCombo, i)
			break
		}
	}

	win32.SetWindowText(pw.urlInput, content.URL)
	win32.SetWindowText(pw.headersInput, content.Headers)
	win32.SetWindowText(pw.queryInput, content.QueryParams)
	win32.SetWindowText(pw.bodyInput, content.Body)
	win32.SetWindowText(pw.responseOutput, content.Response)
	win32.SetWindowText(pw.statusLabel, content.Status)
}

// restoreProjectViewState restores project view tab state to UI
func (pw *ProjectWindow) restoreProjectViewState(content *ProjectViewTabContent) {
	if content == nil {
		return
	}

	// Restore listbox selection
	if content.SelectedIndex >= 0 {
		win32.ListBoxSetCurSel(pw.projectListBox, content.SelectedIndex)
	}
}

// restoreSettingsState restores settings tab state to UI
func (pw *ProjectWindow) restoreSettingsState(content *SettingsTabContent) {
	if content == nil {
		return
	}

	win32.SetWindowText(pw.certInput, content.CertFile)
	win32.SetWindowText(pw.keyInput, content.KeyFile)
	win32.SetWindowText(pw.caInput, content.CACertFile)
	win32.CheckboxSetChecked(pw.skipVerifyChk, content.SkipVerify)
}

// updateProjectList refreshes the project structure list
func (pw *ProjectWindow) updateProjectList() {
	if pw.projectListBox == 0 || pw.currentProject == nil {
		return
	}

	// Clear the listbox
	win32.ListBoxResetContent(pw.projectListBox)

	// Add all requests
	for _, req := range pw.currentProject.Requests {
		displayText := req.Method + " " + req.Name
		if req.URL != "" {
			shortURL := req.URL
			if len(shortURL) > 40 {
				shortURL = shortURL[:40] + "..."
			}
			displayText = req.Method + " " + shortURL
		}
		win32.ListBoxAddString(pw.projectListBox, displayText)
	}
}

// updateRecentList refreshes the recently used projects list
func (pw *ProjectWindow) updateRecentList() {
	if pw.recentListBox == 0 {
		return
	}
	win32.ListBoxResetContent(pw.recentListBox)
	for _, path := range globalSettings.RecentProjects {
		win32.ListBoxAddString(pw.recentListBox, path)
	}
}

// getSelectedRequestIndex returns the selected index in project list
func (pw *ProjectWindow) getSelectedRequestIndex() int {
	return win32.ListBoxGetCurSel(pw.projectListBox)
}

// getBoundRequest returns the Request bound to the current tab, or nil if not a request tab
func (pw *ProjectWindow) getBoundRequest() *Request {
	activeTab := pw.tabs.GetActiveTab()
	if activeTab == nil {
		return nil
	}

	content, ok := activeTab.Data.(*RequestTabContent)
	if !ok || content == nil {
		return nil
	}

	return content.BoundRequest
}

// syncUIToRequest copies the current UI field values to the bound Request object
func (pw *ProjectWindow) syncUIToRequest(req *Request) {
	if req == nil {
		return
	}

	// Get method from combo box
	req.Method = win32.ComboBoxGetText(pw.methodCombo)

	// Get URL
	req.URL = win32.GetWindowText(pw.urlInput)

	// Get body
	req.Body = win32.GetWindowText(pw.bodyInput)

	// Parse headers from text to map
	req.Headers = make(map[string]string)
	headersText := win32.GetWindowText(pw.headersInput)
	for _, line := range strings.Split(headersText, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Parse query params from text to map
	req.QueryParams = make(map[string]string)
	queryText := win32.GetWindowText(pw.queryInput)
	for _, line := range strings.Split(queryText, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			req.QueryParams[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else if len(parts) == 1 && parts[0] != "" {
			req.QueryParams[strings.TrimSpace(parts[0])] = ""
		}
	}
}
