package main

import (
	"fmt"
	"path/filepath"

	"hoermi.com/rest-test/win32"
)

// ProjectWindow holds all UI state for the main application window
type ProjectWindow struct {
	mainWindow win32.ControlFactory
	tabs       *win32.TabManager[any]

	// Current loaded project
	currentProject *Project
	// Global settings
	settings *Settings
}

type TabController interface {
	addNewTab()
	createSettingsTab()
	createWelcomeTab()
	createProjectViewTab()
	createRequestTab(req *Request, path string)
	createPendingRequestTab(req *Request, path string)
	refreshProjectViewTab()
}

type ProjectManager interface {
	newProject()
	openProject()
	openProjectFromPath(filePath string)
	saveProject()
	newRequest()
}

func NewProjectWindow() *ProjectWindow {
	mainWindow := win32.NewWindow("REST Tester", 1100, 850)
	tabs := win32.NewTabManager[any](mainWindow)
	mainWindow.TabDrawer = tabs
	pw := &ProjectWindow{
		mainWindow: mainWindow,
		tabs:       tabs,
	}
	panels := tabs.GetPanels()
	panels.Add(PanelRequest, createRequestPanel(pw.mainWindow, pw))
	panels.Add(PanelProjectView, createProjectViewPanel(pw.mainWindow, pw, pw))
	panels.Add(PanelSettings, createSettingsPanel(pw.mainWindow))
	panels.Add(PanelWelcome, createWelcomePanel(pw.mainWindow, pw))

	settings, err := InitSettings()
	if err != nil {
		mainWindow.MessageBox("Error", "Error initializing settings: "+err.Error())
		return nil
	}
	pw.settings = settings

	// Handle button clicks
	mainWindow.OnCommand = tabs.GetPanels().HandleCommand
	// Handle window resizing
	mainWindow.OnResize = tabs.GetPanels().Resize
	tabs.OnTabClosed = func() {
		// If no tabs left, show welcome tab
		if tabs.GetTabCount() == 0 {
			pw.createWelcomeTab()
		}
	}

	// Wire up the tab manager's menu button callback
	tabs.OnMenuClick = pw.showContextMenu
	tabs.OnNewTab = pw.addNewTab

	// Start with the Welcome Tab
	pw.createWelcomeTab()

	return pw
}

// showContextMenu displays the main context menu
func (pw *ProjectWindow) showContextMenu() {
	menu := pw.mainWindow.CreatePopupMenu()
	if menu == nil {
		return
	}
	defer menu.Destroy()

	menuIDNewTab := 1000
	menuIDSettings := 1001
	menuIDAbout := 1002

	if pw.currentProject == nil {
		menu.AddItem(menuIDNewTab, "➕ New Project")
	} else {
		menu.AddItem(menuIDNewTab, "➕ New Request")
	}
	menu.AddSeparator()
	menu.AddItem(menuIDSettings, "⚙ Settings")
	menu.AddSeparator()
	menu.AddItem(menuIDAbout, "About REST Tester")

	// Show menu at cursor position (since menu button is in tab bar)
	selected := menu.Show()

	switch selected {
	case menuIDNewTab:
		pw.addNewTab()
	case menuIDSettings:
		pw.createSettingsTab()
	case menuIDAbout:
		pw.mainWindow.MessageBox("About", `REST Tester v1.0
A REST API testing tool
Developed by Thomas Hörmann
Published under Apache License 2.0
https://github.com/thomashoermann407-sudo/rest-tester`)
	}
}

func (pw *ProjectWindow) addNewTab() {
	if pw.currentProject == nil {
		pw.newProject()
	} else {
		pw.newRequest()
	}
}

func (pw *ProjectWindow) newRequest() {
	req := pw.currentProject.NewRequest()
	pw.createPendingRequestTab(req, "/")
}

func (pw *ProjectWindow) newProject() {
	pw.currentProject = NewProject("Untitled Project")
	// Open the project view tab
	pw.createProjectViewTab()
}

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

func (pw *ProjectWindow) saveProject() {

	defaultName := pw.currentProject.Name
	filePath, ok := pw.mainWindow.SaveFileDialog(
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := pw.currentProject.Save(filePath); err != nil {
		pw.mainWindow.MessageBox("Error", fmt.Sprintf("Error saving project: %v", err))
		return
	}

	// Update project name from filename
	pw.currentProject.Name = filepath.Base(filePath)
}

// openProjectFromPath opens a project from a specific file path
func (pw *ProjectWindow) openProjectFromPath(filePath string) {
	project, err := LoadProject(filePath)
	if err != nil {
		pw.mainWindow.MessageBox("Error", fmt.Sprintf("Error loading project: %v", err))
		return
	}

	// Add to recent projects
	if err := pw.settings.addRecentProject(filePath); err != nil {
		pw.mainWindow.MessageBox("Warning", fmt.Sprintf("Error adding recent project: %v", err))
	}

	pw.currentProject = project

	// Open the project view tab
	pw.createProjectViewTab()
}

// createRequestTab creates a new request tab bound to a Request object
func (pw *ProjectWindow) createRequestTab(req *Request, path string) {
	pw.createRequestTabInternal(req, path, false)
}

func (pw *ProjectWindow) createPendingRequestTab(req *Request, path string) {
	pw.createRequestTabInternal(req, path, true)
}

// createRequestTabInternal creates a request tab with optional pending state
func (pw *ProjectWindow) createRequestTabInternal(req *Request, path string, pending bool) {
	content := &RequestTabContent{
		BoundRequest: req,
		BoundProject: pw.currentProject,
		Path:         path,
		Settings:     pw.settings,
		Pending:      pending,
	}

	name := "New Request"
	if !pending {
		name = req.Method + " " + req.Name
	}
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

// refreshProjectViewTab refreshes the project view tab if it exists
func (pw *ProjectWindow) refreshProjectViewTab() {
	// Find the project view tab
	if existingTabIndex, ok := pw.tabs.FindTabByPanelGroup(PanelProjectView); ok {
		// SetActiveTab automatically calls SetState, which refreshes the tree
		pw.tabs.SetActiveTab(existingTabIndex)
	}
}
