package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"hoermi.com/rest-test/win32"
)

// Control IDs
const (
	ID_METHOD_COMBO   = 100
	ID_SEND_BTN       = 101
	ID_NEW_BTN        = 102
	ID_OPEN_BTN       = 103
	ID_SAVE_BTN       = 104
	ID_CERT_BTN       = 105
	ID_KEY_BTN        = 106
	ID_CA_BTN         = 107
	ID_SKIP_VERIFY    = 108
	ID_PROJECT_LIST   = 109
	ID_PROJECT_BTN    = 110
	ID_SETTINGS_BTN   = 111
	ID_OPEN_REQ_BTN   = 112
	ID_DELETE_REQ_BTN = 113
	ID_RECENT_LIST    = 114
)

// Global control handles - Request Panel
var (
	mainWindow     *win32.Window
	tabs           *win32.TabManager
	methodCombo    win32.HWND
	urlInput       win32.HWND
	headersInput   win32.HWND
	queryInput     win32.HWND
	bodyInput      win32.HWND
	responseOutput win32.HWND
	statusLabel    win32.HWND
	sendBtn        win32.HWND
	// Labels for request panel
	methodLabel   win32.HWND
	urlLabel      win32.HWND
	headersLabel  win32.HWND
	queryLabel    win32.HWND
	bodyLabel     win32.HWND
	responseLabel win32.HWND
)

// Global control handles - Settings Panel
var (
	certInput     win32.HWND
	keyInput      win32.HWND
	caInput       win32.HWND
	skipVerifyChk win32.HWND
	certBtn       win32.HWND
	keyBtn        win32.HWND
	caBtn         win32.HWND
	certLabel     win32.HWND
	keyLabel      win32.HWND
	caLabel       win32.HWND
	settingsTitle win32.HWND
)

// Global control handles - Project View Panel
var (
	projectListBox win32.HWND
	openReqBtn     win32.HWND
	deleteReqBtn   win32.HWND
	projectInfo    win32.HWND
	saveBtn        win32.HWND
)

// Global control handles - New Tab Panel
var (
	newTabTitle   win32.HWND
	newTabNewBtn  win32.HWND
	newTabOpenBtn win32.HWND
	recentLabel   win32.HWND
	recentListBox win32.HWND
)

// Current project and tab mapping
var (
	currentProject *Project
	recentProjects []string // list of recently opened project paths
)

// Layout state for resizing
var (
	windowWidth       int32   = 1100
	windowHeight      int32   = 850
	paramsHeightRatio float32 = 0.10 // 10% of available height for params/headers
	bodyHeightRatio   float32 = 0.15 // 15% for body
	// responseHeightRatio is calculated (remaining space)
)

// Splitter state for drag-to-resize
var (
	splitter1Y int32 // Y position of splitter between params and body
	splitter2Y int32 // Y position of splitter between body and response
)

func main() {
	runtime.LockOSThread()

	mainWindow = win32.NewWindow("REST Tester", 1100, 850)
	recentProjects = make([]string, 0)

	tabs = mainWindow.EnableTabs()
	// Handle tab events
	tabs.OnTabChanged = func(tabID int) {
		// Restore new tab state
		tabState := mainWindow.TabManager.GetActiveTab().Data
		if ts, ok := tabState.(*TabState); ok {
			RestoreTabState(ts)
		}

	}
	tabs.OnTabClosed = func(tabID int) {
		// If no tabs left, show new tab
		if tabs.GetTabCount() == 0 {
			CreateWelcomeTab()
		}
	}

	// Create all UI panels (no toolbar needed - menu is in tab bar now)
	createRequestPanel()
	createProjectViewPanel()
	createSettingsPanel()
	createNewTabPanel()

	// Initialize panel management
	initPanels()

	// Wire up the tab manager's menu button callback
	tabs.OnMenuClick = showContextMenu

	// Start with the Welcome Tab
	CreateWelcomeTab()

	// Handle button clicks
	mainWindow.OnCommand = handleCommand

	// Handle window resizing
	mainWindow.OnResize = handleResize

	mainWindow.Run()
}

func createRequestPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(12) // Start right after tab bar (tab bar height handled by control offset)

	// === Request Row ===
	methodLabel = mainWindow.CreateLabel("Method", padding, y+3, 50, labelHeight)
	methodCombo = mainWindow.CreateComboBox(padding+55, y, 90, 200, ID_METHOD_COMBO)
	win32.ComboBoxAddString(methodCombo, "GET")
	win32.ComboBoxAddString(methodCombo, "POST")
	win32.ComboBoxAddString(methodCombo, "PUT")
	win32.ComboBoxAddString(methodCombo, "PATCH")
	win32.ComboBoxAddString(methodCombo, "DELETE")
	win32.ComboBoxAddString(methodCombo, "HEAD")
	win32.ComboBoxAddString(methodCombo, "OPTIONS")
	win32.ComboBoxSetCurSel(methodCombo, 0)

	// URL input
	urlLabel = mainWindow.CreateLabel("URL", padding+155, y+3, 30, labelHeight)
	urlInput = mainWindow.CreateInput(padding+190, y, 680, inputHeight)

	// Send button
	sendBtn = mainWindow.CreateButton("  Send  ", padding+885, y, 90, inputHeight+2, ID_SEND_BTN)

	// === Query Parameters Section ===
	y += inputHeight + padding
	queryLabel = mainWindow.CreateLabel("Query Parameters (one per line: key=value)", padding, y, 300, labelHeight)
	y += labelHeight + 3
	queryInput = mainWindow.CreateCodeEdit(padding, y, 520, 70, false)

	// === Headers Section ===
	headersLabel = mainWindow.CreateLabel("Headers (one per line: Header: value)", padding+540, y-labelHeight-3, 300, labelHeight)
	headersInput = mainWindow.CreateCodeEdit(padding+540, y, 520, 70, false)

	// === Body Section ===
	y += 70 + padding
	bodyLabel = mainWindow.CreateLabel("Request Body", padding, y, 150, labelHeight)
	y += labelHeight + 3
	bodyInput = mainWindow.CreateCodeEdit(padding, y, 1050, 120, false)

	// === Response Section ===
	y += 120 + padding
	responseLabel = mainWindow.CreateLabel("Response", padding, y, 80, labelHeight)
	statusLabel = mainWindow.CreateLabel("Ready", padding+90, y, 400, labelHeight)
	y += labelHeight + 5
	responseOutput = mainWindow.CreateCodeEdit(padding, y, 1050, 350, true)
}

func createProjectViewPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(50) // After toolbar

	// Project info
	projectInfo = mainWindow.CreateLabel("Double-click a request to open it in a new tab", padding, y, 500, labelHeight)
	y += labelHeight + padding

	// List of requests
	projectListBox = mainWindow.CreateListBox(padding, y, 500, 500, ID_PROJECT_LIST)

	// Action buttons
	openReqBtn = mainWindow.CreateButton("Open in Tab", padding+520, y, 120, inputHeight, ID_OPEN_REQ_BTN)
	deleteReqBtn = mainWindow.CreateButton("Delete", padding+520, y+inputHeight+5, 120, inputHeight, ID_DELETE_REQ_BTN)
	saveBtn = mainWindow.CreateButton("Save Project", padding+520, y+inputHeight*2+15, 120, inputHeight, ID_SAVE_BTN)

	// Initially hide this panel
	win32.ShowWindow(projectInfo, win32.SW_HIDE)
	win32.ShowWindow(projectListBox, win32.SW_HIDE)
	win32.ShowWindow(openReqBtn, win32.SW_HIDE)
	win32.ShowWindow(deleteReqBtn, win32.SW_HIDE)
	win32.ShowWindow(saveBtn, win32.SW_HIDE)
}

func createSettingsPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	y := int32(50) // After toolbar

	// === Client Certificate Section ===
	settingsTitle = mainWindow.CreateLabel("Global Settings", padding, y, 200, labelHeight+4)
	y += labelHeight + padding

	certLabel = mainWindow.CreateLabel("Client Certificate (PEM)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	certInput = mainWindow.CreateInput(padding, y, 400, inputHeight)
	certBtn = mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_CERT_BTN)
	y += inputHeight + padding

	keyLabel = mainWindow.CreateLabel("Private Key (PEM)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	keyInput = mainWindow.CreateInput(padding, y, 400, inputHeight)
	keyBtn = mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_KEY_BTN)
	y += inputHeight + padding

	caLabel = mainWindow.CreateLabel("CA Bundle (optional)", padding, y, 200, labelHeight)
	y += labelHeight + 3
	caInput = mainWindow.CreateInput(padding, y, 400, inputHeight)
	caBtn = mainWindow.CreateButton("...", padding+405, y, 30, inputHeight, ID_CA_BTN)
	y += inputHeight + padding

	skipVerifyChk = mainWindow.CreateCheckbox("Skip TLS Verification (insecure)", padding, y, 280, inputHeight, ID_SKIP_VERIFY)

	// Initially hide this panel
	win32.ShowWindow(settingsTitle, win32.SW_HIDE)
	win32.ShowWindow(certLabel, win32.SW_HIDE)
	win32.ShowWindow(certInput, win32.SW_HIDE)
	win32.ShowWindow(certBtn, win32.SW_HIDE)
	win32.ShowWindow(keyLabel, win32.SW_HIDE)
	win32.ShowWindow(keyInput, win32.SW_HIDE)
	win32.ShowWindow(keyBtn, win32.SW_HIDE)
	win32.ShowWindow(caLabel, win32.SW_HIDE)
	win32.ShowWindow(caInput, win32.SW_HIDE)
	win32.ShowWindow(caBtn, win32.SW_HIDE)
	win32.ShowWindow(skipVerifyChk, win32.SW_HIDE)
}

func createNewTabPanel() {
	padding := int32(12)
	labelHeight := int32(24)
	inputHeight := int32(32)
	btnWidth := int32(150)
	y := int32(80) // After toolbar, centered area

	// Welcome title
	newTabTitle = mainWindow.CreateLabel("REST Tester - Start", padding, y, 400, labelHeight+10)
	y += labelHeight + 30

	// New and Open buttons
	newTabNewBtn = mainWindow.CreateButton("📄 New Project", padding, y, btnWidth, inputHeight, ID_NEW_BTN)
	newTabOpenBtn = mainWindow.CreateButton("📂 Open Project", padding+btnWidth+10, y, btnWidth, inputHeight, ID_OPEN_BTN)
	y += inputHeight + 30

	// Recent projects section
	recentLabel = mainWindow.CreateLabel("Recent Projects:", padding, y, 200, labelHeight)
	y += labelHeight + 5
	recentListBox = mainWindow.CreateListBox(padding, y, 400, 300, ID_RECENT_LIST)
}

// handleResize is called when the window is resized
func handleResize(width, height int32) {
	windowWidth = width
	windowHeight = height

	// Get current active tab
	if tabs != nil {
		resizeRequestPanel()
		// TODO: Add resize handlers for other panel types if needed
	}
}

// resizeRequestPanel adjusts all controls in the request panel based on window size
func resizeRequestPanel() {
	padding := int32(12)
	labelHeight := int32(20)
	inputHeight := int32(26)
	tabHeight := int32(46) // Tab bar height (matches TabManager.titleBarHeight)

	// Calculate available height (excluding tab bar and padding)
	availableHeight := windowHeight - tabHeight - padding*5 - labelHeight*5 - inputHeight

	// Calculate panel heights based on ratios
	paramsHeight := int32(float32(availableHeight) * paramsHeightRatio)
	if paramsHeight < 60 {
		paramsHeight = 60 // Minimum height
	}

	bodyHeight := int32(float32(availableHeight) * bodyHeightRatio)
	if bodyHeight < 80 {
		bodyHeight = 80 // Minimum height
	}

	responseHeight := availableHeight - paramsHeight - bodyHeight
	if responseHeight < 150 {
		responseHeight = 150 // Minimum height
	}

	// Calculate available width
	availableWidth := windowWidth - padding*2

	// Start position (relative to client area including tab bar offset)
	y := tabHeight + padding

	// === Request Row (fixed height) ===
	// Method combo
	win32.MoveWindow(methodCombo, padding+55, y, 90, 200, true)

	// URL input - expand with window width
	urlWidth := availableWidth - 330 // Leave space for method, label, and send button
	win32.MoveWindow(urlInput, padding+190, y, urlWidth, inputHeight, true)

	// Send button - move to right edge
	win32.MoveWindow(sendBtn, padding+190+urlWidth+15, y, 90, inputHeight+2, true)

	// === Query Parameters & Headers Section ===
	y += inputHeight + padding
	y += labelHeight + 3

	// Split width 50/50 for params and headers
	halfWidth := (availableWidth - padding) / 2

	// Query parameters (left half)
	win32.MoveWindow(queryInput, padding, y, halfWidth, paramsHeight, true)

	// Headers (right half)
	win32.MoveWindow(headersInput, padding+halfWidth+padding, y, halfWidth, paramsHeight, true)

	// Store splitter 1 position (between params and body)
	splitter1Y = y + paramsHeight

	// === Body Section ===
	y += paramsHeight + padding
	y += labelHeight + 3
	win32.MoveWindow(bodyInput, padding, y, availableWidth, bodyHeight, true)

	// Store splitter 2 position (between body and response)
	splitter2Y = y + bodyHeight

	// === Response Section ===
	y += bodyHeight + padding
	y += labelHeight + 5
	win32.MoveWindow(responseOutput, padding, y, availableWidth, responseHeight, true)
}

