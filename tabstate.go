package main

import "time"

// ResponseData holds information about a single HTTP response
type ResponseData struct {
	Body       string            // Response body
	Headers    map[string]string // Response headers
	StatusCode int               // HTTP status code
	Status     string            // Status text (e.g., "200 OK")
	Duration   time.Duration     // Time taken for the request
	Timestamp  time.Time         // When the response was received
}

// RequestTabContent holds state specific to request editing tabs
type RequestTabContent struct {
	BoundRequest *Request // Direct binding to the Request object
	BoundProject *Project // Reference to the project for settings
	Settings     *Settings
	Responses    []ResponseData // Multiple responses (newest first)
}

// TreeNodeInfo stores metadata about a tree item
type TreeNodeInfo struct {
	Type     NodeType
	Segment  string
	Method   string   // Only for NodeTypeMethod
	Request  *Request // Only for NodeTypeMethod
	FullPath string   // Full URL path up to this node
}

// ProjectViewTabContent holds state specific to project view tabs
type ProjectViewTabContent struct {
	BoundProject   *Project // Direct binding to the Project object
	SelectedIndex  int      // Currently selected request index in listbox
	ScrollPosition int      // Scroll position in the listbox
	itemToNodeInfo map[uintptr]*TreeNodeInfo
}

// SettingsTabContent holds state specific to settings tabs
type SettingsTabContent struct {
	Settings *Settings
}

// WelcomeTabContent holds state specific to new tab screen
type WelcomeTabContent struct {
	RecentProjects      []string // List of recent project paths
	SelectedRecentIndex int      // Currently selected recent project index
}
