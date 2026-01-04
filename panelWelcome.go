package main

import (
	"path/filepath"

	"hoermi.com/rest-test/win32"
)

type welcomePanelGroup struct {
	*win32.ControllerGroup
	// Welcome Panel controls
	newTabTitle   *win32.Control
	newTabNewBtn  *win32.ButtonControl
	newTabOpenBtn *win32.ButtonControl
	recentLabel   *win32.Control
	recentListBox *win32.ListBoxControl

	content *WelcomeTabContent
}

func (w *welcomePanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight + layoutPadding

	w.newTabTitle.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding

	w.newTabNewBtn.MoveWindow(layoutPadding, y, layoutButtonWidth, layoutIconInputHeight)
	w.newTabOpenBtn.MoveWindow(layoutPadding+layoutButtonWidth, y, layoutButtonWidth, layoutIconInputHeight)
	y += layoutIconInputHeight + layoutPadding

	w.recentLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	w.recentListBox.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutListHeight)
}

func (w *welcomePanelGroup) SaveState() {
	// Capture selected recent project index
	w.content.SelectedRecentIndex = w.recentListBox.GetCurSel()
}

func (w *welcomePanelGroup) SetState(data any) {
	if content, ok := data.(*WelcomeTabContent); ok {
		w.content = content
		w.recentListBox.ResetContent()
		for _, path := range content.RecentProjects {
			w.recentListBox.AddString(filepath.Base(path))
		}
	}
}

func createWelcomePanel(factory win32.ControlFactory, projectManager ProjectManager) *welcomePanelGroup {
	group := &welcomePanelGroup{
		newTabTitle:   factory.CreateLabel("REST Tester - Start"),
		newTabNewBtn:  factory.CreateButton("📄 New Project", func() { projectManager.newProject() }),
		newTabOpenBtn: factory.CreateButton("📂 Open Project", func() { projectManager.openProject() }),
		recentLabel:   factory.CreateLabel("Recent Projects:"),
	}
	group.recentListBox = factory.CreateListBox(func(list *win32.ListBoxControl) {
		idx := list.GetCurSel()
		if idx < 0 || idx >= len(group.content.RecentProjects) {
			return
		}
		projectManager.openProjectFromPath(group.content.RecentProjects[idx])
	})
	group.ControllerGroup = win32.NewControllerGroup(
		group.newTabTitle, group.newTabNewBtn, group.newTabOpenBtn,
		group.recentLabel, group.recentListBox,
	)
	return group
}
