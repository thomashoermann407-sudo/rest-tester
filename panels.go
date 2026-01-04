package main

import (
	"hoermi.com/rest-test/win32"
)

const (
	// Common spacing
	layoutPadding = int32(12)

	// Standard control heights
	layoutColumnWidth     = int32(400)
	layoutLabelHeight     = int32(20)
	layoutInputHeight     = int32(26)
	layoutIconInputHeight = int32(32)
	layoutButtonWidth     = int32(120)
	layoutListHeight      = int32(300)

	PanelRequest     win32.PanelGroupName = "request"
	PanelProjectView win32.PanelGroupName = "projectView"
	PanelSettings    win32.PanelGroupName = "settings"
	PanelWelcome     win32.PanelGroupName = "welcome"
)
