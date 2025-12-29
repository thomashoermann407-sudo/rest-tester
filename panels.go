package main

import "hoermi.com/rest-test/win32"

// Panel visibility groups
type PanelGroup struct {
	controls []win32.HWND
	visible  bool
}

type Panels struct {
	requestPanel     *PanelGroup
	projectViewPanel *PanelGroup
	settingsPanel    *PanelGroup
	newTabPanel      *PanelGroup
	projectWindow    *ProjectWindow
}

// initPanels initializes the panel groups (call after creating controls)
func initPanels(pw *ProjectWindow) *Panels {
	requestPanel := &PanelGroup{
		controls: []win32.HWND{
			pw.methodCombo, pw.urlInput, pw.headersInput, pw.queryInput, pw.bodyInput,
			pw.responseOutput, pw.statusLabel, pw.sendBtn,
			pw.methodLabel, pw.urlLabel, pw.headersLabel, pw.queryLabel, pw.bodyLabel, pw.responseLabel,
		},
		visible: false,
	}

	projectViewPanel := &PanelGroup{
		controls: []win32.HWND{pw.projectListBox, pw.openReqBtn, pw.deleteReqBtn, pw.projectInfo, pw.saveBtn},
		visible:  false,
	}

	settingsPanel := &PanelGroup{
		controls: []win32.HWND{
			pw.certInput, pw.keyInput, pw.caInput, pw.skipVerifyChk,
			pw.certBtn, pw.keyBtn, pw.caBtn,
			pw.certLabel, pw.keyLabel, pw.caLabel, pw.settingsTitle,
		},
		visible: false,
	}

	newTabPanel := &PanelGroup{
		controls: []win32.HWND{
			pw.newTabTitle, pw.newTabNewBtn, pw.newTabOpenBtn,
			pw.recentLabel, pw.recentListBox,
		},
		visible: true,
	}
	panels := &Panels{
		requestPanel:     requestPanel,
		projectViewPanel: projectViewPanel,
		settingsPanel:    settingsPanel,
		newTabPanel:      newTabPanel,
	}
	//TODO: refactor to avoid circular reference
	pw.panels = panels
	panels.projectWindow = pw
	return panels
}

// showPanel shows a panel and hides others
func (p *Panels) showPanel(panel *PanelGroup) {
	// Hide all panels
	hidePanel(p.requestPanel)
	hidePanel(p.projectViewPanel)
	hidePanel(p.settingsPanel)
	hidePanel(p.newTabPanel)

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
func (p *Panels) showRequestPanel() {
	p.showPanel(p.requestPanel)
}

// showProjectViewPanel shows the project structure panel
func (p *Panels) showProjectViewPanel() {
	p.showPanel(p.projectViewPanel)
	p.projectWindow.updateProjectList()
}

// showSettingsPanel shows the global settings panel
func (p *Panels) showSettingsPanel() {
	p.showPanel(p.settingsPanel)
	p.projectWindow.loadCertificateUI()
}

// showWelcomePanel shows the welcome panel
func (p *Panels) showWelcomePanel() {
	p.showPanel(p.newTabPanel)
	p.projectWindow.updateRecentList()
}
