package main

import (
	"fmt"

	"hoermi.com/rest-test/win32"
)

type projectViewPanelGroup struct {
	*win32.ControllerGroup
	// Environment ListView controls
	envLabel      *win32.Control
	envListView   *win32.ListViewControl
	addEnvBtn     *win32.ButtonControl
	deleteEnvBtn  *win32.ButtonControl
	setDefaultBtn *win32.ButtonControl

	// Project View Panel controls
	projectTreeView *win32.TreeViewControl
	openReqBtn      *win32.ButtonControl
	deleteReqBtn    *win32.ButtonControl
	addPathBtn      *win32.ButtonControl
	addRequestBtn   *win32.ButtonControl
	projectInfo     *win32.Control
	saveBtn         *win32.ButtonControl
	timeoutLabel    *win32.Control
	timeoutInput    *win32.Control

	content        *ProjectViewTabContent
	tabController  TabController
	controlFactory win32.ControlFactory
}

// Context menu IDs
const (
	menuIDAddRequest = iota + 1000
	menuIDDelete
	menuIDEdit
)

func (p *projectViewPanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight + layoutPadding
	dy := layoutLabelHeight + layoutPadding
	btnX := layoutPadding + layoutColumnWidth + layoutPadding

	// Environment ListView section
	p.envLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding/2
	envListHeight := int32(100) // Height for environment list
	p.envListView.MoveWindow(layoutPadding, y, layoutColumnWidth, envListHeight)

	// Environment buttons next to ListView
	btnY := y
	p.addEnvBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.deleteEnvBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.setDefaultBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)

	// Move to next section
	y += envListHeight + layoutPadding

	// Project tree section
	p.projectInfo.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	p.projectTreeView.MoveWindow(layoutPadding, y+dy, layoutColumnWidth, layoutListHeight)

	// Right side buttons
	btnY = y + dy
	p.openReqBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.deleteReqBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.addPathBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.addRequestBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.saveBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)

	// Timeout settings below the tree
	y += dy + layoutListHeight + layoutPadding
	p.timeoutLabel.MoveWindow(layoutPadding, y, int32(200), layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding/2
	p.timeoutInput.MoveWindow(layoutPadding, y, int32(150), layoutInputHeight)
}

func (p *projectViewPanelGroup) SaveState() {
	// Save timeout setting
	timeoutText := p.timeoutInput.GetText()
	if timeoutText != "" {
		var timeout int64
		fmt.Sscanf(timeoutText, "%d", &timeout)
		if timeout > 0 {
			p.content.BoundProject.Settings.TimeoutInMs = timeout
		}
	}

	// Capture tree view expansion/selection state
	p.captureTreeState()
}

// populateEnvironmentList fills the ListView with environments
func (p *projectViewPanelGroup) populateEnvironmentList() {
	if p.envListView == nil || p.content == nil || p.content.BoundProject == nil {
		return
	}

	p.envListView.DeleteAllItems()

	defaultIdx := p.content.BoundProject.Settings.DefaultEnvironmentIdx
	for i, env := range p.content.BoundProject.Environments {
		name := env.Name
		if name == "" {
			name = "(Unnamed)"
		}

		// Add default marker
		defaultMarker := ""
		if i == defaultIdx {
			defaultMarker = "*"
		}

		p.envListView.InsertItem(i, defaultMarker, uintptr(i))
		p.envListView.SetItemText(i, 1, name)
		p.envListView.SetItemText(i, 2, env.BaseURL)
	}
}

// getSelectedEnvironmentIndex returns the index of the selected environment
func (p *projectViewPanelGroup) getSelectedEnvironmentIndex() int {
	return p.envListView.GetSelectedIndex()
}

// addEnvironment adds a new environment and starts editing it
func (p *projectViewPanelGroup) addEnvironment() {
	if p.content == nil || p.content.BoundProject == nil {
		return
	}

	// Create a new environment with default values
	newEnv := Environment{
		Name:    "",
		BaseURL: "http://localhost:8080",
	}

	// Add to the project
	p.content.BoundProject.Environments = append(p.content.BoundProject.Environments, newEnv)

	// Refresh the list to show the new environment
	p.populateEnvironmentList()

	// Get the index of the newly added environment
	newIndex := len(p.content.BoundProject.Environments) - 1

	// Select the new item in the ListView
	p.envListView.SetCurSel(newIndex)

	// Start editing the name field (column 1) so user can immediately type
	p.envListView.StartEdit(newIndex, 1)
}

