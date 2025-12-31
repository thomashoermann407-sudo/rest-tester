package main

import (
	"fmt"
	"strings"

	"hoermi.com/rest-test/win32"
)

// Panel visibility groups
type PanelGroup interface {
	Controller() []win32.Controller
	Resize(tabHeight, width, height int32)

	// TODO: Use generic
	SaveState()
	SetState(data any)

	HandleCommand(id int, notifyCode int)
}

type Panels map[PanelGroupName]PanelGroup

type PanelGroupName string

const (
	PanelRequest     PanelGroupName = "request"
	PanelProjectView PanelGroupName = "projectView"
	PanelSettings    PanelGroupName = "settings"
	PanelWelcome     PanelGroupName = "welcome"
)

type requestPanelGroup struct {
	methodCombo    *win32.ComboBoxControl
	urlInput       *win32.Control
	headersInput   *win32.Control
	queryInput     *win32.Control
	bodyInput      *win32.Control
	responseOutput *win32.Control
	statusLabel    *win32.Control
	sendBtn        *win32.ButtonControl
	// Labels for request panel
	methodLabel   *win32.Control
	urlLabel      *win32.Control
	headersLabel  *win32.Control
	queryLabel    *win32.Control
	bodyLabel     *win32.Control
	responseLabel *win32.Control

	projectWindow *ProjectWindow
	content       *RequestTabContent
}

