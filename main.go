package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"runtime"
	"strings"
)

func main() {
	runtime.LockOSThread()

	pw := NewProjectWindow()
	pw.settings = InitSettings()
	tabs := pw.tabs
	// Handle tab events
	tabs.OnBeforeTabChange = func() {
		// Save state of the tab we're leaving
		pw.SaveCurrentTabState()
	}
	tabs.OnTabChanged = func() {
		// Restore new tab state
		tabData := tabs.GetActiveTab().Data
		pw.RestoreTabState(tabData)
	}
	tabs.OnTabClosed = func() {
		// If no tabs left, show new tab
		if tabs.GetTabCount() == 0 {
			pw.CreateWelcomeTab()
		}
	}

	// Initialize panel management
	pw.panels = initPanels(pw)

	// Wire up the tab manager's menu button callback
	tabs.OnMenuClick = pw.showContextMenu

	// Start with the Welcome Tab
	pw.CreateWelcomeTab()
	// Manually restore state for the first tab since OnTabChanged won't fire
	if activeTab := tabs.GetActiveTab(); activeTab != nil {
		pw.RestoreTabState(activeTab.Data)
	}

	// Handle button clicks
	pw.mainWindow.OnCommand = pw.panels.handleCommand

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

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", false
	}

	// Convert to Windows line endings
	return strings.ReplaceAll(string(formatted), "\n", "\r\n"), true
}

// formatXML pretty-prints XML
func formatXML(input string) (string, bool) {
	var buf strings.Builder
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
	return strings.ReplaceAll(buf.String(), "\n", "\r\n"), true
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