// deleteEnvironment deletes the selected environment
func (p *projectViewPanelGroup) deleteEnvironment() {
	idx := p.getSelectedEnvironmentIndex()
	if idx < 0 || idx >= len(p.content.BoundProject.Environments) {
		p.controlFactory.MessageBox("Delete Environment", "Please select an environment to delete.")
		return
	}

	// Confirm deletion
	env := p.content.BoundProject.Environments[idx]
	name := env.Name
	if name == "" {
		name = env.BaseURL
	}
	if p.controlFactory.MessageBoxYesNo("Delete Environment", fmt.Sprintf("Are you sure you want to delete the environment '%s'?", name)) != win32.ID_YES {
		return
	}

	// Remove from slice
	p.content.BoundProject.Environments = append(
		p.content.BoundProject.Environments[:idx],
		p.content.BoundProject.Environments[idx+1:]...)

	// Adjust default environment index if needed
	if p.content.BoundProject.Settings.DefaultEnvironmentIdx == idx {
		p.content.BoundProject.Settings.DefaultEnvironmentIdx = -1
	} else if p.content.BoundProject.Settings.DefaultEnvironmentIdx > idx {
		p.content.BoundProject.Settings.DefaultEnvironmentIdx--
	}

	p.populateEnvironmentList()
}

// setDefaultEnvironment marks the selected environment as default
func (p *projectViewPanelGroup) setDefaultEnvironment() {
	idx := p.getSelectedEnvironmentIndex()
	if idx < 0 || idx >= len(p.content.BoundProject.Environments) {
		p.controlFactory.MessageBox("Set Default", "Please select an environment to set as default.")
		return
	}

	p.content.BoundProject.Settings.DefaultEnvironmentIdx = idx
	p.populateEnvironmentList()
}

// populateTreeNode recursively populates the tree view from a request node
func (p *projectViewPanelGroup) populateTreeNode(parentHandle uintptr, node *RequestNode, pathPrefix string) {
	if node == nil {
		return
	}

	// Build the current path
	currentPath := pathPrefix
	var segmentHandle uintptr

	if node.Segment != "" {
		// Non-root node: create a tree item for this segment
		if currentPath != "" && currentPath != "/" {
			currentPath += "/"
		}
		currentPath += node.Segment

		// Insert a node for this path segment
		segmentHandle = p.projectTreeView.InsertItem(parentHandle, win32.TVI_LAST, node.Segment, 0)

		// Store node info
		p.content.itemToNodeInfo[segmentHandle] = &TreeNodeInfo{
			Type:     NodeTypePath,
			Segment:  node.Segment,
			FullPath: currentPath,
		}
	} else {
		// Root node: don't create a tree item, just use parent handle
		segmentHandle = parentHandle
	}

	// Add request items for each HTTP method at this node
	for _, req := range node.Requests {
		displayText := fmt.Sprintf("[%s] %s", req.Method, req.Name)
		itemHandle := p.projectTreeView.InsertItem(segmentHandle, win32.TVI_LAST, displayText, 0)
		// Map the tree item to the request with node info
		p.content.itemToNodeInfo[itemHandle] = &TreeNodeInfo{
			Type:     NodeTypeRequest,
			Method:   req.Method,
			Request:  req,
			FullPath: currentPath,
		}
	}

	// Recursively add children
	for _, child := range node.Children {
		p.populateTreeNode(segmentHandle, child, currentPath)
	}
}

// captureTreeState stores expanded nodes and selection into tab content
func (p *projectViewPanelGroup) captureTreeState() {
	if p.projectTreeView == nil || p.content == nil || p.content.itemToNodeInfo == nil {
		return
	}

	var expanded []string

	// Record selection
	if sel := p.projectTreeView.GetSelection(); sel != 0 {
		if info := p.content.itemToNodeInfo[sel]; info != nil {
			p.content.SelectedPath = info.FullPath
		}
	}

	var walk func(h uintptr)
	walk = func(h uintptr) {
		if h == 0 {
			return
		}
		if info := p.content.itemToNodeInfo[h]; info != nil && p.projectTreeView.IsExpanded(h) {
			expanded = append(expanded, info.FullPath)
		}
		for child := p.projectTreeView.GetChild(h); child != 0; child = p.projectTreeView.GetNextSibling(child) {
			walk(child)
		}
	}

	for root := p.projectTreeView.GetRoot(); root != 0; root = p.projectTreeView.GetNextSibling(root) {
		walk(root)
	}

	p.content.ExpandedPaths = expanded
}

