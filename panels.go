package main

import "hoermi.com/rest-test/win32"

// Panel visibility groups
type PanelGroup []win32.Controler

type Panels map[PanelName]PanelGroup

type PanelName string

const (
	PanelRequest     PanelName = "request"
	PanelProjectView PanelName = "projectView"
	PanelSettings    PanelName = "settings"
	PanelWelcome     PanelName = "welcome"
)

// initPanels initializes the panel groups (call after creating controls)
func initPanels(pw *ProjectWindow) Panels {
	panels := make(Panels)
	panels[PanelRequest] = PanelGroup{
		pw.methodCombo, pw.urlInput, pw.headersInput, pw.queryInput, pw.bodyInput,
		pw.responseOutput, pw.statusLabel, pw.sendBtn,
		pw.methodLabel, pw.urlLabel, pw.headersLabel, pw.queryLabel, pw.bodyLabel, pw.responseLabel,
	}

	panels[PanelProjectView] = PanelGroup{
		pw.projectListBox, pw.openReqBtn, pw.deleteReqBtn, pw.projectInfo, pw.saveBtn,
	}

	panels[PanelSettings] = PanelGroup{
		pw.certInput, pw.keyInput, pw.caInput, pw.skipVerifyChk,
		pw.certBtn, pw.keyBtn, pw.caBtn,
		pw.certLabel, pw.keyLabel, pw.caLabel, pw.settingsTitle,
		pw.saveSettingsBtn,
	}

	panels[PanelWelcome] = PanelGroup{
		pw.newTabTitle, pw.newTabNewBtn, pw.newTabOpenBtn,
		pw.recentLabel, pw.recentListBox,
	}

	return panels
}

// show shows a panel
func (panelGroup PanelGroup) show() {
	if panelGroup == nil {
		return
	}
	// Show the requested panel
	for _, ctrl := range panelGroup {
		ctrl.Show()
	}
}

// hide hides a panel
func (panelGroup PanelGroup) hide() {
	if panelGroup == nil {
		return
	}
	for _, ctrl := range panelGroup {
		ctrl.Hide()
	}
}

// showRequestPanel shows the request editing panel
func (p Panels) show(panel PanelName) {
	for name, pg := range p {
		if name == panel {
			pg.show()
		} else {
			pg.hide()
		}
	}
}
