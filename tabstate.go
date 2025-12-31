package main

// TabType represents the type of content displayed in a tab
type TabType int

const (
	TabTypeRequest TabType = iota
	TabTypeProjectView
	TabTypeSettings
	TabTypeWelcome
)

// TabContent is the interface implemented by all tab content types
type TabContent interface {
	// TabType returns the type of this tab content
	TabType() TabType
}

// RequestTabContent holds state specific to request editing tabs
type RequestTabContent struct {
	BoundRequest *Request // Direct binding to the Request object
	Settings     *Settings
	Response     string
	Status       string
}

func (r *RequestTabContent) TabType() TabType { return TabTypeRequest }

// ProjectViewTabContent holds state specific to project view tabs
type ProjectViewTabContent struct {
	BoundProject   *Project // Direct binding to the Project object
	SelectedIndex  int      // Currently selected request index in listbox
	ScrollPosition int      // Scroll position in the listbox
}

func (p *ProjectViewTabContent) TabType() TabType { return TabTypeProjectView }

// SettingsTabContent holds state specific to settings tabs
type SettingsTabContent struct {
	Certificate *CertificateConfig
}

func (s *SettingsTabContent) TabType() TabType { return TabTypeSettings }

// WelcomeTabContent holds state specific to new tab screen
type WelcomeTabContent struct {
	RecentProjects      []string // List of recent project paths
	SelectedRecentIndex int      // Currently selected recent project index
}

func (n *WelcomeTabContent) TabType() TabType { return TabTypeWelcome }
