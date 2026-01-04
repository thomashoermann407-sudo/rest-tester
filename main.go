package main

import (
	"runtime"
)

func main() {
	runtime.LockOSThread()

	pw := NewProjectWindow()
	settings, err := InitSettings()
	if err != nil {
		pw.mainWindow.MessageBox("Error initializing settings: "+err.Error(), "Error")
		return
	}
	pw.settings = settings
	tabs := pw.tabs
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

	// Handle button clicks
	pw.mainWindow.OnCommand = pw.tabs.GetPanels().HandleCommand
	// Handle window resizing
	pw.mainWindow.OnResize = pw.tabs.GetPanels().Resize

	pw.mainWindow.Run()
}
