package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"runtime"
	"strings"
)

// Control IDs
// Todo: make ids dependent on window
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

func main() {
	runtime.LockOSThread()

	InitGlobalSettings()

	pw := NewProjectWindow()
	tabs := pw.tabs
	// Handle tab events
	tabs.OnBeforeTabChange = func(oldTabID int) {
		// Save state of the tab we're leaving
		pw.SaveCurrentTabState()
	}
	tabs.OnTabChanged = func(tabID int) {
		// Restore new tab state
		tabData := tabs.GetActiveTab().Data
		if content, ok := tabData.(TabContent); ok {
			pw.RestoreTabState(content)
		}

	}
	tabs.OnTabClosed = func(tabID int) {
		// If no tabs left, show new tab
		if tabs.GetTabCount() == 0 {
			pw.CreateNewTabTab()
		}
	}

	// Create all UI panels (no toolbar needed - menu is in tab bar now)
	pw.createRequestPanel()
	pw.createProjectViewPanel()
	pw.createSettingsPanel()
	pw.createNewTabPanel()

	// Initialize panel management
	initPanels(pw)

	// Wire up the tab manager's menu button callback
	tabs.OnMenuClick = pw.showContextMenu

	// Start with the Welcome Tab
	pw.CreateNewTabTab()
	// Manually restore state for the first tab since OnTabChanged won't fire
	if activeTab := tabs.GetActiveTab(); activeTab != nil {
		if content, ok := activeTab.Data.(TabContent); ok {
			pw.RestoreTabState(content)
		}
	}

	// Handle button clicks
	pw.mainWindow.OnCommand = pw.handleCommand

	// Handle window resizing
	pw.mainWindow.OnResize = pw.handleResize

	pw.mainWindow.Run()
}

// Menu item IDs for context menu
const (
	MENU_SETTINGS    = 1001
	MENU_PROJECT     = 1002
	MENU_NEW_REQUEST = 1003
	MENU_ABOUT       = 1004
)

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
	var data any
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

	lines := strings.SplitSeq(queryText, "\n")
	for line := range lines {
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
