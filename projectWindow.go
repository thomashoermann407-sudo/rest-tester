package main

import (
	"fmt"
	"strings"

	"hoermi.com/rest-test/win32"
)

// Layout constants for consistent UI spacing
const (
	// Common spacing
	layoutPadding   = int32(12)
	layoutGapSmall  = int32(5)
	layoutGapMedium = int32(10)
	layoutGapLarge  = int32(30)

	// Standard control heights
	layoutLabelHeight = int32(20)
	layoutInputHeight = int32(26)
	layoutButtonWidth = int32(120)

	// Welcome panel uses larger controls
	layoutWelcomeLabelHeight = int32(24)
	layoutWelcomeInputHeight = int32(32)
	layoutWelcomeButtonWidth = int32(150)

	// Default window size
	defaultWindowWidth  = int32(1100)
	defaultWindowHeight = int32(850)
)

// ProjectWindow holds all UI state for the main application window
type ProjectWindow struct {
	mainWindow     *win32.Window
	tabs           *win32.TabManager
	methodCombo    *win32.ClickControl
	urlInput       *win32.Control
	headersInput   *win32.Control
	queryInput     *win32.Control
	bodyInput      *win32.Control
	responseOutput *win32.Control
	statusLabel    *win32.Control
	sendBtn        *win32.ClickControl
	// Labels for request panel
	methodLabel   *win32.Control
	urlLabel      *win32.Control
	headersLabel  *win32.Control
	queryLabel    *win32.Control
	bodyLabel     *win32.Control
	responseLabel *win32.Control

	// Project View Panel controls
	projectListBox *win32.ClickControl
	openReqBtn     *win32.ClickControl
	deleteReqBtn   *win32.ClickControl
	projectInfo    *win32.Control
	saveBtn        *win32.ClickControl

	// New Tab Panel controls
	newTabTitle   *win32.Control
	newTabNewBtn  *win32.ClickControl
	newTabOpenBtn *win32.ClickControl
	recentLabel   *win32.Control
	recentListBox *win32.ClickControl

	// Settings Panel controls
	certInput       *win32.Control
	keyInput        *win32.Control
	caInput         *win32.Control
	skipVerifyChk   *win32.ClickControl
	certBtn         *win32.ClickControl
	keyBtn          *win32.ClickControl
	caBtn           *win32.ClickControl
	certLabel       *win32.Control
	keyLabel        *win32.Control
	caLabel         *win32.Control
	settingsTitle   *win32.Control
	saveSettingsBtn *win32.ClickControl

	// Layout state for resizing
	windowWidth       int32
	windowHeight      int32
	paramsHeightRatio float32
	bodyHeightRatio   float32

	panels Panels

	// Current loaded project
	currentProject *Project
	// Global settings
	settings *Settings
}

func NewProjectWindow() *ProjectWindow {
	mainWindow := win32.NewWindow("REST Tester", defaultWindowWidth, defaultWindowHeight)
	tabs := mainWindow.EnableTabs()

	pw := &ProjectWindow{
		mainWindow:        mainWindow,
		tabs:              tabs,
		windowWidth:       defaultWindowWidth,
		windowHeight:      defaultWindowHeight,
		paramsHeightRatio: 0.10,
		bodyHeightRatio:   0.15,
	}
	return pw
}

