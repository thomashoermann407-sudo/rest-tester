package main

import "hoermi.com/rest-test/win32"

// TabType represents the type of content displayed in a tab
type TabType int

const (
	TabTypeRequest TabType = iota
	TabTypeProjectView
	TabTypeSettings
	TabTypeNewTab
)

// RequestTabContent holds state specific to request editing tabs
type RequestTabContent struct {
	RequestID   string
	Method      string
	URL         string
	Headers     string
	QueryParams string
	Body        string
	Response    string
	Status      string
}

// ProjectViewTabContent holds state specific to project view tabs
type ProjectViewTabContent struct {
	SelectedIndex  int // Currently selected request index in listbox
	ScrollPosition int // Scroll position in the listbox
}

// SettingsTabContent holds state specific to settings tabs
type SettingsTabContent struct {
	CertFile   string
	KeyFile    string
	CACertFile string
	SkipVerify bool
}

// NewTabTabContent holds state specific to new tab screen
type NewTabTabContent struct {
	SelectedRecentIndex int // Currently selected recent project index
}

// TabState holds the local state for each tab with type-specific content
type TabState struct {
	Type TabType

	// Type-specific content (only one should be non-nil)
	RequestContent     *RequestTabContent
	ProjectViewContent *ProjectViewTabContent
	SettingsContent    *SettingsTabContent
	NewTabContent      *NewTabTabContent
}

// CreateRequestTab creates a new request tab
func CreateRequestTab(name string, requestID string) int {
	tabState := &TabState{
		Type: TabTypeRequest,
		RequestContent: &RequestTabContent{
			RequestID: requestID,
			Method:    "GET",
			Headers:   "Content-Type: application/json\r\nAccept: application/json",
			Status:    "Ready",
		},
	}
	tabID := tabs.AddTab(name, tabState)
	return tabID
}

// CreateProjectViewTab creates the project structure view tab
func CreateProjectViewTab() int {

	tabID := tabs.AddTab("📁 Project", nil)
	tabs.SetActiveTab(tabID)
	showProjectViewPanel()
	return tabID
}

// CreateSettingsTab creates the global settings tab
func CreateSettingsTab() int {
	tabID := tabs.AddTab("⚙ Settings", nil)
	tabs.SetActiveTab(tabID)
	showSettingsPanel()
	return tabID
}

// CreateNewTabTab creates the "New Tab" start tab
func CreateWelcomeTab() int {
	tabID := tabs.AddTab("Welcome", nil)
	showWelcomePanel()
	return tabID
}

// RestoreTabState restores UI state from a tab's saved state
func RestoreTabState(state *TabState) {
	if state == nil {
		return
	}

	switch state.Type {
	case TabTypeRequest:
		showRequestPanel()
		restoreRequestState(state)
		resizeRequestPanel() // Apply current window size to controls
	case TabTypeProjectView:
		showProjectViewPanel()
		restoreProjectViewState(state)
	case TabTypeSettings:
		showSettingsPanel()
		restoreSettingsState(state)
	case TabTypeNewTab:
		showWelcomePanel()
	}
}

// restoreRequestState restores request tab state to UI
func restoreRequestState(state *TabState) {
	if state.RequestContent == nil {
		return
	}

	content := state.RequestContent

	// Set method
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for i, m := range methods {
		if m == content.Method {
			win32.ComboBoxSetCurSel(methodCombo, i)
			break
		}
	}

	win32.SetWindowText(urlInput, content.URL)
	win32.SetWindowText(headersInput, content.Headers)
	win32.SetWindowText(queryInput, content.QueryParams)
	win32.SetWindowText(bodyInput, content.Body)
	win32.SetWindowText(responseOutput, content.Response)
	win32.SetWindowText(statusLabel, content.Status)
}

// restoreProjectViewState restores project view tab state to UI
func restoreProjectViewState(state *TabState) {
	if state.ProjectViewContent == nil {
		return
	}

	content := state.ProjectViewContent

	// Restore listbox selection
	if content.SelectedIndex >= 0 {
		win32.ListBoxSetCurSel(projectListBox, content.SelectedIndex)
	}
}

// restoreSettingsState restores settings tab state to UI
func restoreSettingsState(state *TabState) {
	if state.SettingsContent == nil {
		return
	}

	content := state.SettingsContent

	win32.SetWindowText(certInput, content.CertFile)
	win32.SetWindowText(keyInput, content.KeyFile)
	win32.SetWindowText(caInput, content.CACertFile)
	win32.CheckboxSetChecked(skipVerifyChk, content.SkipVerify)
}
