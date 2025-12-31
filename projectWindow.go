package main

import (
	"fmt"

	"hoermi.com/rest-test/win32"
)

// Layout constants for consistent UI spacing
const (
	// Common spacing
	layoutPadding = int32(12)

	// Standard control heights
	layoutLabelHeight     = int32(20)
	layoutInputHeight     = int32(26)
	layoutIconInputHeight = int32(32)
	layoutButtonWidth     = int32(120)

	// Default window size
	defaultWindowWidth  = int32(1100)
	defaultWindowHeight = int32(850)
)

// ProjectWindow holds all UI state for the main application window
type ProjectWindow struct {
	mainWindow *win32.Window
	tabs       *win32.TabManager[TabContent]

	panels Panels

	// Current loaded project
	currentProject *Project
	// Global settings
	settings *Settings
}

func NewProjectWindow() *ProjectWindow {
	mainWindow := win32.NewWindow("REST Tester", defaultWindowWidth, defaultWindowHeight)
	tabs := win32.NewTabManager[TabContent](mainWindow.Hwnd)
	mainWindow.TabManager = tabs
	pw := &ProjectWindow{
		mainWindow: mainWindow,
		tabs:       tabs,
	}
	return pw
}

// handleResize is called when the window is resized
func (pw *ProjectWindow) handleResize(width, height int32) {
	if pw.tabs != nil {
		pw.panels.get(PanelRequest).Resize(pw.tabs.GetHeight(), width, height)
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
		pw.MessageBox("REST Tester v1.0\nA modern REST API testing tool", "About")
	}
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

// newProject creates a new empty project
func (pw *ProjectWindow) newProject() {
	pw.currentProject = NewProject("Untitled Project")
	// Open the project view tab
	pw.CreateProjectViewTab()
}

func (pw *ProjectWindow) MessageBox(title, message string) {
	win32.MessageBox(pw.mainWindow.Hwnd, message, title, win32.MB_OK)
}

func (pw *ProjectWindow) SaveFileDialog(title, filter, defaultExt, defaultName string) (string, bool) {
	return win32.SaveFileDialog(
		pw.mainWindow.Hwnd,
		title,
		filter,
		defaultExt,
		defaultName,
	)
}

// OpenFileDialog shows a file open dialog and returns the selected file path
func (pw *ProjectWindow) OpenFileDialog(title, filter, defaultExt string) (string, bool) {
	return win32.OpenFileDialog(
		pw.mainWindow.Hwnd,
		title,
		filter,
		defaultExt,
	)
}

// openProject opens a project from file dialog
func (pw *ProjectWindow) openProject() {
	filePath, ok := pw.OpenFileDialog(
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
		pw.MessageBox(fmt.Sprintf("Error loading project: %v", err), "Error")
		return
	}

	// Add to recent projects
	pw.settings.addRecentProject(filePath)

	pw.currentProject = project

	// Open the project view tab
	pw.CreateProjectViewTab()
}

// CreateRequestTab creates a new request tab bound to a Request object
func (pw *ProjectWindow) CreateRequestTab(req *Request) {
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
		Settings:     pw.settings,
		Status:       "Ready",
	}
	name := req.Method + " " + req.Name
	pw.tabs.AddTab(name, content)
}

// CreateProjectViewTab creates the project structure view tab
func (pw *ProjectWindow) CreateProjectViewTab() {
	content := &ProjectViewTabContent{
		BoundProject:  pw.currentProject, // Bind to current project
		SelectedIndex: -1,
	}
	pw.tabs.AddTab("📁 Project", content)
	pw.panels.get(PanelProjectView).SetState(content)
	pw.panels.show(PanelProjectView)
}

// CreateSettingsTab creates the global settings tab
func (pw *ProjectWindow) CreateSettingsTab() {
	content := &SettingsTabContent{
		Certificate: pw.settings.Certificate,
	}
	pw.tabs.AddTab("⚙ Settings", content)
	pw.panels.get(PanelSettings).SetState(content)
	pw.panels.show(PanelSettings)
}

// CreateWelcomeTab creates the "New Tab" start tab
func (pw *ProjectWindow) CreateWelcomeTab() {
	content := &WelcomeTabContent{
		RecentProjects:      pw.settings.RecentProjects,
		SelectedRecentIndex: -1,
	}
	pw.tabs.AddTab("Welcome", content)
	pw.panels.get(PanelWelcome).SetState(content)
	pw.panels.show(PanelWelcome)
}

// SaveCurrentTabState saves the current UI state to the active tab's TabContent
func (pw *ProjectWindow) SaveCurrentTabState() {
	activeTab := pw.tabs.GetActiveTab()
	if activeTab == nil || activeTab.Data == nil {
		return
	}

	// Save state based on tab type using type assertion
	switch activeTab.Data.(type) {
	case *RequestTabContent:
		pw.panels.get(PanelRequest).SaveState()
	case *ProjectViewTabContent:
		pw.panels.get(PanelProjectView).SaveState()
	case *SettingsTabContent:
		pw.panels.get(PanelSettings).SaveState()
	case *WelcomeTabContent:
		pw.panels.get(PanelWelcome).SaveState()
	}
}

// RestoreTabState restores UI state from a tab's saved content
func (pw *ProjectWindow) RestoreTabState(content TabContent) { //TODO: Move to TabManager?
	if content == nil {
		return
	}

	switch c := content.(type) {
	case *RequestTabContent:
		pw.panels.show(PanelRequest)
		pw.panels.get(PanelRequest).SetState(c)
	case *ProjectViewTabContent:
		pw.panels.show(PanelProjectView)
		pw.panels.get(PanelProjectView).SetState(c)
	case *SettingsTabContent:
		pw.panels.show(PanelSettings)
		pw.panels.get(PanelSettings).SetState(c)
	case *WelcomeTabContent:
		pw.panels.show(PanelWelcome)
		pw.panels.get(PanelWelcome).SetState(c)
	}
}