// restoreTreeState re-applies expansion and selection after repopulating tree
func (p *projectViewPanelGroup) restoreTreeState() {
	if p.projectTreeView == nil || p.content == nil || p.content.itemToNodeInfo == nil {
		return
	}

	pathToHandle := make(map[string]uintptr)
	for handle, info := range p.content.itemToNodeInfo {
		if info != nil {
			pathToHandle[info.FullPath] = handle
		}
	}

	for _, path := range p.content.ExpandedPaths {
		if h, ok := pathToHandle[path]; ok {
			p.projectTreeView.ExpandItem(h, true)
		}
	}

	if sel := p.content.SelectedPath; sel != "" {
		if h, ok := pathToHandle[sel]; ok {
			p.projectTreeView.SelectItem(h)
		}
	}
}

func (p *projectViewPanelGroup) SetState(data any) {
	if content, ok := data.(*ProjectViewTabContent); ok {
		p.content = content
		if p.projectTreeView == nil || content.BoundProject == nil {
			return
		}

		// Populate the environment list
		p.populateEnvironmentList()

		// Remember current tree state before rebuild
		p.captureTreeState()

		// Clear the tree view
		p.projectTreeView.DeleteAllItems()
		p.content.itemToNodeInfo = make(map[uintptr]*TreeNodeInfo)

		// Populate the tree from the root (start with empty path for root node)
		p.populateTreeNode(win32.TVI_ROOT, content.BoundProject.Tree, "")
		p.restoreTreeState()

		// Set timeout value
		timeout := content.BoundProject.Settings.TimeoutInMs
		if timeout == 0 {
			timeout = 30000 // Default 30 seconds
		}
		p.timeoutInput.SetText(fmt.Sprintf("%d", timeout))
	}
}

// getSelectedRequest returns the request associated with the selected tree item
func (p *projectViewPanelGroup) getSelectedRequest() (*Request, string) {
	itemHandle := p.projectTreeView.GetSelection()
	if itemHandle == 0 {
		return nil, ""
	}
	nodeInfo := p.content.itemToNodeInfo[itemHandle]
	if nodeInfo != nil && nodeInfo.Type == NodeTypeRequest {
		return nodeInfo.Request, nodeInfo.FullPath
	}
	return nil, ""
}

// getSelectedNodeInfo returns the node info for the selected tree item
func (p *projectViewPanelGroup) getSelectedNodeInfo() *TreeNodeInfo {
	itemHandle := p.projectTreeView.GetSelection()
	if itemHandle == 0 {
		return nil
	}
	return p.content.itemToNodeInfo[itemHandle]
}

// openSelectedRequest opens the selected request from project tree in a new tab
func (p *projectViewPanelGroup) openSelectedRequest(projectWindow TabController) {
	req, path := p.getSelectedRequest()
	if req == nil {
		return
	}

	// Open in new tab bound to the request
	projectWindow.createRequestTab(req, path)
}

// showContextMenu displays a context menu for the tree item
func (p *projectViewPanelGroup) showContextMenu(factory win32.ControlFactory, itemHandle uintptr) {
	nodeInfo := p.content.itemToNodeInfo[itemHandle]

	menu := factory.CreatePopupMenu()
	defer menu.Destroy()

	menu.AddItem(menuIDAddRequest, "Add Request")
	menu.AddItem(menuIDEdit, "Edit")
	menu.AddSeparator()
	menu.AddItem(menuIDDelete, "Delete")

	selectedID := menu.Show()
	switch selectedID {
	case menuIDAddRequest:
		p.addNode(nodeInfo)
	case menuIDDelete:
		p.deleteNode(itemHandle, nodeInfo)
	case menuIDEdit:
		if nodeInfo != nil && nodeInfo.Request != nil {
			p.openSelectedRequest(p.tabController)
		}
	}
}

// addNode adds a new path segment or request to a node
func (p *projectViewPanelGroup) addNode(nodeInfo *TreeNodeInfo) {
	if nodeInfo == nil {
		nodeInfo = &TreeNodeInfo{Type: NodeTypePath, FullPath: ""}
	}
	req := p.content.BoundProject.NewRequest()
	p.content.BoundProject.AddRequestToTree(nodeInfo.FullPath, req)
	p.SetState(p.content)
}

// deleteNode deletes a node from the tree
func (p *projectViewPanelGroup) deleteNode(_ uintptr, nodeInfo *TreeNodeInfo) {
	if nodeInfo == nil {
		return
	}

	p.controlFactory.MessageBox("Warning", "Not yet implemented: deleting individual requests")
	//TODO: implement deletion
	switch nodeInfo.Type {
	case NodeTypeRequest:
		if nodeInfo.Request != nil {
			//p.content.BoundProject.RemoveRequestFromTree(nodeInfo.Request)
		}
	case NodeTypePath:
		// Delete all requests under this path
		//segments, _ := ParseURLPath("/" + nodeInfo.FullPath)
		//node := p.content.BoundProject.Tree.FindNode(segments)
		//if node != nil {
		//	p.deleteNodeRecursive(node)
		//}
	}

	p.SetState(p.content)
}