func handleCommand(id int) {
	switch id {
	case ID_SEND_BTN:
		saveCertificateConfig()
		go sendRequest()
	case ID_NEW_BTN:
		newProject()
	case ID_OPEN_BTN:
		openProject()
	case ID_SAVE_BTN:
		saveProject()
	case ID_CERT_BTN:
		if path, ok := win32.OpenFileDialog(mainWindow.Hwnd, "Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(certInput, path)
		}
	case ID_KEY_BTN:
		if path, ok := win32.OpenFileDialog(mainWindow.Hwnd, "Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(keyInput, path)
		}
	case ID_CA_BTN:
		if path, ok := win32.OpenFileDialog(mainWindow.Hwnd, "Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			win32.SetWindowText(caInput, path)
		}
	case ID_OPEN_REQ_BTN:
		openSelectedRequest()
	case ID_DELETE_REQ_BTN:
		deleteSelectedRequest()
	case ID_RECENT_LIST:
		openRecentProject()
	}
}

// Menu item IDs for context menu
const (
	MENU_SETTINGS    = 1001
	MENU_PROJECT     = 1002
	MENU_NEW_REQUEST = 1003
	MENU_ABOUT       = 1004
)

// showContextMenu displays the main context menu
func showContextMenu() {
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
	selected := menu.Show(mainWindow.Hwnd)

	switch selected {
	case MENU_SETTINGS:
		CreateSettingsTab()
	case MENU_PROJECT:
		CreateProjectViewTab()
	case MENU_NEW_REQUEST:
		addNewRequestTab()
	case MENU_ABOUT:
		win32.MessageBox(mainWindow.Hwnd, "REST Tester v1.0\nA modern REST API testing tool", "About", win32.MB_OK)
	}
}

// openRecentProject opens a project from the recent list
func openRecentProject() {
	idx := win32.ListBoxGetCurSel(recentListBox)
	if idx < 0 || idx >= len(recentProjects) {
		return
	}
	openProjectFromPath(recentProjects[idx])
}

// addNewRequestTab creates a new request tab
func addNewRequestTab() {
	// First ensure we have a project
	if currentProject == nil {
		currentProject = NewProject("Untitled Project")
	}

	// Create a new request
	req := NewRequest(fmt.Sprintf("Request %d", len(currentProject.Requests)+1))
	req.Headers["Content-Type"] = "application/json"
	req.Headers["Accept"] = "application/json"
	currentProject.AddRequest(req)

	// Create tab
	tabID := CreateRequestTab(req.Method+" "+req.Name, req.ID)
	tabs.SetActiveTab(tabID)
}

// addRecentProject adds a path to the recent projects list
func addRecentProject(path string) {
	// Remove if already exists
	for i, p := range recentProjects {
		if p == path {
			recentProjects = append(recentProjects[:i], recentProjects[i+1:]...)
			break
		}
	}
	// Add to front
	recentProjects = append([]string{path}, recentProjects...)
	// Keep max 10
	if len(recentProjects) > 10 {
		recentProjects = recentProjects[:10]
	}
}

