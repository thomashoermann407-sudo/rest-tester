package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"hoermi.com/rest-test/win32"
)

const (
	// Common spacing
	layoutPadding = int32(12)

	// Standard control heights
	layoutColumnWidth     = int32(400)
	layoutLabelHeight     = int32(20)
	layoutInputHeight     = int32(26)
	layoutIconInputHeight = int32(32)
	layoutButtonWidth     = int32(120)
	layoutListHeight      = int32(300)

	PanelRequest     win32.PanelGroupName = "request"
	PanelProjectView win32.PanelGroupName = "projectView"
	PanelSettings    win32.PanelGroupName = "settings"
	PanelWelcome     win32.PanelGroupName = "welcome"
)

var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

type requestPanelGroup struct {
	*win32.ControllerGroup
	methodCombo    *win32.ComboBoxControl
	urlInput       *win32.Control
	headersInput   *win32.Control
	queryInput     *win32.Control
	bodyInput      *win32.Control
	responseOutput *win32.Control
	statusLabel    *win32.Control
	sendBtn        *win32.ButtonControl
	methodLabel    *win32.Control
	urlLabel       *win32.Control
	headersLabel   *win32.Control
	queryLabel     *win32.Control
	bodyLabel      *win32.Control
	responseLabel  *win32.Control

	content *RequestTabContent
}