func (pw *ProjectWindow) createRequestPanel() {
	y := layoutPadding

	// === Request Row ===
	// Note: Actual positions are recalculated by resizeRequestPanel()
	methodLabelWidth := int32(50)
	methodComboWidth := int32(90)
	urlLabelWidth := int32(30)
	sendBtnWidth := int32(90)

	pw.methodLabel = pw.mainWindow.CreateLabel("Method", layoutPadding, y+3, methodLabelWidth, layoutLabelHeight)
	pw.methodCombo = pw.mainWindow.CreateComboBox(layoutPadding+methodLabelWidth+layoutGapSmall, y, methodComboWidth, 200)
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		pw.methodCombo.ComboBoxAddString(method)
	}
	pw.methodCombo.ComboBoxSetCurSel(0)

	urlX := layoutPadding + methodLabelWidth + layoutGapSmall + methodComboWidth + layoutPadding
	pw.urlLabel = pw.mainWindow.CreateLabel("URL", urlX, y+3, urlLabelWidth, layoutLabelHeight)
	pw.urlInput = pw.mainWindow.CreateInput(urlX+urlLabelWidth+layoutGapSmall, y, 680, layoutInputHeight)
	pw.sendBtn = pw.mainWindow.CreateButton("Send", defaultWindowWidth-layoutPadding-sendBtnWidth, y, sendBtnWidth, layoutInputHeight)

	// === Query Parameters & Headers Section ===
	y += layoutInputHeight + layoutPadding
	pw.queryLabel = pw.mainWindow.CreateLabel("Query Parameters (one per line: key=value)", layoutPadding, y, 300, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.queryInput = pw.mainWindow.CreateCodeEdit(layoutPadding, y, 520, 70, false)

	halfWidth := (defaultWindowWidth - layoutPadding*3) / 2
	pw.headersLabel = pw.mainWindow.CreateLabel("Headers (one per line: Header: value)", layoutPadding+halfWidth+layoutPadding, y-layoutLabelHeight-layoutGapSmall, 300, layoutLabelHeight)
	pw.headersInput = pw.mainWindow.CreateCodeEdit(layoutPadding+halfWidth+layoutPadding, y, 520, 70, false)

	// === Body Section ===
	y += 70 + layoutPadding
	pw.bodyLabel = pw.mainWindow.CreateLabel("Request Body", layoutPadding, y, 150, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.bodyInput = pw.mainWindow.CreateCodeEdit(layoutPadding, y, defaultWindowWidth-layoutPadding*2, 120, false)

	// === Response Section ===
	y += 120 + layoutPadding
	pw.responseLabel = pw.mainWindow.CreateLabel("Response", layoutPadding, y, 80, layoutLabelHeight)
	pw.statusLabel = pw.mainWindow.CreateLabel("Ready", layoutPadding+90, y, 400, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.responseOutput = pw.mainWindow.CreateCodeEdit(layoutPadding, y, defaultWindowWidth-layoutPadding*2, 350, true)
}

func (pw *ProjectWindow) createProjectViewPanel() {
	y := layoutPadding * 4 // Below tab bar area
	listWidth := int32(500)
	listHeight := int32(500)

	pw.projectInfo = pw.mainWindow.CreateLabel("Double-click a request to open it in a new tab", layoutPadding, y, listWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding

	pw.projectListBox = pw.mainWindow.CreateListBox(layoutPadding, y, listWidth, listHeight)

	// Action buttons to the right of the list
	btnX := layoutPadding + listWidth + layoutPadding
	pw.openReqBtn = pw.mainWindow.CreateButton("Open in Tab", btnX, y, layoutButtonWidth, layoutInputHeight)
	pw.deleteReqBtn = pw.mainWindow.CreateButton("Delete", btnX, y+layoutInputHeight+layoutGapSmall, layoutButtonWidth, layoutInputHeight)
	pw.saveBtn = pw.mainWindow.CreateButton("Save Project", btnX, y+layoutInputHeight*2+layoutPadding, layoutButtonWidth, layoutInputHeight)
	// Visibility is managed by panels.go initPanels()
}

func (pw *ProjectWindow) createSettingsPanel() {
	y := layoutPadding * 4 // Below tab bar area
	inputWidth := int32(400)
	browseBtnWidth := int32(30)

	pw.settingsTitle = pw.mainWindow.CreateLabel("Global Settings", layoutPadding, y, 200, layoutLabelHeight+4)
	y += layoutLabelHeight + layoutPadding

	pw.certLabel = pw.mainWindow.CreateLabel("Client Certificate (PEM)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.certInput = pw.mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	pw.certBtn = pw.mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutGapSmall, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	pw.keyLabel = pw.mainWindow.CreateLabel("Private Key (PEM)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.keyInput = pw.mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	pw.keyBtn = pw.mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutGapSmall, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	pw.caLabel = pw.mainWindow.CreateLabel("CA Bundle (optional)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutGapSmall
	pw.caInput = pw.mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	pw.caBtn = pw.mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutGapSmall, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	pw.skipVerifyChk = pw.mainWindow.CreateCheckbox("Skip TLS Verification (insecure)", layoutPadding, y, 280, layoutInputHeight)
	y += layoutInputHeight + layoutPadding
	pw.saveSettingsBtn = pw.mainWindow.CreateButton("Save Settings", layoutPadding, y, layoutButtonWidth, layoutInputHeight)
}

func (pw *ProjectWindow) createNewTabPanel() {
	y := layoutPadding * 6 // More space at top for welcome area
	recentListWidth := int32(400)
	recentListHeight := int32(300)

	pw.newTabTitle = pw.mainWindow.CreateLabel("REST Tester - Start", layoutPadding, y, 400, layoutWelcomeLabelHeight+layoutGapMedium)
	y += layoutWelcomeLabelHeight + layoutGapLarge

	pw.newTabNewBtn = pw.mainWindow.CreateButton("📄 New Project", layoutPadding, y, layoutWelcomeButtonWidth, layoutWelcomeInputHeight)
	pw.newTabOpenBtn = pw.mainWindow.CreateButton("📂 Open Project", layoutPadding+layoutWelcomeButtonWidth+layoutGapMedium, y, layoutWelcomeButtonWidth, layoutWelcomeInputHeight)
	y += layoutWelcomeInputHeight + layoutGapLarge

	pw.recentLabel = pw.mainWindow.CreateLabel("Recent Projects:", layoutPadding, y, 200, layoutWelcomeLabelHeight)
	y += layoutWelcomeLabelHeight + layoutGapSmall
	pw.recentListBox = pw.mainWindow.CreateListBox(layoutPadding, y, recentListWidth, recentListHeight)
}

// handleResize is called when the window is resized
func (pw *ProjectWindow) handleResize(width, height int32) {
	pw.windowWidth = width
	pw.windowHeight = height
	pw.tabs.SetWidth(width)

	if pw.tabs != nil {
		pw.resizeRequestPanel()
	}
}

// resizeRequestPanel adjusts all controls in the request panel based on window size
func (pw *ProjectWindow) resizeRequestPanel() {
	tabHeight := pw.tabs.GetHeight()

	// Calculate available height (excluding tab bar and padding)
	availableHeight := pw.windowHeight - tabHeight - layoutPadding*5 - layoutLabelHeight*5 - layoutInputHeight

	// Calculate panel heights based on ratios
	minParamsHeight := int32(60)
	minBodyHeight := int32(80)
	minResponseHeight := int32(150)

	paramsHeight := max(int32(float32(availableHeight)*pw.paramsHeightRatio), minParamsHeight)
	bodyHeight := max(int32(float32(availableHeight)*pw.bodyHeightRatio), minBodyHeight)
	responseHeight := max(availableHeight-paramsHeight-bodyHeight, minResponseHeight)

	availableWidth := pw.windowWidth - layoutPadding*2

	y := tabHeight + layoutPadding

	// === Request Row (fixed height) ===
	methodLabelWidth := int32(50)
	methodComboWidth := int32(90)
	sendBtnWidth := int32(90)

	pw.methodCombo.MoveWindow(layoutPadding+methodLabelWidth+layoutGapSmall, y, methodComboWidth, 200, true)

	urlLabelWidth := int32(30)
	urlX := layoutPadding + methodLabelWidth + layoutGapSmall + methodComboWidth + layoutPadding + urlLabelWidth + layoutGapSmall
	urlWidth := availableWidth - methodLabelWidth - methodComboWidth - urlLabelWidth - sendBtnWidth - layoutPadding*2 - layoutGapSmall*2
	pw.urlInput.MoveWindow(urlX, y, urlWidth, layoutInputHeight, true)
	pw.sendBtn.MoveWindow(pw.windowWidth-layoutPadding-sendBtnWidth, y, sendBtnWidth, layoutInputHeight, true)

	// === Query Parameters & Headers Section ===
	y += layoutInputHeight + layoutPadding
	y += layoutLabelHeight + layoutGapSmall

	// Split width 50/50 for params and headers
	halfWidth := (availableWidth - layoutPadding) / 2

	pw.queryInput.MoveWindow(layoutPadding, y, halfWidth, paramsHeight, true)
	pw.headersInput.MoveWindow(layoutPadding+halfWidth+layoutPadding, y, halfWidth, paramsHeight, true)

	// === Body Section ===
	y += paramsHeight + layoutPadding
	y += layoutLabelHeight + layoutGapSmall
	pw.bodyInput.MoveWindow(layoutPadding, y, availableWidth, bodyHeight, true)

	// === Response Section ===
	y += bodyHeight + layoutPadding
	y += layoutLabelHeight + layoutGapSmall
	pw.responseOutput.MoveWindow(layoutPadding, y, availableWidth, responseHeight, true)
}

func (pw *ProjectWindow) handleCommand(id int, notifyCode int) {
	switch id {
	case pw.sendBtn.ID:
		// Get the bound request from the current tab
		request := pw.getBoundRequest()
		if request == nil {
			pw.statusLabel.SetText("❌ No request")
			pw.responseOutput.SetText("Error: No request bound to this tab")
			return
		}

		// Sync UI values to the bound request before sending
		pw.syncUIToRequest(request)

		pw.statusLabel.SetText("⏳ Sending...")
		pw.responseOutput.SetText("")
		go sendRequest(request, pw.settings, func(response string, err error) {
			if err != nil {
				pw.statusLabel.SetText("❌ Error")
				pw.responseOutput.SetText(fmt.Sprintf("Error sending request:\r\n%v", err))
				return
			}
			pw.statusLabel.SetText("✅ Success")
			pw.responseOutput.SetText(response)
		})
	case pw.newTabNewBtn.ID:
		pw.newProject()
	case pw.newTabOpenBtn.ID:
		pw.openProject()
	case pw.saveBtn.ID:
		pw.saveProject()
	case pw.certBtn.ID:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			pw.certInput.SetText(path)
		}
	case pw.keyBtn.ID:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			pw.keyInput.SetText(path)
		}
	case pw.caBtn.ID:
		if path, ok := win32.OpenFileDialog(pw.mainWindow.Hwnd, "Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			pw.caInput.SetText(path)
		}
	case pw.openReqBtn.ID:
		pw.openSelectedRequest()
	case pw.deleteReqBtn.ID:
		pw.deleteSelectedRequest()
	case pw.projectListBox.ID:
		// Open request on double-click
		if notifyCode == win32.LBN_DBLCLK {
			pw.openSelectedRequest()
		}
	case pw.recentListBox.ID:
		// Only open on double-click, not selection change
		if notifyCode == win32.LBN_DBLCLK {
			pw.openRecentProject()
		}
	case pw.saveSettingsBtn.ID:
		pw.saveCertificateConfig()
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
	idx := win32.ListBoxGetCurSel(pw.recentListBox.Hwnd)
	if idx < 0 || idx >= len(pw.settings.RecentProjects) {
		return
	}
	pw.openProjectFromPath(pw.settings.RecentProjects[idx])
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

	// Remove from project
	pw.currentProject.RemoveRequest(idx)
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
	pw.settings.addRecentProject(filePath)

	pw.currentProject = project

	// Open the project view tab
	pw.CreateProjectViewTab()
	pw.loadCertificateUI()
}

// saveProject saves the current project to file
func (pw *ProjectWindow) saveProject() {

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
	if pw.settings.Certificate == nil {
		pw.settings.Certificate = &CertificateConfig{}
	}
	pw.settings.Certificate.CertFile = pw.certInput.GetText()
	pw.settings.Certificate.KeyFile = pw.keyInput.GetText()
	pw.settings.Certificate.CACertFile = pw.caInput.GetText()
	pw.settings.Certificate.SkipVerify = win32.CheckboxIsChecked(pw.skipVerifyChk.Hwnd)
	pw.settings.save()
}

// loadCertificateUI loads certificate settings into the UI
func (pw *ProjectWindow) loadCertificateUI() {
	if pw.settings.Certificate == nil {
		pw.clearCertificateUI()
		return
	}
	pw.certInput.SetText(pw.settings.Certificate.CertFile)
	pw.keyInput.SetText(pw.settings.Certificate.KeyFile)
	pw.caInput.SetText(pw.settings.Certificate.CACertFile)
	win32.CheckboxSetChecked(pw.skipVerifyChk.Hwnd, pw.settings.Certificate.SkipVerify)
}

// clearCertificateUI clears all certificate input fields
func (pw *ProjectWindow) clearCertificateUI() {
	pw.certInput.SetText("")
	pw.keyInput.SetText("")
	pw.caInput.SetText("")
	win32.CheckboxSetChecked(pw.skipVerifyChk.Hwnd, false)
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
	pw.panels.show(PanelProjectView)
	pw.updateProjectList()
	return tabID
}

// CreateSettingsTab creates the global settings tab
func (pw *ProjectWindow) CreateSettingsTab() int {
	content := &SettingsTabContent{
		SkipVerify: pw.settings.Certificate != nil && pw.settings.Certificate.SkipVerify,
	}
	tabID := pw.tabs.AddTab("⚙ Settings", content)
	pw.tabs.SetActiveTab(tabID)
	pw.panels.show(PanelSettings)
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
		c.Method = win32.ComboBoxGetText(pw.methodCombo.Hwnd)
		c.URL = pw.urlInput.GetText()
		c.Headers = pw.headersInput.GetText()
		c.QueryParams = pw.queryInput.GetText()
		c.Body = pw.bodyInput.GetText()
		c.Response = pw.responseOutput.GetText()
		c.Status = pw.statusLabel.GetText()

		// Also sync to the bound Request object for persistence
		if c.BoundRequest != nil {
			pw.syncUIToRequest(c.BoundRequest)
		}

	case *ProjectViewTabContent:
		// Capture selected request index
		c.SelectedIndex = win32.ListBoxGetCurSel(pw.projectListBox.Hwnd)

	case *SettingsTabContent:
		// Capture certificate settings from UI
		c.CertFile = pw.certInput.GetText()
		c.KeyFile = pw.keyInput.GetText()
		c.CACertFile = pw.caInput.GetText()
		c.SkipVerify = win32.CheckboxIsChecked(pw.skipVerifyChk.Hwnd)

	case *NewTabTabContent:
		// Capture selected recent project index
		c.SelectedRecentIndex = win32.ListBoxGetCurSel(pw.recentListBox.Hwnd)
	}
}

// RestoreTabState restores UI state from a tab's saved content
func (pw *ProjectWindow) RestoreTabState(content TabContent) {
	if content == nil {
		return
	}

	switch c := content.(type) {
	case *RequestTabContent:
		pw.panels.show(PanelRequest)
		pw.restoreRequestState(c)
		pw.resizeRequestPanel() // Apply current window size to controls
	case *ProjectViewTabContent:
		pw.panels.show(PanelProjectView)
		pw.updateProjectList()
		pw.restoreProjectViewState(c)
	case *SettingsTabContent:
		pw.panels.show(PanelSettings)
		pw.loadCertificateUI()
		pw.restoreSettingsState(c)
	case *NewTabTabContent:
		pw.panels.show(PanelWelcome)
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
			pw.methodCombo.ComboBoxSetCurSel(i)
			break
		}
	}

	pw.urlInput.SetText(content.URL)
	pw.headersInput.SetText(content.Headers)
	pw.queryInput.SetText(content.QueryParams)
	pw.bodyInput.SetText(content.Body)
	pw.responseOutput.SetText(content.Response)
	pw.statusLabel.SetText(content.Status)
}

// restoreProjectViewState restores project view tab state to UI
func (pw *ProjectWindow) restoreProjectViewState(content *ProjectViewTabContent) {
	if content == nil {
		return
	}

	// Restore listbox selection
	if content.SelectedIndex >= 0 {
		win32.ListBoxSetCurSel(pw.projectListBox.Hwnd, content.SelectedIndex)
	}
}

// restoreSettingsState restores settings tab state to UI
func (pw *ProjectWindow) restoreSettingsState(content *SettingsTabContent) {
	if content == nil {
		return
	}

	pw.certInput.SetText(content.CertFile)
	pw.keyInput.SetText(content.KeyFile)
	pw.caInput.SetText(content.CACertFile)
	win32.CheckboxSetChecked(pw.skipVerifyChk.Hwnd, content.SkipVerify)
}

// updateProjectList refreshes the project structure list
func (pw *ProjectWindow) updateProjectList() {
	if pw.projectListBox == nil || pw.currentProject == nil {
		return
	}

	// Clear the listbox
	win32.ListBoxResetContent(pw.projectListBox.Hwnd)

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
		win32.ListBoxAddString(pw.projectListBox.Hwnd, displayText)
	}
}

// updateRecentList refreshes the recently used projects list
func (pw *ProjectWindow) updateRecentList() {
	if pw.recentListBox == nil {
		return
	}
	win32.ListBoxResetContent(pw.recentListBox.Hwnd)
	for _, path := range pw.settings.RecentProjects {
		// Display only the filename, not the full path
		displayName := path
		if idx := strings.LastIndex(path, "\\"); idx >= 0 {
			displayName = path[idx+1:]
		}
		win32.ListBoxAddString(pw.recentListBox.Hwnd, displayName)
	}
}

// getSelectedRequestIndex returns the selected index in project list
func (pw *ProjectWindow) getSelectedRequestIndex() int {
	return win32.ListBoxGetCurSel(pw.projectListBox.Hwnd)
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
	req.Method = win32.ComboBoxGetText(pw.methodCombo.Hwnd)

	// Get URL
	req.URL = pw.urlInput.GetText()

	// Get body
	req.Body = pw.bodyInput.GetText()

	// Parse headers from text to map
	req.Headers = make(map[string]string)
	headersText := pw.headersInput.GetText()
	for line := range strings.SplitSeq(headersText, "\n") {
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
	queryText := pw.queryInput.GetText()
	for line := range strings.SplitSeq(queryText, "\n") {
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
