package main

// TabType represents the type of content displayed in a tab
type TabType int

const (
	TabTypeRequest TabType = iota
	TabTypeProjectView
	TabTypeSettings
	TabTypeNewTab
)

// TabContent is the interface implemented by all tab content types
type TabContent interface {
	// TabType returns the type of this tab content
	TabType() TabType
}

// RequestTabContent holds state specific to request editing tabs
type RequestTabContent struct {
	BoundRequest *Request // Direct binding to the Request object
	Method       string
	URL          string
	Headers      string
	QueryParams  string
	Body         string
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
	CertFile   string
	KeyFile    string
	CACertFile string
	SkipVerify bool
}

func (s *SettingsTabContent) TabType() TabType { return TabTypeSettings }

// NewTabTabContent holds state specific to new tab screen
type NewTabTabContent struct {
	SelectedRecentIndex int // Currently selected recent project index
}

func (n *NewTabTabContent) TabType() TabType { return TabTypeNewTab }
