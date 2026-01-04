package main

import (
	"fmt"
	"strings"

	"hoermi.com/rest-test/win32"
)

type projectViewPanelGroup struct {
	*win32.ControllerGroup
	// Project View Panel controls
	projectTreeView *win32.TreeViewControl
	openReqBtn      *win32.ButtonControl
	deleteReqBtn    *win32.ButtonControl
	addHostBtn      *win32.ButtonControl
	addPathBtn      *win32.ButtonControl
	addMethodBtn    *win32.ButtonControl
	projectInfo     *win32.Control
	saveBtn         *win32.ButtonControl
	timeoutLabel    *win32.Control
	timeoutInput    *win32.Control

	content    *ProjectViewTabContent
	tabManager TabManager
}

// Menu IDs for context menu
const (
	menuIDAddHost = iota + 1000
	menuIDAddPath
	menuIDAddMethod
	menuIDDelete
	menuIDEdit
)

func (p *projectViewPanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight + layoutPadding
	dy := layoutLabelHeight + layoutPadding
	btnX := layoutPadding + layoutColumnWidth + layoutPadding
	p.projectInfo.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	p.projectTreeView.MoveWindow(layoutPadding, y+dy, layoutColumnWidth, layoutListHeight)

	// Right side buttons
	btnY := y + dy
	p.openReqBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.deleteReqBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.addHostBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.addPathBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
	btnY += dy
	p.addMethodBtn.MoveWindow(btnX, btnY, layoutButtonWidth, layoutInputHeight)
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
}

// populateTreeNode recursively populates the tree view from a request node
func (p *projectViewPanelGroup) populateTreeNode(parentHandle uintptr, node *RequestNode, pathPrefix string, isRoot bool) {
	if node == nil {
		return
	}

	// Build the current path
	currentPath := pathPrefix
	if node.Segment != "" {
		if currentPath != "" {
			currentPath += "/"
		}
		currentPath += node.Segment

		// Determine node type
		nodeType := NodeTypePath
		if isRoot {
			nodeType = NodeTypeHost
		}

		// Insert a node for this path segment
		segmentHandle := p.projectTreeView.InsertItem(parentHandle, win32.TVI_LAST, node.Segment, 0)

		// Store node info
		p.content.itemToNodeInfo[segmentHandle] = &TreeNodeInfo{
			Type:     nodeType,
			Segment:  node.Segment,
			FullPath: currentPath,
		}

		// Add request items for each HTTP method at this node
		for method, req := range node.Requests {
			displayText := fmt.Sprintf("[%s]", method)
			itemHandle := p.projectTreeView.InsertItem(segmentHandle, win32.TVI_LAST, displayText, 0)
			// Map the tree item to the request with node info
			p.content.itemToNodeInfo[itemHandle] = &TreeNodeInfo{
				Type:     NodeTypeMethod,
				Method:   method,
				Request:  req,
				FullPath: currentPath,
			}
		}

		// Recursively add children
		for _, child := range node.Children {
			p.populateTreeNode(segmentHandle, child, currentPath, false)
		}
	} else {
		// Root node - just process children (these are hosts)
		for _, child := range node.Children {
			p.populateTreeNode(parentHandle, child, currentPath, true)
		}
	}
}

func (p *projectViewPanelGroup) SetState(data any) {
	if content, ok := data.(*ProjectViewTabContent); ok {
		p.content = content
		if p.projectTreeView == nil || content.BoundProject == nil {
			return
		}

		// Clear the tree view
		p.projectTreeView.DeleteAllItems()
		p.content.itemToNodeInfo = make(map[uintptr]*TreeNodeInfo)

		// Populate the tree from the root
		p.populateTreeNode(win32.TVI_ROOT, content.BoundProject.Tree.Root, "", false)

		// Set timeout value
		timeout := content.BoundProject.Settings.TimeoutInMs
		if timeout == 0 {
			timeout = 30000 // Default 30 seconds
		}
		p.timeoutInput.SetText(fmt.Sprintf("%d", timeout))
	}
}

