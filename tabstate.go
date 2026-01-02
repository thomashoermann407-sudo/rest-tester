package main

// RequestTabContent holds state specific to request editing tabs
type RequestTabContent struct {
	BoundRequest *Request // Direct binding to the Request object
	BoundProject *Project // Reference to the project for settings
	Settings     *Settings
	Response     string
	Status       string
}

// ProjectViewTabContent holds state specific to project view tabs
type ProjectViewTabContent struct {
	BoundProject   *Project // Direct binding to the Project object
	SelectedIndex  int      // Currently selected request index in listbox
	ScrollPosition int      // Scroll position in the listbox
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