func (r *requestPanelGroup) Controller() []win32.Controller {
	return []win32.Controller{
		r.methodCombo, r.urlInput, r.headersInput, r.queryInput, r.bodyInput,
		r.responseOutput, r.statusLabel, r.sendBtn,
		r.methodLabel, r.urlLabel, r.headersLabel, r.queryLabel, r.bodyLabel, r.responseLabel,
	}
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

	r.methodCombo.MoveWindow(layoutPadding+methodLabelWidth+layoutPadding, y, methodComboWidth, 200, true)

	urlLabelWidth := int32(30)
	urlX := layoutPadding + methodLabelWidth + layoutPadding + methodComboWidth + layoutPadding + urlLabelWidth + layoutPadding
	urlWidth := availableWidth - methodLabelWidth - methodComboWidth - urlLabelWidth - sendBtnWidth - layoutPadding*2 - layoutPadding*2
	r.urlInput.MoveWindow(urlX, y, urlWidth, layoutInputHeight, true)
	r.sendBtn.MoveWindow(width-layoutPadding-sendBtnWidth, y, sendBtnWidth, layoutInputHeight, true)

	// === Query Parameters & Headers Section ===
	y += layoutInputHeight + layoutPadding
	y += layoutLabelHeight + layoutPadding

	// Split width 50/50 for params and headers
	halfWidth := (availableWidth - layoutPadding) / 2

	r.queryInput.MoveWindow(layoutPadding, y, halfWidth, paramsHeight, true)
	r.headersInput.MoveWindow(layoutPadding+halfWidth+layoutPadding, y, halfWidth, paramsHeight, true)

	// === Body Section ===
	y += paramsHeight + layoutPadding
	y += layoutLabelHeight + layoutPadding
	r.bodyInput.MoveWindow(layoutPadding, y, availableWidth, bodyHeight, true)

	// === Response Section ===
	y += bodyHeight + layoutPadding
	y += layoutLabelHeight + layoutPadding
	r.responseOutput.MoveWindow(layoutPadding, y, availableWidth, responseHeight, true)
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
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for i, m := range methods {
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

func (r *requestPanelGroup) HandleCommand(id int, notifyCode int) {
	switch id {
	case r.sendBtn.ID:
		// Get the bound request from the current tab
		r.SaveState()
		request := r.content.BoundRequest
		if request == nil {
			r.statusLabel.SetText("❌ No request")
			r.responseOutput.SetText("Error: No request bound to this tab")
			return
		}

		r.statusLabel.SetText("⏳ Sending...")
		r.responseOutput.SetText("")
		go sendRequest(request, r.content.Settings, func(response string, err error) {
			if err != nil {
				r.statusLabel.SetText("❌ Error")
				r.responseOutput.SetText(fmt.Sprintf("Error sending request:\r\n%v", err))
				return
			}
			r.statusLabel.SetText("✅ Success")
			r.responseOutput.SetText(response)
		})
	}
}

func createRequestPanel(projectWindow *ProjectWindow) *requestPanelGroup {
	mainWindow := projectWindow.mainWindow
	panelGroup := &requestPanelGroup{
		projectWindow:  projectWindow,
		methodLabel:    mainWindow.CreateLabel("Method", 0, 0, 0, 0),
		methodCombo:    mainWindow.CreateComboBox(0, 0, 0, 0),
		urlLabel:       mainWindow.CreateLabel("URL", 0, 0, 0, 0),
		urlInput:       mainWindow.CreateInput(0, 0, 680, layoutInputHeight),
		sendBtn:        mainWindow.CreateButton("Send", 0, 0, 0, layoutInputHeight),
		queryLabel:     mainWindow.CreateLabel("Query Parameters (one per line: key=value)", 0, 0, 300, layoutLabelHeight),
		queryInput:     mainWindow.CreateCodeEdit(0, 0, 520, 70, false),
		headersLabel:   mainWindow.CreateLabel("Headers (one per line: Header: value)", 0, 0, 300, layoutLabelHeight),
		headersInput:   mainWindow.CreateCodeEdit(0, 0, 520, 70, false),
		bodyLabel:      mainWindow.CreateLabel("Request Body", 0, 0, 150, layoutLabelHeight),
		bodyInput:      mainWindow.CreateCodeEdit(0, 0, defaultWindowWidth-layoutPadding*2, 120, false),
		responseLabel:  mainWindow.CreateLabel("Response", 0, 0, 80, layoutLabelHeight),
		statusLabel:    mainWindow.CreateLabel("Ready", 0, 0, 400, layoutLabelHeight),
		responseOutput: mainWindow.CreateCodeEdit(0, 0, defaultWindowWidth-layoutPadding*2, 350, true),
	}

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"} {
		panelGroup.methodCombo.AddString(method)
	}
	panelGroup.methodCombo.SetCurSel(0)

	return panelGroup
}

type projectViewPanelGroup struct {
	// Project View Panel controls
	projectListBox *win32.ListBoxControl
	openReqBtn     *win32.ButtonControl
	deleteReqBtn   *win32.ButtonControl
	projectInfo    *win32.Control
	saveBtn        *win32.ButtonControl

	projectWindow *ProjectWindow
	content       *ProjectViewTabContent
}

func (p *projectViewPanelGroup) Controller() []win32.Controller {
	return []win32.Controller{
		p.projectListBox, p.openReqBtn, p.deleteReqBtn, p.projectInfo, p.saveBtn,
	}
}

func (p *projectViewPanelGroup) Resize(tabHeight, width, height int32) {
	// Currently no dynamic resizing needed for project view panel
}

func (p *projectViewPanelGroup) SaveState() {
	p.content.SelectedIndex = p.projectListBox.GetCurSel()
}

func (p *projectViewPanelGroup) SetState(data any) {
	content, ok := data.(*ProjectViewTabContent)
	if !ok {
		return
	}
	if content == nil {
		return
	}
	p.content = content
	if p.projectListBox == nil || content.BoundProject == nil {
		return
	}

	// Clear the listbox
	p.projectListBox.ResetContent()

	// Add all requests
	for _, req := range content.BoundProject.Requests {
		displayText := req.Method + " " + req.Name
		if req.URL != "" {
			shortURL := req.URL
			if len(shortURL) > 40 {
				shortURL = shortURL[:40] + "..."
			}
			displayText = req.Method + " " + shortURL
		}
		p.projectListBox.AddString(displayText)
	}
	// Restore listbox selection
	if content.SelectedIndex >= 0 {
		p.projectListBox.SetCurSel(content.SelectedIndex)
	}
}

func (p *projectViewPanelGroup) HandleCommand(id int, notifyCode int) {
	switch id {
	case p.saveBtn.ID:
		p.saveProject()
	case p.openReqBtn.ID:
		p.openSelectedRequest()
	case p.deleteReqBtn.ID:
		p.deleteSelectedRequest()
	case p.projectListBox.ID:
		// Open request on double-click
		if notifyCode == win32.LBN_DBLCLK {
			p.openSelectedRequest()
		}
	}
}

// saveProject saves the current project to file
func (p *projectViewPanelGroup) saveProject() {

	defaultName := p.content.BoundProject.Name + ".rtp"
	filePath, ok := p.projectWindow.SaveFileDialog(
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := p.content.BoundProject.Save(filePath); err != nil {
		p.projectWindow.MessageBox(fmt.Sprintf("Error saving project: %v", err), "Error")
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
func (p *projectViewPanelGroup) openSelectedRequest() {
	idx := p.projectListBox.GetCurSel()
	if idx < 0 || idx >= len(p.content.BoundProject.Requests) {
		return
	}

	req := p.content.BoundProject.Requests[idx]

	// Open in new tab bound to the request
	p.projectWindow.CreateRequestTab(req)
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
	mainWindow := projectWindow.mainWindow
	y := layoutPadding * 4 // Below tab bar area
	dy := layoutLabelHeight + layoutPadding
	listWidth := int32(500)
	listHeight := int32(500)
	btnX := layoutPadding + listWidth + layoutPadding

	panelGroup := &projectViewPanelGroup{
		projectWindow:  projectWindow,
		projectInfo:    mainWindow.CreateLabel("Double-click a request to open it in a new tab", layoutPadding, y, listWidth, layoutLabelHeight),
		projectListBox: mainWindow.CreateListBox(layoutPadding, y+dy, listWidth, listHeight),
		openReqBtn:     mainWindow.CreateButton("Open in Tab", btnX, y+dy, layoutButtonWidth, layoutInputHeight),
		deleteReqBtn:   mainWindow.CreateButton("Delete", btnX, y+2*dy, layoutButtonWidth, layoutInputHeight),
		saveBtn:        mainWindow.CreateButton("Save Project", btnX, y+3*dy, layoutButtonWidth, layoutInputHeight),
	}

	// Action buttons to the right of the list
	return panelGroup
}

type settingsPanelGroup struct {
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

	projectWindow *ProjectWindow
	content       *SettingsTabContent
}

func (s *settingsPanelGroup) Controller() []win32.Controller {
	return []win32.Controller{
		s.certInput, s.keyInput, s.caInput, s.skipVerifyChk,
		s.certBtn, s.keyBtn, s.caBtn,
		s.certLabel, s.keyLabel, s.caLabel, s.settingsTitle,
		s.saveSettingsBtn,
	}
}

func (s *settingsPanelGroup) Resize(tabHeight, width, height int32) {
	// Currently no dynamic resizing needed for settings panel
}

func (s *settingsPanelGroup) SaveState() {
	s.content.Certificate.CertFile = s.certInput.GetText()
	s.content.Certificate.KeyFile = s.keyInput.GetText()
	s.content.Certificate.CACertFile = s.caInput.GetText()
	s.content.Certificate.SkipVerify = s.skipVerifyChk.IsChecked()
}

func (s *settingsPanelGroup) SetState(data any) {
	content, ok := data.(*SettingsTabContent)
	if !ok {
		return
	}
	if content == nil {
		return
	}
	s.content = content
	s.certInput.SetText(content.Certificate.CertFile)
	s.keyInput.SetText(content.Certificate.KeyFile)
	s.caInput.SetText(content.Certificate.CACertFile)
	s.skipVerifyChk.SetChecked(content.Certificate.SkipVerify)
}

func (s *settingsPanelGroup) HandleCommand(id int, notifyCode int) {
	switch id {
	case s.certBtn.ID:
		if path, ok := s.projectWindow.OpenFileDialog("Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			s.certInput.SetText(path)
		}
	case s.keyBtn.ID:
		if path, ok := s.projectWindow.OpenFileDialog("Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			s.keyInput.SetText(path)
		}
	case s.caBtn.ID:
		if path, ok := s.projectWindow.OpenFileDialog("Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			s.caInput.SetText(path)
		}
	case s.saveSettingsBtn.ID:
		s.saveCertificateConfig()
	}
}

// saveCertificateConfig saves the certificate settings from UI to project
func (s *settingsPanelGroup) saveCertificateConfig() {
	s.SaveState()
	s.projectWindow.settings.save()
}

func createSettingsPanel(projectWindow *ProjectWindow) *settingsPanelGroup {
	mainWindow := projectWindow.mainWindow
	y := layoutPadding * 4 // Below tab bar area
	inputWidth := int32(400)
	browseBtnWidth := int32(30)

	settingsPanel := &settingsPanelGroup{}
	settingsPanel.projectWindow = projectWindow
	settingsPanel.settingsTitle = mainWindow.CreateLabel("Global Settings", layoutPadding, y, 200, layoutLabelHeight+4)
	y += layoutLabelHeight + layoutPadding

	settingsPanel.certLabel = mainWindow.CreateLabel("Client Certificate (PEM)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	settingsPanel.certInput = mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	settingsPanel.certBtn = mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutPadding, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	settingsPanel.keyLabel = mainWindow.CreateLabel("Private Key (PEM)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	settingsPanel.keyInput = mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	settingsPanel.keyBtn = mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutPadding, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	settingsPanel.caLabel = mainWindow.CreateLabel("CA Bundle (optional)", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	settingsPanel.caInput = mainWindow.CreateInput(layoutPadding, y, inputWidth, layoutInputHeight)
	settingsPanel.caBtn = mainWindow.CreateButton("...", layoutPadding+inputWidth+layoutPadding, y, browseBtnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	settingsPanel.skipVerifyChk = mainWindow.CreateCheckbox("Skip TLS Verification (insecure)", layoutPadding, y, 280, layoutInputHeight)
	y += layoutInputHeight + layoutPadding
	settingsPanel.saveSettingsBtn = mainWindow.CreateButton("Save Settings", layoutPadding, y, layoutButtonWidth, layoutInputHeight)

	return settingsPanel
}

type welcomePanelGroup struct {
	// Welcome Panel controls
	newTabTitle   *win32.Control
	newTabNewBtn  *win32.ButtonControl
	newTabOpenBtn *win32.ButtonControl
	recentLabel   *win32.Control
	recentListBox *win32.ListBoxControl

	projectWindow *ProjectWindow
	content       *WelcomeTabContent
}

func (w *welcomePanelGroup) Controller() []win32.Controller {
	return []win32.Controller{
		w.newTabTitle, w.newTabNewBtn, w.newTabOpenBtn,
		w.recentLabel, w.recentListBox,
	}
}

func (w *welcomePanelGroup) Resize(tabHeight, width, height int32) {
	// Currently no dynamic resizing needed for welcome panel
}

func (w *welcomePanelGroup) SaveState() {
	// Capture selected recent project index
	w.content.SelectedRecentIndex = w.recentListBox.GetCurSel()
}

func (w *welcomePanelGroup) SetState(data any) {
	content, ok := data.(*WelcomeTabContent)
	if !ok {
		return
	}
	w.content = content
	w.recentListBox.ResetContent()
	for _, path := range content.RecentProjects {
		// Display only the filename, not the full path
		displayName := path
		if idx := strings.LastIndex(path, "\\"); idx >= 0 {
			displayName = path[idx+1:]
		}
		w.recentListBox.AddString(displayName)
	}
}

func (w *welcomePanelGroup) HandleCommand(id int, notifyCode int) {
	switch id {
	case w.newTabNewBtn.ID:
		w.projectWindow.newProject()
	case w.newTabOpenBtn.ID:
		w.projectWindow.openProject()
	case w.recentListBox.ID:
		// Only open on double-click, not selection change
		if notifyCode == win32.LBN_DBLCLK {
			idx := w.recentListBox.GetCurSel()
			if idx < 0 || idx >= len(w.projectWindow.settings.RecentProjects) {
				return
			}
			w.projectWindow.openProjectFromPath(w.projectWindow.settings.RecentProjects[idx])
		}
	}
}

func createWelcomePanel(projectWindow *ProjectWindow) *welcomePanelGroup {
	mainWindow := projectWindow.mainWindow
	y := layoutPadding * 6 // More space at top for welcome area
	recentListWidth := int32(400)
	recentListHeight := int32(300)

	newTabTitle := mainWindow.CreateLabel("REST Tester - Start", layoutPadding, y, 400, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding

	newTabNewBtn := mainWindow.CreateButton("📄 New Project", layoutPadding, y, layoutButtonWidth, layoutIconInputHeight)
	newTabOpenBtn := mainWindow.CreateButton("📂 Open Project", layoutPadding+layoutButtonWidth, y, layoutButtonWidth, layoutIconInputHeight)
	y += layoutIconInputHeight + layoutPadding

	recentLabel := mainWindow.CreateLabel("Recent Projects:", layoutPadding, y, 200, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	recentListBox := mainWindow.CreateListBox(layoutPadding, y, recentListWidth, recentListHeight)

	return &welcomePanelGroup{
		projectWindow: projectWindow,
		newTabTitle:   newTabTitle,
		newTabNewBtn:  newTabNewBtn,
		newTabOpenBtn: newTabOpenBtn,
		recentLabel:   recentLabel,
		recentListBox: recentListBox,
	}
}

// initPanels initializes the panel groups (call after creating controls)
func initPanels(projectWindow *ProjectWindow) Panels {
	panels := make(Panels)
	panels[PanelRequest] = createRequestPanel(projectWindow)
	panels[PanelProjectView] = createProjectViewPanel(projectWindow)
	panels[PanelSettings] = createSettingsPanel(projectWindow)
	panels[PanelWelcome] = createWelcomePanel(projectWindow)
	for _, pg := range panels {
		pg.Resize(projectWindow.mainWindow.TabManager.GetHeight(), projectWindow.mainWindow.GetWidth(), projectWindow.mainWindow.GetHeight())
	}
	return panels
}

// show shows the request editing panel
func (p Panels) show(panel PanelGroupName) {
	for name, pg := range p {
		if name == panel {
			for _, ctrl := range pg.Controller() {
				ctrl.Show()
			}
		} else {
			for _, ctrl := range pg.Controller() {
				ctrl.Hide()
			}
		}
	}
}

func (p Panels) get(panel PanelGroupName) PanelGroup {
	return p[panel]
}

func (p Panels) handleCommand(id int, notifyCode int) {
	for _, pg := range p {
		pg.HandleCommand(id, notifyCode)
	}
}
