package main

import (
	"fmt"

	"hoermi.com/rest-test/win32"
)

// ProjectWindow holds all UI state for the main application window
type ProjectWindow struct {
	mainWindow *win32.Window
	tabs       *win32.TabManager[any]

	// Current loaded project
	currentProject *Project
	// Global settings
	settings *Settings
}

func NewProjectWindow() *ProjectWindow {
	mainWindow := win32.NewWindow("REST Tester", 1100, 850)
	tabs := win32.NewTabManager[any](mainWindow)
	mainWindow.TabManager = tabs
	pw := &ProjectWindow{
		mainWindow: mainWindow,
		tabs:       tabs,
	}
	panels := tabs.GetPanels()
	panels.Add(PanelRequest, createRequestPanel(pw))
	panels.Add(PanelProjectView, createProjectViewPanel(pw))
	panels.Add(PanelSettings, createSettingsPanel(pw))
	panels.Add(PanelWelcome, createWelcomePanel(pw))
	return pw
}

// showContextMenu displays the main context menu
func (pw *ProjectWindow) showContextMenu() {
	menu := pw.mainWindow.CreatePopupMenu()
	if menu == nil {
		return
	}
	defer menu.Destroy()

	menuProject := 1001
	menuNewRequest := 1002
	menuSettings := 1003
	menuAbout := 1004

	menu.AddItem(menuProject, "📁 Project View")
	menu.AddItem(menuNewRequest, "➕ New Request")
	menu.AddSeparator()
	menu.AddItem(menuSettings, "⚙ Settings")
	menu.AddSeparator()
	menu.AddItem(menuAbout, "About REST Tester")

	// Show menu at cursor position (since menu button is in tab bar)
	selected := menu.Show()

	switch selected {
	case menuProject:
		pw.createProjectViewTab()
	case menuNewRequest:
		pw.addNewRequestTab()
	case menuSettings:
		pw.createSettingsTab()
	case menuAbout:
		pw.mainWindow.MessageBox("REST Tester v1.0\nA modern REST API testing tool", "About")
	}
}

// addNewRequestTab creates a new request tab
func (pw *ProjectWindow) addNewRequestTab() {
	// First ensure we have a project
	if pw.currentProject == nil {
		return
	}

	req := pw.currentProject.NewRequest()

	// Create tab bound to the request
	pw.createRequestTab(req)
}

// newProject creates a new empty project
func (pw *ProjectWindow) newProject() {
	pw.currentProject = NewProject("Untitled Project")
	// Open the project view tab
	pw.createProjectViewTab()
}

// openProject opens a project from file dialog
func (pw *ProjectWindow) openProject() {
	filePath, ok := pw.mainWindow.OpenFileDialog(
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
		pw.mainWindow.MessageBox(fmt.Sprintf("Error loading project: %v", err), "Error")
		return
	}

	// Add to recent projects
	pw.settings.addRecentProject(filePath)

	pw.currentProject = project

	// Open the project view tab
	pw.createProjectViewTab()
}

// createRequestTab creates a new request tab bound to a Request object
func (pw *ProjectWindow) createRequestTab(req *Request) {
	content := &RequestTabContent{
		BoundRequest: req,
		BoundProject: pw.currentProject,
		Settings:     pw.settings,
		Status:       "Ready",
	}
	name := req.Method + " " + req.Name
	pw.tabs.AddTab(name, content, PanelRequest)
}

// createProjectViewTab creates the project structure view tab
// If a project view tab already exists, it will be focused instead
func (pw *ProjectWindow) createProjectViewTab() {
	// Check if a project view tab already exists
	if existingTabIndex, ok := pw.tabs.FindTabByPanelGroup(PanelProjectView); ok {
		// Tab exists, just focus it and update its data
		tab := pw.tabs.SetActiveTab(existingTabIndex)
		// Update the bound project in case it changed
		if content, ok := tab.Data.(*ProjectViewTabContent); ok {
			content.BoundProject = pw.currentProject
			content.SelectedIndex = -1
		}
		return
	}

	// Create new tab
	content := &ProjectViewTabContent{
		BoundProject:  pw.currentProject, // Bind to current project
		SelectedIndex: -1,
	}
	pw.tabs.AddTab("📁 Project", content, PanelProjectView)
}

// createSettingsTab creates the global settings tab
// If a settings tab already exists, it will be focused instead
func (pw *ProjectWindow) createSettingsTab() {
	// Check if a settings tab already exists
	if existingTabIndex, ok := pw.tabs.FindTabByPanelGroup(PanelSettings); ok {
		// Tab exists, just focus it
		pw.tabs.SetActiveTab(existingTabIndex)
		return
	}

	// Create new tab
	content := &SettingsTabContent{
		Settings: pw.settings,
	}
	pw.tabs.AddTab("⚙ Settings", content, PanelSettings)
}

// createWelcomeTab creates the "New Tab" start tab
func (pw *ProjectWindow) createWelcomeTab() {
	content := &WelcomeTabContent{
		RecentProjects:      pw.settings.RecentProjects,
		SelectedRecentIndex: -1,
	}
	pw.tabs.AddTab("Welcome", content, PanelWelcome)
}