const (
	environmentColumnDefault = iota
	environmentColumnName
	environmentColumnBaseURL
)

func createProjectViewPanel(factory win32.ControlFactory, tabController TabController, projectManager ProjectManager) *projectViewPanelGroup {
	group := &projectViewPanelGroup{
		envLabel:       factory.CreateLabel("Environments:"),
		projectInfo:    factory.CreateLabel("Double-click a request to open it in a new tab"),
		timeoutLabel:   factory.CreateLabel("Request Timeout (milliseconds):"),
		timeoutInput:   factory.CreateInput(),
		controlFactory: factory,
		tabController:  tabController,
	}

	// Create environment ListView
	group.envListView = factory.CreateListView()
	group.envListView.InsertColumn(environmentColumnDefault, "*", 20)
	group.envListView.InsertColumn(environmentColumnName, "Name", 150)
	group.envListView.InsertColumn(environmentColumnBaseURL, "Base URL", 200)

	// Set up edit end callback for in-place editing
	group.envListView.SetOnEditEnd(func(row, col int, newText string) {
		if group.content != nil && group.content.BoundProject != nil {
			if row >= 0 && row < len(group.content.BoundProject.Environments) {
				env := &group.content.BoundProject.Environments[row]
				switch col {
				case environmentColumnName:
					// Update name
					env.Name = newText
				case environmentColumnBaseURL:
					// Update BaseURL - validate it's not empty
					if newText != "" {
						env.BaseURL = newText
					} else {
						// Restore original value if empty
						group.envListView.SetItemText(row, col, env.BaseURL)
						factory.MessageBox("Invalid Input", "Base URL cannot be empty")
					}
				}
				// Refresh the list to update the display
				group.populateEnvironmentList()
			}
		}
	})

	// Environment management buttons
	group.addEnvBtn = factory.CreateButton("Add Env", func() {
		group.addEnvironment()
	})
	group.deleteEnvBtn = factory.CreateButton("Delete Env", func() {
		group.deleteEnvironment()
	})
	group.setDefaultBtn = factory.CreateButton("Set Default", func() {
		group.setDefaultEnvironment()
	})

	// Create project tree view
	group.projectTreeView = factory.CreateTreeView(func(tvc *win32.TreeViewControl) {
		group.openSelectedRequest(tabController)
	})

	// Set up right-click handler
	group.projectTreeView.SetOnRightClick(func(tvc *win32.TreeViewControl, itemHandle uintptr) {
		group.showContextMenu(factory, itemHandle)
	})

	group.openReqBtn = factory.CreateButton("Open in Tab", func() {
		group.openSelectedRequest(tabController)
	})
	group.deleteReqBtn = factory.CreateButton("Delete", func() {
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo != nil {
			group.deleteNode(group.projectTreeView.GetSelection(), nodeInfo)
		}
	})
	group.addPathBtn = factory.CreateButton("Add Path", func() { // TODO: Unifiy Add Path and Add Request
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo == nil {
			nodeInfo = &TreeNodeInfo{Type: NodeTypePath, FullPath: ""}
		}

		// Create a new pending request with a suggested path
		req := group.content.BoundProject.NewRequest()

		// Create a pending request tab with the suggested path
		tabController.createPendingRequestTab(req, nodeInfo.FullPath+"/newpath")
	})
	group.addRequestBtn = factory.CreateButton("Add Request", func() {
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo == nil {
			nodeInfo = &TreeNodeInfo{Type: NodeTypePath, FullPath: ""}
		}
		group.addNode(nodeInfo)
	})
	group.saveBtn = factory.CreateButton("Save Project", func() {
		group.SaveState() // Save timeout before saving to file
		projectManager.saveProject()
	})

	group.ControllerGroup = win32.NewControllerGroup(
		group.envLabel,
		group.envListView,
		group.addEnvBtn,
		group.deleteEnvBtn,
		group.setDefaultBtn,
		group.projectTreeView,
		group.openReqBtn,
		group.deleteReqBtn,
		group.addPathBtn,
		group.addRequestBtn,
		group.projectInfo,
		group.saveBtn,
		group.timeoutLabel,
		group.timeoutInput,
	)
	return group
}
