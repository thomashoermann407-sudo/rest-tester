package main

import "hoermi.com/rest-test/win32"

// Panel visibility groups
type PanelGroup struct {
	controls []win32.HWND
	visible  bool
}

var (
	requestPanel     *PanelGroup
	projectViewPanel *PanelGroup
	settingsPanel    *PanelGroup
	newTabPanel      *PanelGroup
)

// initPanels initializes the panel groups (call after creating controls)
func initPanels() {
	requestPanel = &PanelGroup{
		controls: []win32.HWND{
			methodCombo, urlInput, headersInput, queryInput, bodyInput,
			responseOutput, statusLabel, sendBtn,
			methodLabel, urlLabel, headersLabel, queryLabel, bodyLabel, responseLabel,
		},
		visible: false,
	}

	projectViewPanel = &PanelGroup{
		controls: []win32.HWND{projectListBox, openReqBtn, deleteReqBtn, projectInfo, saveBtn},
		visible:  false,
	}

	settingsPanel = &PanelGroup{
		controls: []win32.HWND{
			certInput, keyInput, caInput, skipVerifyChk,
			certBtn, keyBtn, caBtn,
			certLabel, keyLabel, caLabel, settingsTitle,
		},
		visible: false,
	}

	newTabPanel = &PanelGroup{
		controls: []win32.HWND{
			newTabTitle, newTabNewBtn, newTabOpenBtn,
			recentLabel, recentListBox,
		},
		visible: true,
	}
}

// showPanel shows a panel and hides others
func showPanel(panel *PanelGroup) {
	// Hide all panels
	hidePanel(requestPanel)
	hidePanel(projectViewPanel)
	hidePanel(settingsPanel)
	hidePanel(newTabPanel)

	// Show the requested panel
	if panel != nil {
		panel.visible = true
		for _, hwnd := range panel.controls {
			if hwnd != 0 {
				win32.ShowWindow(hwnd, win32.SW_SHOW)
			}
		}
	}
}

// hidePanel hides a panel
func hidePanel(panel *PanelGroup) {
	if panel == nil {
		return
	}
	panel.visible = false
	for _, hwnd := range panel.controls {
		if hwnd != 0 {
			win32.ShowWindow(hwnd, win32.SW_HIDE)
		}
	}
}

// showRequestPanel shows the request editing panel
func showRequestPanel() {
	showPanel(requestPanel)
}

// showProjectViewPanel shows the project structure panel
func showProjectViewPanel() {
	showPanel(projectViewPanel)
	updateProjectList()
}

// showSettingsPanel shows the global settings panel
func showSettingsPanel() {
	showPanel(settingsPanel)
	loadCertificateUI()
}

// showWelcomePanel shows the welcome panel
func showWelcomePanel() {
	showPanel(newTabPanel)
	updateRecentList()
}

// updateProjectList refreshes the project structure list
func updateProjectList() {
	if projectListBox == 0 || currentProject == nil {
		return
	}

	// Clear the listbox
	win32.ListBoxResetContent(projectListBox)

	// Add all requests
	for _, req := range currentProject.Requests {
		displayText := req.Method + " " + req.Name
		if req.URL != "" {
			shortURL := req.URL
			if len(shortURL) > 40 {
				shortURL = shortURL[:40] + "..."
			}
			displayText = req.Method + " " + shortURL
		}
		win32.ListBoxAddString(projectListBox, displayText)
	}
}

// updateRecentList refreshes the recently used projects list
func updateRecentList() {
	if recentListBox == 0 {
		return
	}
	win32.ListBoxResetContent(recentListBox)
	for _, path := range recentProjects {
		win32.ListBoxAddString(recentListBox, path)
	}
}

// getSelectedRequestIndex returns the selected index in project list
func getSelectedRequestIndex() int {
	return win32.ListBoxGetCurSel(projectListBox)
}