// openSelectedRequest opens the selected request from project list in a new tab
func openSelectedRequest() {
	idx := getSelectedRequestIndex()
	if idx < 0 || idx >= len(currentProject.Requests) {
		return
	}

	req := currentProject.Requests[idx]

	// Open in new tab
	tabID := CreateRequestTab(req.Method+" "+req.Name, req.ID)
	tabs.SetActiveTab(tabID)
	// Manually restore state since OnTabChanged might not fire if this is the first/only tab
}

// deleteSelectedRequest removes the selected request from project
func deleteSelectedRequest() {
	idx := getSelectedRequestIndex()
	if idx < 0 || idx >= len(currentProject.Requests) {
		return
	}

	req := currentProject.Requests[idx]

	// Remove from project
	currentProject.RemoveRequest(req.ID)
	updateProjectList()
}

// newProject creates a new empty project
func newProject() {
	currentProject = NewProject("Untitled Project")
	// Open the project view tab
	CreateProjectViewTab()
}

// openProject opens a project from file dialog
func openProject() {
	filePath, ok := win32.OpenFileDialog(
		mainWindow.Hwnd,
		"Open Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
	)
	if !ok {
		return
	}
	openProjectFromPath(filePath)
}

// openProjectFromPath opens a project from a specific file path
func openProjectFromPath(filePath string) {
	project, err := LoadProject(filePath)
	if err != nil {
		win32.MessageBox(mainWindow.Hwnd, fmt.Sprintf("Error loading project: %v", err), "Error", win32.MB_OK)
		return
	}

	// Add to recent projects
	addRecentProject(filePath)

	currentProject = project

	// Open the project view tab
	CreateProjectViewTab()
	loadCertificateUI()
}

// saveProject saves the current project to file
func saveProject() {
	saveCertificateConfig()

	defaultName := currentProject.Name + ".rtp"
	filePath, ok := win32.SaveFileDialog(
		mainWindow.Hwnd,
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := currentProject.Save(filePath); err != nil {
		win32.MessageBox(mainWindow.Hwnd, fmt.Sprintf("Error saving project: %v", err), "Error", win32.MB_OK)
		return
	}

	// Update project name from filename
	name := filePath
	if idx := strings.LastIndex(name, "\\"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".rtp")
	currentProject.Name = name
}

// saveCertificateConfig saves the certificate settings from UI to project
func saveCertificateConfig() {
	if currentProject == nil {
		return
	}
	if currentProject.Certificate == nil {
		currentProject.Certificate = &CertificateConfig{}
	}
	currentProject.Certificate.CertFile = win32.GetWindowText(certInput)
	currentProject.Certificate.KeyFile = win32.GetWindowText(keyInput)
	currentProject.Certificate.CACertFile = win32.GetWindowText(caInput)
	currentProject.Certificate.SkipVerify = win32.CheckboxIsChecked(skipVerifyChk)
}

// loadCertificateUI loads certificate settings into the UI
func loadCertificateUI() {
	if currentProject == nil || currentProject.Certificate == nil {
		clearCertificateUI()
		return
	}
	win32.SetWindowText(certInput, currentProject.Certificate.CertFile)
	win32.SetWindowText(keyInput, currentProject.Certificate.KeyFile)
	win32.SetWindowText(caInput, currentProject.Certificate.CACertFile)
	win32.CheckboxSetChecked(skipVerifyChk, currentProject.Certificate.SkipVerify)
}

// clearCertificateUI clears all certificate input fields
func clearCertificateUI() {
	win32.SetWindowText(certInput, "")
	win32.SetWindowText(keyInput, "")
	win32.SetWindowText(caInput, "")
	win32.CheckboxSetChecked(skipVerifyChk, false)
}

func sendRequest() {
	// Get values from controls
	method := win32.ComboBoxGetText(methodCombo)
	url := win32.GetWindowText(urlInput)
	headersText := win32.GetWindowText(headersInput)
	queryText := win32.GetWindowText(queryInput)
	body := win32.GetWindowText(bodyInput)

	if url == "" {
		win32.SetWindowText(statusLabel, "⚠ Error - URL is empty")
		win32.SetWindowText(responseOutput, "Please enter a URL")
		return
	}

	// Append query parameters to URL
	url = buildURLWithQueryParams(url, queryText)

	win32.SetWindowText(statusLabel, "⏳ Sending...")
	win32.SetWindowText(responseOutput, "")

	// Create request
	var reqBody io.Reader
	if method == "POST" || method == "PUT" || method == "PATCH" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		win32.SetWindowText(statusLabel, "❌ Error")
		win32.SetWindowText(responseOutput, fmt.Sprintf("Error creating request:\r\n%v", err))
		return
	}

	// Parse and add headers
	headers := strings.Split(headersText, "\n")
	for _, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Create HTTP client with TLS configuration
	client, err := createHTTPClient()
	if err != nil {
		win32.SetWindowText(statusLabel, "❌ Certificate Error")
		win32.SetWindowText(responseOutput, fmt.Sprintf("Error loading certificates:\r\n%v", err))
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		win32.SetWindowText(statusLabel, "❌ Connection Error")
		win32.SetWindowText(responseOutput, fmt.Sprintf("Error sending request:\r\n%v", err))
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		win32.SetWindowText(statusLabel, fmt.Sprintf("⚠ %s", resp.Status))
		win32.SetWindowText(responseOutput, fmt.Sprintf("Error reading response:\r\n%v", err))
		return
	}

	// Determine status icon
	statusIcon := "✓"
	if resp.StatusCode >= 400 {
		statusIcon = "✗"
	}

	// Get content type to determine formatting
	contentType := resp.Header.Get("Content-Type")
	formattedBody := formatResponse(string(respBody), contentType)

	// Update UI
	win32.SetWindowText(statusLabel, fmt.Sprintf("%s %s", statusIcon, resp.Status))
	win32.SetWindowText(responseOutput, formattedBody)
}