func (r *requestPanelGroup) Resize(tabHeight, width, height int32) {
	paramsHeightRatio := 0.10
	bodyHeightRatio := 0.15

	// Calculate available height (excluding tab bar and padding)
	availableHeight := height - tabHeight - layoutPadding*5 - layoutLabelHeight*5 - layoutInputHeight

	// Calculate panel heights based on ratios
	minParamsHeight := int32(60)
	minBodyHeight := int32(80)
	minResponseHeight := int32(150)

	paramsHeight := max(int32(float64(availableHeight)*paramsHeightRatio), minParamsHeight)
	bodyHeight := max(int32(float64(availableHeight)*bodyHeightRatio), minBodyHeight)
	responseHeight := max(availableHeight-paramsHeight-bodyHeight, minResponseHeight)

	availableWidth := width - layoutPadding*2

	y := tabHeight + layoutPadding

	// === Request Row (fixed height) ===
	methodLabelWidth := int32(50)
	methodComboWidth := int32(90)
	sendBtnWidth := int32(90)

	// Position method label and combo
	r.methodLabel.MoveWindow(layoutPadding, y+3, methodLabelWidth, layoutLabelHeight)
	r.methodCombo.MoveWindow(layoutPadding+methodLabelWidth+layoutPadding, y, methodComboWidth, 200)

	// Position URL label and input
	urlLabelWidth := int32(30)
	urlX := layoutPadding + methodLabelWidth + layoutPadding + methodComboWidth + layoutPadding
	r.urlLabel.MoveWindow(urlX, y+3, urlLabelWidth, layoutLabelHeight)

	urlInputX := urlX + urlLabelWidth + layoutPadding
	urlWidth := availableWidth - methodLabelWidth - methodComboWidth - urlLabelWidth - sendBtnWidth - layoutPadding*4
	r.urlInput.MoveWindow(urlInputX, y, urlWidth, layoutInputHeight)
	r.sendBtn.MoveWindow(width-layoutPadding-sendBtnWidth, y, sendBtnWidth, layoutInputHeight)

	// === Query Parameters & Headers Section ===
	y += layoutInputHeight + layoutPadding

	// Position section labels
	halfWidth := (availableWidth - layoutPadding) / 2
	r.queryLabel.MoveWindow(layoutPadding, y, 300, layoutLabelHeight)
	r.headersLabel.MoveWindow(layoutPadding+halfWidth+layoutPadding, y, 300, layoutLabelHeight)

	y += layoutLabelHeight + layoutPadding
	r.queryInput.MoveWindow(layoutPadding, y, halfWidth, paramsHeight)
	r.headersInput.MoveWindow(layoutPadding+halfWidth+layoutPadding, y, halfWidth, paramsHeight)

	// === Body Section ===
	y += paramsHeight + layoutPadding
	r.bodyLabel.MoveWindow(layoutPadding, y, 150, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	r.bodyInput.MoveWindow(layoutPadding, y, availableWidth, bodyHeight)

	// === Response Section ===
	y += bodyHeight + layoutPadding
	r.responseLabel.MoveWindow(layoutPadding, y, 80, layoutLabelHeight)
	r.statusLabel.MoveWindow(layoutPadding+90, y, 400, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	r.responseOutput.MoveWindow(layoutPadding, y, availableWidth, responseHeight)
}

func (r *requestPanelGroup) SaveState() {
	req := r.content.BoundRequest
	req.Method = r.methodCombo.GetText()
	req.URL = r.urlInput.GetText()
	req.Body = r.bodyInput.GetText()
	req.Headers = ParseParams(r.headersInput.GetText())
	req.QueryParams = ParseParams(r.queryInput.GetText())
	r.content.Response = r.responseOutput.GetText()
	r.content.Status = r.statusLabel.GetText()
}

func (r *requestPanelGroup) SetState(data any) {
	content, ok := data.(*RequestTabContent)
	if !ok {
		return
	}
	if content == nil {
		return
	}
	r.content = content
	req := content.BoundRequest

	// Set method
	for i, m := range httpMethods {
		if m == req.Method {
			r.methodCombo.SetCurSel(i)
			break
		}
	}

	r.urlInput.SetText(req.URL)
	r.headersInput.SetText(req.Headers.Format())
	r.queryInput.SetText(req.QueryParams.Format())
	r.bodyInput.SetText(req.Body)
	r.responseOutput.SetText(content.Response)
	r.statusLabel.SetText(content.Status)
}

func createRequestPanel(projectWindow *ProjectWindow) *requestPanelGroup {
	group := &requestPanelGroup{
		methodLabel:    projectWindow.mainWindow.CreateLabel("Method"),
		methodCombo:    projectWindow.mainWindow.CreateComboBox(),
		urlLabel:       projectWindow.mainWindow.CreateLabel("URL"),
		urlInput:       projectWindow.mainWindow.CreateInput(),
		queryLabel:     projectWindow.mainWindow.CreateLabel("Query Parameters (one per line: key=value)"),
		queryInput:     projectWindow.mainWindow.CreateCodeEdit(false),
		headersLabel:   projectWindow.mainWindow.CreateLabel("Headers (one per line: Header: value)"),
		headersInput:   projectWindow.mainWindow.CreateCodeEdit(false),
		bodyLabel:      projectWindow.mainWindow.CreateLabel("Request Body"),
		bodyInput:      projectWindow.mainWindow.CreateCodeEdit(false),
		responseLabel:  projectWindow.mainWindow.CreateLabel("Response"),
		statusLabel:    projectWindow.mainWindow.CreateLabel("Ready"),
		responseOutput: projectWindow.mainWindow.CreateCodeEdit(true),
	}
	group.sendBtn = projectWindow.mainWindow.CreateButton("Send", func() {
		// Get the bound request from the current tab
		group.SaveState()
		request := group.content.BoundRequest
		if request == nil {
			group.statusLabel.SetText("❌ No request")
			group.responseOutput.SetText("Error: No request bound to this tab")
			return
		}

		group.statusLabel.SetText("⏳ Sending...")
		group.responseOutput.SetText("")

		// Get timeout from project settings (default 30000ms if not set)
		timeoutInMs := int64(30000)
		if group.content.BoundProject != nil && group.content.BoundProject.Settings.TimeoutInMs > 0 {
			timeoutInMs = group.content.BoundProject.Settings.TimeoutInMs
		}

		// Send request in background goroutine
		go sendRequest(request, group.content.Settings, timeoutInMs, func(response string, err error) {
			// Marshal the UI update back to the main thread using PostUICallback
			projectWindow.mainWindow.PostUICallback(func() {
				if err != nil {
					group.statusLabel.SetText("❌ Error")
					group.responseOutput.SetText(fmt.Sprintf("Error sending request:\r\n%v", err))
					return
				}
				group.statusLabel.SetText("✅ Success")
				group.responseOutput.SetText(response)
			})
		})
	})
	for _, method := range httpMethods {
		group.methodCombo.AddString(method)
	}
	group.methodCombo.SetCurSel(0)
	group.ControllerGroup = win32.NewControllerGroup(group.methodCombo, group.urlInput, group.headersInput, group.queryInput, group.bodyInput,
		group.responseOutput, group.statusLabel, group.sendBtn,
		group.methodLabel, group.urlLabel, group.headersLabel, group.queryLabel, group.bodyLabel, group.responseLabel,
	)
	return group
}

type projectViewPanelGroup struct {
	*win32.ControllerGroup
	// Project View Panel controls
	projectListBox *win32.ListBoxControl
	openReqBtn     *win32.ButtonControl
	deleteReqBtn   *win32.ButtonControl
	projectInfo    *win32.Control
	saveBtn        *win32.ButtonControl
	timeoutLabel   *win32.Control
	timeoutInput   *win32.Control

	content *ProjectViewTabContent
}

func (p *projectViewPanelGroup) Resize(tabHeight, width, height int32) {
	y := layoutPadding * 4 // Below tab bar area
	dy := layoutLabelHeight + layoutPadding
	listWidth := int32(500)
	listHeight := int32(450)
	btnX := layoutPadding + listWidth + layoutPadding
	p.projectInfo.MoveWindow(layoutPadding, y, listWidth, layoutLabelHeight)
	p.projectListBox.MoveWindow(layoutPadding, y+dy, listWidth, listHeight)
	p.openReqBtn.MoveWindow(btnX, y+dy, layoutButtonWidth, layoutInputHeight)
	p.deleteReqBtn.MoveWindow(btnX, y+2*dy, layoutButtonWidth, layoutInputHeight)
	p.saveBtn.MoveWindow(btnX, y+3*dy, layoutButtonWidth, layoutInputHeight)

	// Timeout settings below the list
	y += dy + listHeight + layoutPadding
	p.timeoutLabel.MoveWindow(layoutPadding, y, int32(200), layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding/2
	p.timeoutInput.MoveWindow(layoutPadding, y, int32(150), layoutInputHeight)
}

func (p *projectViewPanelGroup) SaveState() {
	p.content.SelectedIndex = p.projectListBox.GetCurSel()

	// Save timeout setting
	timeoutText := p.timeoutInput.GetText()
	if timeoutText != "" {
		var timeout int64
		fmt.Sscanf(timeoutText, "%d", &timeout)
		if timeout > 0 {
			p.content.BoundProject.Settings.TimeoutInMs = timeout
		}
	}
}

func (p *projectViewPanelGroup) SetState(data any) {
	if content, ok := data.(*ProjectViewTabContent); ok {
		p.content = content
		if p.projectListBox == nil || content.BoundProject == nil {
			return
		}

		// Clear the listbox
		p.projectListBox.ResetContent()

		// Add all requests
		for _, req := range content.BoundProject.Requests {
			displayText := req.Method + " " + filepath.Base(req.URL)
			p.projectListBox.AddString(displayText)
		}
		// Restore listbox selection
		if content.SelectedIndex >= 0 {
			p.projectListBox.SetCurSel(content.SelectedIndex)
		}

		// Set timeout value
		timeout := content.BoundProject.Settings.TimeoutInMs
		if timeout == 0 {
			timeout = 30000 // Default 30 seconds
		}
		p.timeoutInput.SetText(fmt.Sprintf("%d", timeout))
	}
}

// saveProject saves the current project to file
func (p *projectViewPanelGroup) saveProject(mainWindow *win32.Window) {

	defaultName := p.content.BoundProject.Name + ".rtp"
	filePath, ok := mainWindow.SaveFileDialog(
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := p.content.BoundProject.Save(filePath); err != nil {
		mainWindow.MessageBox(fmt.Sprintf("Error saving project: %v", err), "Error")
		return
	}

	// Update project name from filename
	name := filePath
	if idx := strings.LastIndex(name, "\\"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".rtp")
	p.content.BoundProject.Name = name
}

// openSelectedRequest opens the selected request from project list in a new tab
func (p *projectViewPanelGroup) openSelectedRequest(projectWindow *ProjectWindow) {
	idx := p.projectListBox.GetCurSel()
	if idx < 0 || idx >= len(p.content.BoundProject.Requests) {
		return
	}

	req := p.content.BoundProject.Requests[idx]

	// Open in new tab bound to the request
	projectWindow.createRequestTab(req)
}

// deleteSelectedRequest removes the selected request from project
func (p *projectViewPanelGroup) deleteSelectedRequest() {
	idx := p.projectListBox.GetCurSel()
	if idx < 0 || idx >= len(p.content.BoundProject.Requests) {
		return
	}

	// Remove from project
	p.content.BoundProject.RemoveRequest(idx)
	p.SetState(p.content)
}

func createProjectViewPanel(projectWindow *ProjectWindow) *projectViewPanelGroup {
	group := &projectViewPanelGroup{
		projectInfo:  projectWindow.mainWindow.CreateLabel("Double-click a request to open it in a new tab"),
		timeoutLabel: projectWindow.mainWindow.CreateLabel("Request Timeout (milliseconds):"),
		timeoutInput: projectWindow.mainWindow.CreateInput(),
	}
	group.projectListBox = projectWindow.mainWindow.CreateListBox(func(lbc *win32.ListBoxControl) {
		group.openSelectedRequest(projectWindow)
	})
	group.openReqBtn = projectWindow.mainWindow.CreateButton("Open in Tab", func() {
		group.openSelectedRequest(projectWindow)
	})
	group.deleteReqBtn = projectWindow.mainWindow.CreateButton("Delete", func() {
		group.deleteSelectedRequest()
	})
	group.saveBtn = projectWindow.mainWindow.CreateButton("Save Project", func() {
		group.SaveState() // Save timeout before saving to file
		group.saveProject(projectWindow.mainWindow)
	})
	group.ControllerGroup = win32.NewControllerGroup(
		group.projectListBox, group.openReqBtn, group.deleteReqBtn,
		group.projectInfo, group.saveBtn, group.timeoutLabel, group.timeoutInput,
	)
	return group
}

type settingsPanelGroup struct {
	*win32.ControllerGroup
	// Settings Panel controls
	certInput       *win32.Control
	keyInput        *win32.Control
	caInput         *win32.Control
	skipVerifyChk   *win32.CheckBoxControl
	certBtn         *win32.ButtonControl
	keyBtn          *win32.ButtonControl
	caBtn           *win32.ButtonControl
	certLabel       *win32.Control
	keyLabel        *win32.Control
	caLabel         *win32.Control
	settingsTitle   *win32.Control
	saveSettingsBtn *win32.ButtonControl

	content *SettingsTabContent
}

func (s *settingsPanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight

	s.settingsTitle.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight+4)
	y += layoutLabelHeight + layoutPadding

	s.certLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.certInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.certBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.keyLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.keyInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.keyBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.caLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.caInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.caBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.skipVerifyChk.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding
	s.saveSettingsBtn.MoveWindow(layoutPadding, y, layoutButtonWidth, layoutInputHeight)
}

func (s *settingsPanelGroup) SaveState() {
	s.content.Settings.Certificate.CertFile = s.certInput.GetText()
	s.content.Settings.Certificate.KeyFile = s.keyInput.GetText()
	s.content.Settings.Certificate.CACertFile = s.caInput.GetText()
	s.content.Settings.Certificate.SkipVerify = s.skipVerifyChk.IsChecked()
}

func (s *settingsPanelGroup) SetState(data any) {
	if content, ok := data.(*SettingsTabContent); ok {
		s.content = content
		s.certInput.SetText(content.Settings.Certificate.CertFile)
		s.keyInput.SetText(content.Settings.Certificate.KeyFile)
		s.caInput.SetText(content.Settings.Certificate.CACertFile)
		s.skipVerifyChk.SetChecked(content.Settings.Certificate.SkipVerify)
	}
}

func createSettingsPanel(projectWindow *ProjectWindow) *settingsPanelGroup {
	group := &settingsPanelGroup{
		settingsTitle: projectWindow.mainWindow.CreateLabel("Global Settings"),
		certLabel:     projectWindow.mainWindow.CreateLabel("Client Certificate (PEM)"),
		certInput:     projectWindow.mainWindow.CreateInput(),
		keyLabel:      projectWindow.mainWindow.CreateLabel("Private Key (PEM)"),
		keyInput:      projectWindow.mainWindow.CreateInput(),
		caLabel:       projectWindow.mainWindow.CreateLabel("CA Bundle (optional)"),
		caInput:       projectWindow.mainWindow.CreateInput(),
		skipVerifyChk: projectWindow.mainWindow.CreateCheckbox("Skip TLS Verification (insecure)"),
	}
	group.certBtn = projectWindow.mainWindow.CreateButton("...", func() {
		if path, ok := projectWindow.mainWindow.OpenFileDialog("Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			group.certInput.SetText(path)
		}

	})
	group.keyBtn = projectWindow.mainWindow.CreateButton("...", func() {
		if path, ok := projectWindow.mainWindow.OpenFileDialog("Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			group.keyInput.SetText(path)
		}

	})
	group.caBtn = projectWindow.mainWindow.CreateButton("...", func() {
		if path, ok := projectWindow.mainWindow.OpenFileDialog("Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			group.caInput.SetText(path)
		}
	})
	group.saveSettingsBtn = projectWindow.mainWindow.CreateButton("Save Settings", func() {
		group.SaveState()
		group.content.Settings.save()
	})
	group.ControllerGroup = win32.NewControllerGroup(
		group.certInput, group.keyInput, group.caInput,
		group.skipVerifyChk, group.certBtn, group.keyBtn, group.caBtn,
		group.settingsTitle, group.certLabel, group.keyLabel, group.caLabel,
		group.saveSettingsBtn,
	)
	return group
}

type welcomePanelGroup struct {
	*win32.ControllerGroup
	// Welcome Panel controls
	newTabTitle   *win32.Control
	newTabNewBtn  *win32.ButtonControl
	newTabOpenBtn *win32.ButtonControl
	recentLabel   *win32.Control
	recentListBox *win32.ListBoxControl

	content *WelcomeTabContent
}

func (w *welcomePanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight

	w.newTabTitle.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding

	w.newTabNewBtn.MoveWindow(layoutPadding, y, layoutButtonWidth, layoutIconInputHeight)
	w.newTabOpenBtn.MoveWindow(layoutPadding+layoutButtonWidth, y, layoutButtonWidth, layoutIconInputHeight)
	y += layoutIconInputHeight + layoutPadding

	w.recentLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	w.recentListBox.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutListHeight)
}

func (w *welcomePanelGroup) SaveState() {
	// Capture selected recent project index
	w.content.SelectedRecentIndex = w.recentListBox.GetCurSel()
}

func (w *welcomePanelGroup) SetState(data any) {
	if content, ok := data.(*WelcomeTabContent); ok {
		w.content = content
		w.recentListBox.ResetContent()
		for _, path := range content.RecentProjects {
			w.recentListBox.AddString(filepath.Base(path))
		}
	}
}

func createWelcomePanel(projectWindow *ProjectWindow) *welcomePanelGroup {
	group := &welcomePanelGroup{
		newTabTitle:   projectWindow.mainWindow.CreateLabel("REST Tester - Start"),
		newTabNewBtn:  projectWindow.mainWindow.CreateButton("📄 New Project", func() { projectWindow.newProject() }),
		newTabOpenBtn: projectWindow.mainWindow.CreateButton("📂 Open Project", func() { projectWindow.openProject() }),
		recentLabel:   projectWindow.mainWindow.CreateLabel("Recent Projects:"),
		recentListBox: projectWindow.mainWindow.CreateListBox(func(list *win32.ListBoxControl) {
			idx := list.GetCurSel()
			if idx < 0 || idx >= len(projectWindow.settings.RecentProjects) {
				return
			}
			projectWindow.openProjectFromPath(projectWindow.settings.RecentProjects[idx])
		}),
	}
	group.ControllerGroup = win32.NewControllerGroup(
		group.newTabTitle, group.newTabNewBtn, group.newTabOpenBtn,
		group.recentLabel, group.recentListBox,
	)
	return group
}
