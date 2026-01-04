package main

import (
	"fmt"
	"strings"
	"time"

	"hoermi.com/rest-test/win32"
)

var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

type requestPanelGroup struct {
	*win32.ControllerGroup
	methodCombo      *win32.ComboBoxControl
	envCombo         *win32.ComboBoxControl
	urlInput         *win32.Control
	headersInput     *win32.Control
	queryInput       *win32.Control
	bodyInput        *win32.Control
	responseTabCtrl  *win32.TabControlControl
	responseBody     *win32.Control
	responseHeaders  *win32.Control
	responseInfo     *win32.Control
	statusLabel      *win32.Control
	sendBtn          *win32.ButtonControl
	clearResponseBtn *win32.ButtonControl
	manageEnvBtn     *win32.ButtonControl
	methodLabel      *win32.Control
	envLabel         *win32.Control
	urlLabel         *win32.Control
	headersLabel     *win32.Control
	queryLabel       *win32.Control
	bodyLabel        *win32.Control
	responseLabel    *win32.Control

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
	envLabelWidth := int32(30)
	envComboWidth := int32(150)
	sendBtnWidth := int32(90)
	clearBtnWidth := int32(100)
	manageEnvBtnWidth := int32(90)

	// Position method label and combo
	r.methodLabel.MoveWindow(layoutPadding, y+3, methodLabelWidth, layoutLabelHeight)
	r.methodCombo.MoveWindow(layoutPadding+methodLabelWidth+layoutPadding, y, methodComboWidth, 200)

	// Position environment label and combo
	envX := layoutPadding + methodLabelWidth + layoutPadding + methodComboWidth + layoutPadding
	r.envLabel.MoveWindow(envX, y+3, envLabelWidth, layoutLabelHeight)
	r.envCombo.MoveWindow(envX+envLabelWidth+layoutPadding, y, envComboWidth, 200)

	// Position path label and input
	pathX := envX + envLabelWidth + layoutPadding + envComboWidth + layoutPadding
	r.urlLabel.MoveWindow(pathX, y+3, int32(30), layoutLabelHeight)

	pathInputX := pathX + 30 + layoutPadding
	pathWidth := availableWidth - methodLabelWidth - methodComboWidth - envLabelWidth - envComboWidth - 30 - sendBtnWidth - clearBtnWidth - manageEnvBtnWidth - layoutPadding*8
	r.urlInput.MoveWindow(pathInputX, y, pathWidth, layoutInputHeight)
	r.manageEnvBtn.MoveWindow(width-layoutPadding-sendBtnWidth-clearBtnWidth-manageEnvBtnWidth-layoutPadding*2, y, manageEnvBtnWidth, layoutInputHeight)
	r.sendBtn.MoveWindow(width-layoutPadding-sendBtnWidth-clearBtnWidth-layoutPadding, y, sendBtnWidth, layoutInputHeight)
	r.clearResponseBtn.MoveWindow(width-layoutPadding-clearBtnWidth, y, clearBtnWidth, layoutInputHeight)

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

	// === Response Section with TabControl ===
	y += bodyHeight + layoutPadding
	r.responseLabel.MoveWindow(layoutPadding, y, 80, layoutLabelHeight)
	r.statusLabel.MoveWindow(layoutPadding+90, y, 400, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding

	// TabControl for multiple responses
	tabCtrlHeight := int32(25)
	r.responseTabCtrl.MoveWindow(layoutPadding, y, availableWidth, responseHeight)

	// Content area below tabs
	contentY := y + tabCtrlHeight + layoutPadding
	contentHeight := responseHeight - tabCtrlHeight - layoutPadding*2

	// Split response content: info, body, and headers
	infoHeight := int32(25)
	remainingHeight := contentHeight - infoHeight - layoutPadding
	bodyHeadersHeight := remainingHeight / 2

	r.responseInfo.MoveWindow(layoutPadding*2, contentY, availableWidth-layoutPadding*2, infoHeight)
	r.responseBody.MoveWindow(layoutPadding*2, contentY+infoHeight+layoutPadding, availableWidth-layoutPadding*2, bodyHeadersHeight)
	r.responseHeaders.MoveWindow(layoutPadding*2, contentY+infoHeight+layoutPadding+bodyHeadersHeight+layoutPadding, availableWidth-layoutPadding*2, bodyHeadersHeight)
}

func (r *requestPanelGroup) SaveState() {
	req := r.content.BoundRequest
	req.Method = r.methodCombo.GetText()
	req.Host = r.envCombo.GetText()
	req.Path = r.urlInput.GetText()
	req.Body = r.bodyInput.GetText()
	req.Headers = ParseParams(r.headersInput.GetText())
	req.QueryParams = ParseParams(r.queryInput.GetText())
	// Responses are managed separately, no need to save here
}

func (r *requestPanelGroup) SetState(data any) {
	if content, ok := data.(*RequestTabContent); ok {
		r.content = content
		req := content.BoundRequest

		// Set method
		for i, m := range httpMethods {
			if m == req.Method {
				r.methodCombo.SetCurSel(i)
				break
			}
		}

		// Populate environment dropdown
		r.envCombo.Clear()
		if content.BoundProject != nil {
			for _, env := range content.BoundProject.Environments {
				r.envCombo.AddString(env.Name)
			}
			if len(content.BoundProject.Environments) > 0 {
				r.envCombo.SetCurSel(0) // Select first environment by default
			}
		}

		r.urlInput.SetText(req.Path)
		r.headersInput.SetText(req.Headers.Format())
		r.queryInput.SetText(req.QueryParams.Format())
		r.bodyInput.SetText(req.Body)

		// Update response tabs
		r.updateResponseTabs()
	}
}

// updateResponseTabs rebuilds the response tabs from the content
func (r *requestPanelGroup) updateResponseTabs() {
	r.responseTabCtrl.DeleteAllItems()

	if len(r.content.Responses) == 0 {
		r.responseInfo.SetText("No responses yet. Click 'Send' to make a request.")
		r.responseBody.SetText("")
		r.responseHeaders.SetText("")
		r.statusLabel.SetText("Ready")
		return
	}

	// Add tab for each response (newest first)
	for i, resp := range r.content.Responses {
		tabName := fmt.Sprintf("#%d - %s", len(r.content.Responses)-i, resp.Status)
		r.responseTabCtrl.InsertItem(i, tabName, uintptr(i))
	}

	// Select the first (newest) tab
	r.responseTabCtrl.SetCurSel(0)
	r.displayResponse(0)
}

// displayResponse shows a specific response by index
func (r *requestPanelGroup) displayResponse(index int) {
	if index < 0 || index >= len(r.content.Responses) {
		return
	}

	resp := r.content.Responses[index]

	// Update info label
	infoText := fmt.Sprintf("Duration: %v | Time: %s",
		resp.Duration.Round(1000), // Round to microseconds
		resp.Timestamp.Format("15:04:05"))
	r.responseInfo.SetText(infoText)
	r.statusLabel.SetText(fmt.Sprintf("✅ %s", resp.Status))

	// Update body
	r.responseBody.SetText(resp.Body)

	// Format headers for display
	var headerLines []string
	for name, value := range resp.Headers {
		headerLines = append(headerLines, fmt.Sprintf("%s: %s", name, value))
	}
	r.responseHeaders.SetText(strings.Join(headerLines, "\r\n"))
}

func createRequestPanel(factory win32.ControlFactory) *requestPanelGroup {
	group := &requestPanelGroup{
		methodLabel:     factory.CreateLabel("Method"),
		methodCombo:     factory.CreateComboBox(),
		envLabel:        factory.CreateLabel("Env"),
		envCombo:        factory.CreateEditableComboBox(),
		urlLabel:        factory.CreateLabel("Path"),
		urlInput:        factory.CreateInput(),
		queryLabel:      factory.CreateLabel("Query Parameters (one per line: key=value)"),
		queryInput:      factory.CreateCodeEdit(false),
		headersLabel:    factory.CreateLabel("Headers (one per line: Header: value)"),
		headersInput:    factory.CreateCodeEdit(false),
		bodyLabel:       factory.CreateLabel("Request Body"),
		bodyInput:       factory.CreateCodeEdit(false),
		responseLabel:   factory.CreateLabel("Response"),
		statusLabel:     factory.CreateLabel("Ready"),
		responseTabCtrl: factory.CreateTabControl(),
		responseInfo:    factory.CreateLabel(""),
		responseBody:    factory.CreateCodeEdit(true),
		responseHeaders: factory.CreateCodeEdit(true),
	}

	// Set up tab change handler
	group.responseTabCtrl.SetOnSelChange(func(tc *win32.TabControlControl) {
		selectedIndex := tc.GetCurSel()
		group.displayResponse(selectedIndex)
	})

	group.clearResponseBtn = factory.CreateButton("Clear", func() {
		if group.content != nil {
			group.content.Responses = nil
			group.updateResponseTabs()
		}
	})

	group.manageEnvBtn = factory.CreateButton("Manage...", func() {
		// TODO: Open environment management dialog
		factory.MessageBox("Environment management dialog will be implemented here.", "Environment Management")
	})

	group.sendBtn = factory.CreateButton("Send", func() {
		// Get the bound request from the current tab
		group.SaveState()
		request := group.content.BoundRequest
		if request == nil {
			group.statusLabel.SetText("❌ No request")
			group.responseBody.SetText("Error: No request bound to this tab")
			return
		}

		// Get environment base URL (either from selected item or manually entered text)
		var baseURL string
		envIndex := group.envCombo.GetCurSel()

		if envIndex >= 0 && group.content.BoundProject != nil && envIndex < len(group.content.BoundProject.Environments) {
			// Use selected environment
			env := group.content.BoundProject.Environments[envIndex]
			baseURL = env.BaseURL
		} else {
			// Use manually entered text
			baseURL = group.envCombo.GetText()
		}

		if baseURL == "" {
			group.statusLabel.SetText("❌ No environment")
			group.responseBody.SetText("Error: No environment or base URL specified")
			return
		}

		group.statusLabel.SetText("⏳ Sending...")
		group.responseBody.SetText("")

		// Get timeout from project settings (default 30000ms if not set)
		timeoutInMs := int64(30000)
		if group.content.BoundProject != nil && group.content.BoundProject.Settings.TimeoutInMs > 0 {
			timeoutInMs = group.content.BoundProject.Settings.TimeoutInMs
		}

		// Send request in background goroutine
		go request.sendRequest(group.content.Settings, timeoutInMs, func(responseData *ResponseData, err error) {
			// Marshal the UI update back to the main thread using PostUICallback
			factory.PostUICallback(func() {
				if err != nil {
					// Create error response data
					errorResponse := ResponseData{
						Body:       fmt.Sprintf("Error sending request:\r\n%v", err),
						Headers:    make(map[string]string),
						StatusCode: 0,
						Status:     "Error",
						Duration:   0,
						Timestamp:  time.Now(),
					}
					group.statusLabel.SetText("❌ Error")

					// Add error response to the list
					group.content.Responses = append([]ResponseData{errorResponse}, group.content.Responses...)
					group.updateResponseTabs()
					return
				}

				// Add new response to the beginning of the list (newest first)
				group.content.Responses = append([]ResponseData{*responseData}, group.content.Responses...)

				// Update the response tabs
				group.updateResponseTabs()
			})
		})
	})

	for _, method := range httpMethods {
		group.methodCombo.AddString(method)
	}
	group.methodCombo.SetCurSel(0)

	group.ControllerGroup = win32.NewControllerGroup(
		group.methodCombo, group.envCombo, group.urlInput, group.headersInput, group.queryInput, group.bodyInput,
		group.responseBody, group.responseHeaders, group.responseInfo, group.responseTabCtrl,
		group.statusLabel, group.sendBtn, group.clearResponseBtn, group.manageEnvBtn,
		group.methodLabel, group.envLabel, group.urlLabel, group.headersLabel, group.queryLabel, group.bodyLabel, group.responseLabel,
	)
	return group
}