// createHTTPClient creates an HTTP client with optional TLS client certificate
func createHTTPClient() (*http.Client, error) {
	client := &http.Client{}

	// Check if we have certificate configuration
	if currentProject == nil || currentProject.Certificate == nil {
		return client, nil
	}

	cert := currentProject.Certificate
	certFile := strings.TrimSpace(cert.CertFile)
	keyFile := strings.TrimSpace(cert.KeyFile)

	// No certificate configured
	if certFile == "" && keyFile == "" && !cert.SkipVerify {
		return client, nil
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{}

	// Load client certificate if provided
	if certFile != "" && keyFile != "" {
		clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	// Load CA certificate if provided
	if cert.CACertFile != "" {
		caCert, err := os.ReadFile(cert.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Skip TLS verification if requested
	tlsConfig.InsecureSkipVerify = cert.SkipVerify

	// Create transport with TLS config
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client.Transport = transport
	return client, nil
}

// formatResponse formats the response body based on content type
func formatResponse(body string, contentType string) string {
	contentType = strings.ToLower(contentType)

	// Try JSON formatting
	if strings.Contains(contentType, "json") || strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
		if formatted, ok := formatJSON(body); ok {
			return formatted
		}
	}

	// Try XML formatting
	if strings.Contains(contentType, "xml") || strings.HasPrefix(strings.TrimSpace(body), "<") {
		if formatted, ok := formatXML(body); ok {
			return formatted
		}
	}

	// Plain text - just return as-is with Windows line endings
	return strings.ReplaceAll(body, "\n", "\r\n")
}

// formatJSON pretty-prints JSON
func formatJSON(input string) (string, bool) {
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", false
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(data); err != nil {
		return "", false
	}

	// Convert to Windows line endings
	result := strings.TrimSuffix(buf.String(), "\n")
	return strings.ReplaceAll(result, "\n", "\r\n"), true
}

// formatXML pretty-prints XML
func formatXML(input string) (string, bool) {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(strings.NewReader(input))
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", false
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", false
	}

	// Convert to Windows line endings
	result := buf.String()
	return strings.ReplaceAll(result, "\n", "\r\n"), true
}

// buildURLWithQueryParams appends query parameters to a URL
func buildURLWithQueryParams(baseURL string, queryText string) string {
	if strings.TrimSpace(queryText) == "" {
		return baseURL
	}

	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}

	lines := strings.Split(queryText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Handle both key=value and key formats
		baseURL += separator + line
		separator = "&"
	}

	return baseURL
}