// saveProject saves the current project to file
func (p *projectViewPanelGroup) saveProject(factory win32.ControlFactory) {

	defaultName := p.content.BoundProject.Name + ".rtp"
	filePath, ok := factory.SaveFileDialog(
		"Save Project",
		"REST Project Files (*.rtp)|*.rtp|All Files (*.*)|*.*|",
		"rtp",
		defaultName,
	)
	if !ok {
		return
	}

	if err := p.content.BoundProject.Save(filePath); err != nil {
		factory.MessageBox(fmt.Sprintf("Error saving project: %v", err), "Error")
		return
	}

	// Update project name from filename
	name := filePath
	if idx := strings.LastIndex(name, "\\"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".rtp")
	p.content.BoundProject.Name = name
}

// getSelectedRequest returns the request associated with the selected tree item
func (p *projectViewPanelGroup) getSelectedRequest() *Request {
	itemHandle := p.projectTreeView.GetSelection()
	if itemHandle == 0 {
		return nil
	}
	nodeInfo := p.content.itemToNodeInfo[itemHandle]
	if nodeInfo != nil && nodeInfo.Type == NodeTypeMethod {
		return nodeInfo.Request
	}
	return nil
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
func (p *projectViewPanelGroup) openSelectedRequest(projectWindow TabManager) {
	req := p.getSelectedRequest()
	if req == nil {
		return
	}

	// Open in new tab bound to the request
	projectWindow.createRequestTab(req)
}

// showContextMenu displays a context menu for the tree item
func (p *projectViewPanelGroup) showContextMenu(factory win32.ControlFactory, itemHandle uintptr) {
	nodeInfo := p.content.itemToNodeInfo[itemHandle]

	menu := factory.CreatePopupMenu()
	defer menu.Destroy()

	if nodeInfo == nil {
		// Root level - can add hosts
		menu.AddItem(menuIDAddHost, "Add Host")
	} else {
		switch nodeInfo.Type {
		case NodeTypeHost:
			menu.AddItem(menuIDAddPath, "Add Path")
			menu.AddItem(menuIDAddMethod, "Add Method")
			menu.AddSeparator()
			menu.AddItem(menuIDDelete, "Delete Host")
		case NodeTypePath:
			menu.AddItem(menuIDAddPath, "Add Sub-Path")
			menu.AddItem(menuIDAddMethod, "Add Method")
			menu.AddSeparator()
			menu.AddItem(menuIDDelete, "Delete Path")
		case NodeTypeMethod:
			menu.AddItem(menuIDEdit, "Edit Request")
			menu.AddSeparator()
			menu.AddItem(menuIDDelete, "Delete Method")
		}
	}

	selectedID := menu.Show()
	p.handleMenuAction(selectedID, itemHandle, nodeInfo)
}

// handleMenuAction handles the selected context menu action
func (p *projectViewPanelGroup) handleMenuAction(menuID int, itemHandle uintptr, nodeInfo *TreeNodeInfo) {
	switch menuID {
	case menuIDAddHost:
		p.addHost()
	case menuIDAddPath:
		p.addPath(nodeInfo)
	case menuIDAddMethod:
		p.addMethod(nodeInfo)
	case menuIDDelete:
		p.deleteNode(itemHandle, nodeInfo)
	case menuIDEdit:
		if nodeInfo != nil && nodeInfo.Request != nil {
			p.openSelectedRequest(p.tabManager)
		}
	}
}

// addHost adds a new host to the tree
func (p *projectViewPanelGroup) addHost() {
	// Create a dummy request to populate the tree
	req := NewRequest("New Request")
	req.Host = "TODO: get host from user input"
	req.Method = "GET"

	p.content.BoundProject.AddRequestToTree(req)
	p.SetState(p.content)
}

// addPath adds a new path segment to a node
func (p *projectViewPanelGroup) addPath(nodeInfo *TreeNodeInfo) {
	if nodeInfo == nil {
		return
	}

	pathSegment := "TODO: get path segment from user input"
	// Build the full path
	fullPath := nodeInfo.FullPath + "/" + pathSegment

	// Create a dummy request to populate the tree
	req := NewRequest("New Request")
	req.Path = "/" + fullPath
	req.Method = "GET"

	p.content.BoundProject.AddRequestToTree(req)
	p.SetState(p.content)
}

// addMethod adds a new HTTP method to a node
func (p *projectViewPanelGroup) addMethod(nodeInfo *TreeNodeInfo) {
	if nodeInfo == nil {
		return
	}

	method := "TODO: get HTTP method from user input"
	method = strings.ToUpper(method)

	// Create a new request with path (no host)
	req := NewRequest(method + " " + nodeInfo.FullPath)
	req.Path = "/" + nodeInfo.FullPath
	req.Method = method
	req.Headers["Content-Type"] = "application/json"
	req.Headers["Accept"] = "application/json"

	p.content.BoundProject.AddRequestToTree(req)
	p.SetState(p.content)
}

// deleteNode deletes a node from the tree
func (p *projectViewPanelGroup) deleteNode(_ uintptr, nodeInfo *TreeNodeInfo) {
	if nodeInfo == nil {
		return
	}

	switch nodeInfo.Type {
	case NodeTypeMethod:
		if nodeInfo.Request != nil {
			p.content.BoundProject.RemoveRequestFromTree(nodeInfo.Request)
		}
	case NodeTypeHost, NodeTypePath:
		// Delete all requests under this path
		segments, _ := ParseURLPath("/" + nodeInfo.FullPath)
		node := p.content.BoundProject.Tree.Root.FindNode(segments)
		if node != nil {
			p.deleteNodeRecursive(node)
		}
	}

	p.SetState(p.content)
}

// deleteNodeRecursive deletes all requests in a node and its children
func (p *projectViewPanelGroup) deleteNodeRecursive(node *RequestNode) {
	// Delete all requests in this node
	for _, req := range node.Requests {
		p.content.BoundProject.RemoveRequestFromTree(req)
	}

	// Recursively delete children
	for _, child := range node.Children {
		p.deleteNodeRecursive(child)
	}
}

func createProjectViewPanel(factory win32.ControlFactory, tabManager TabManager) *projectViewPanelGroup {
	group := &projectViewPanelGroup{
		projectInfo:  factory.CreateLabel("Double-click a request to open it in a new tab"),
		timeoutLabel: factory.CreateLabel("Request Timeout (milliseconds):"),
		timeoutInput: factory.CreateInput(),
	}

	group.projectTreeView = factory.CreateTreeView(func(tvc *win32.TreeViewControl) {
		group.openSelectedRequest(tabManager)
	})

	// Set up right-click handler
	group.projectTreeView.SetOnRightClick(func(tvc *win32.TreeViewControl, itemHandle uintptr) {
		group.showContextMenu(factory, itemHandle)
	})

	group.openReqBtn = factory.CreateButton("Open in Tab", func() {
		group.openSelectedRequest(tabManager)
	})
	group.deleteReqBtn = factory.CreateButton("Delete", func() {
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo != nil {
			group.deleteNode(group.projectTreeView.GetSelection(), nodeInfo)
		}
	})
	group.addHostBtn = factory.CreateButton("Add Host", func() {
		group.addHost()
	})
	group.addPathBtn = factory.CreateButton("Add Path", func() {
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo != nil {
			group.addPath(nodeInfo)
		}
	})
	group.addMethodBtn = factory.CreateButton("Add Method", func() {
		nodeInfo := group.getSelectedNodeInfo()
		if nodeInfo != nil {
			group.addMethod(nodeInfo)
		}
	})
	group.saveBtn = factory.CreateButton("Save Project", func() {
		group.SaveState() // Save timeout before saving to file
		group.saveProject(factory)
	})

	group.ControllerGroup = win32.NewControllerGroup(
		group.projectTreeView, group.openReqBtn, group.deleteReqBtn,
		group.addHostBtn, group.addPathBtn, group.addMethodBtn,
		group.projectInfo, group.saveBtn, group.timeoutLabel, group.timeoutInput,
	)
	return group
}
