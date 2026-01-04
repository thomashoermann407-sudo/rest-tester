package win32

// Tab represents a single tab in the tab bar
type Tab[T any] struct {
	Title          string
	Data           T
	PanelGroupName PanelGroupName
}

// hitTestResult represents what was hit in the tab bar
type hitTestResult int

const (
	hitNone hitTestResult = iota
	hitTab
	hitCloseButton
	hitAddButton
	hitMenuButton
)

// TabManager manages a Chrome-style tab bar integrated with title bar
type TabManager[T any] struct {
	parentHwnd     hWnd
	tabs           []*Tab[T]
	panels         *Panels
	activeTabIndex int
	hoverTabIndex  int
	hoverAddBtn    bool
	hoverMenuBtn   bool

	// Dimensions
	titleBarHeight int32
	tabHeight      int32
	tabMinWidth    int32
	tabMaxWidth    int32
	tabPadding     int32
	tabGap         int32 // Gap between tabs
	closeSize      int32
	addBtnSize     int32
	cornerRadius   int32
	menuBtnSize    int32

	// Colors
	tabBgColor      colorRef
	tabActiveColor  colorRef
	tabHoverColor   colorRef
	textColor       colorRef
	textActiveColor colorRef
	closeBtnColor   colorRef
	closeBtnHover   colorRef
	closeBtnHoverBg colorRef

	// Fonts
	font hFont

	// Pens
	bgBrush      hBrush
	tabBorderPen hPen
	btnPen       hPen

	// Callbacks
	OnTabClosed func()
	OnNewTab    func()
	OnMenuClick func()
}

// NewTabManager creates a new tab manager
func NewTabManager[T any](window *Window) *TabManager[T] {
	titleBarHeight := int32(46)
	return &TabManager[T]{
		parentHwnd:     window.hwnd,
		tabs:           make([]*Tab[T], 0),
		panels:         NewPanels(titleBarHeight, window.width, window.height),
		activeTabIndex: -1,
		hoverTabIndex:  -1,
		hoverAddBtn:    false,
		hoverMenuBtn:   false,

		titleBarHeight: titleBarHeight,
		tabHeight:      34,
		tabMinWidth:    80,
		tabMaxWidth:    200,
		tabPadding:     12,
		tabGap:         2,
		closeSize:      16,
		addBtnSize:     28,
		cornerRadius:   8,
		menuBtnSize:    38,

		tabBgColor:      rgb(243, 243, 243),
		tabActiveColor:  rgb(255, 255, 255),
		tabHoverColor:   rgb(235, 235, 235),
		textColor:       rgb(96, 96, 96),
		textActiveColor: rgb(32, 32, 32),
		closeBtnColor:   rgb(128, 128, 128),
		closeBtnHover:   rgb(255, 255, 255),
		closeBtnHoverBg: rgb(196, 43, 28),

		bgBrush:      createSolidBrush(rgb(243, 243, 243)),
		tabBorderPen: createPen(PS_SOLID, 1, rgb(229, 229, 229)),
		btnPen:       createPen(PS_SOLID, 1, rgb(32, 32, 32)),

		font: createFont(-12, 0, 0, 0, FW_NORMAL, 0, 0, 0, DEFAULT_CHARSET, OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS, CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE, "Segoe UI"),
	}
}

func (tm *TabManager[T]) AddTab(title string, data T, panelGroupName PanelGroupName) {
	tab := &Tab[T]{
		Title:          title,
		Data:           data,
		PanelGroupName: panelGroupName,
	}
	tm.tabs = append(tm.tabs, tab)

	tm.Invalidate()
	tm.SetActiveTab(len(tm.tabs) - 1)
}

// RemoveTab removes a tab by index
func (tm *TabManager[T]) RemoveTab(tabIndex int) {
	tm.tabs = append(tm.tabs[:tabIndex], tm.tabs[tabIndex+1:]...)
	// If we removed the active tab, activate another
	if tm.activeTabIndex == tabIndex {
		if len(tm.tabs) > 0 {
			// Prefer the tab at the same position, or the last one
			newIndex := tabIndex
			if newIndex >= len(tm.tabs) {
				newIndex = len(tm.tabs) - 1
			}
			tm.activeTabIndex = newIndex
			tm.onTabChanged()
		} else {
			tm.activeTabIndex = -1
		}
	}

	if tm.OnTabClosed != nil {
		tm.OnTabClosed()
	}

	tm.Invalidate()

}

// SetActiveTab sets the active tab by index
func (tm *TabManager[T]) SetActiveTab(tabIndex int) *Tab[T] {
	if tabIndex >= 0 && tabIndex < len(tm.tabs) {
		if tm.activeTabIndex != tabIndex {
			// Call before change callback to allow saving state
			if activeTab := tm.getActiveTab(); activeTab != nil {
				tm.panels.get(activeTab.PanelGroupName).SaveState()
			}

			tm.activeTabIndex = tabIndex
			tm.onTabChanged()
			tm.Invalidate()
		}
	}
	return tm.getActiveTab()
}

func (tm *TabManager[T]) onTabChanged() {
	tab := tm.getActiveTab()
	tm.panels.Show(tab.PanelGroupName)
	tm.panels.get(tab.PanelGroupName).SetState(tab.Data)
}

// Todo: remove
func (tm *TabManager[T]) GetPanels() *Panels {
	return tm.panels
}

// getActiveTab returns the currently active tab
func (tm *TabManager[T]) getActiveTab() *Tab[T] {
	if tm.activeTabIndex < 0 || tm.activeTabIndex >= len(tm.tabs) {
		return nil
	}
	return tm.tabs[tm.activeTabIndex]
}

// GetTabCount returns the number of tabs
func (tm *TabManager[T]) GetTabCount() int {
	return len(tm.tabs)
}

// FindTabByPanelGroup finds the index of the first tab with the given panel group
// Returns -1 if no tab is found
func (tm *TabManager[T]) FindTabByPanelGroup(panelGroupName PanelGroupName) (int, bool) {
	for i, tab := range tm.tabs {
		if tab.PanelGroupName == panelGroupName {
			return i, true
		}
	}
	return -1, false
}

// GetHeight returns the total title bar height
func (tm *TabManager[T]) GetHeight() int32 {
	return tm.titleBarHeight
}

// Invalidate triggers a repaint of the tab bar area
func (tm *TabManager[T]) Invalidate() {
	if tm.parentHwnd != 0 {
		// Get the actual client rect width for proper invalidation
		var clientRect rect
		getClientRect(tm.parentHwnd, &clientRect)

		rect := &rect{
			Left:   0,
			Top:    0,
			Right:  clientRect.Right, // Use actual window width
			Bottom: tm.titleBarHeight,
		}
		invalidateRect(tm.parentHwnd, rect, true)
		// Force immediate repaint to ensure tabs are redrawn when added/removed
		updateWindow(tm.parentHwnd)
	}
}

// getTabRect calculates the rectangle for a tab at the given index
func (tm *TabManager[T]) getTabRect(index int, totalWidth int32) *rect {
	numTabs := int32(len(tm.tabs))
	if numTabs == 0 {
		return &rect{}
	}

	// Reserve space for: menu button on the right side
	rightReserved := tm.menuBtnSize + tm.tabPadding*2

	// Available width for tabs and add button
	availableWidth := totalWidth - rightReserved - tm.addBtnSize - tm.tabPadding*2

	// Calculate tab width
	tabWidth := max(min((availableWidth-(numTabs-1)*tm.tabGap)/numTabs, tm.tabMaxWidth), tm.tabMinWidth)

	// Tab Y position - center vertically in title bar
	topMargin := (tm.titleBarHeight-tm.tabHeight)/2 + 2

	left := tm.tabPadding + int32(index)*(tabWidth+tm.tabGap)
	return &rect{
		Left:   left,
		Top:    topMargin,
		Right:  left + tabWidth,
		Bottom: topMargin + tm.tabHeight,
	}
}

// getCloseRect calculates the close button rectangle for a tab
func (tm *TabManager[T]) getCloseRect(tabRect *rect) *rect {
	padding := int32(8)
	centerY := (tabRect.Top + tabRect.Bottom) / 2
	return &rect{
		Left:   tabRect.Right - tm.closeSize - padding,
		Top:    centerY - tm.closeSize/2,
		Right:  tabRect.Right - padding,
		Bottom: centerY + tm.closeSize/2,
	}
}

// getAddButtonRect returns the rectangle for the add button
func (tm *TabManager[T]) getAddButtonRect(totalWidth int32) *rect {
	numTabs := int32(len(tm.tabs))
	rightReserved := tm.menuBtnSize + tm.tabPadding*2
	availableWidth := totalWidth - rightReserved - tm.addBtnSize - tm.tabPadding*2

	tabWidth := max(min((availableWidth-(numTabs-1)*tm.tabGap)/max(numTabs, 1), tm.tabMaxWidth), tm.tabMinWidth)

	topMargin := (tm.titleBarHeight-tm.tabHeight)/2 + 2
	centerY := topMargin + tm.tabHeight/2

	left := tm.tabPadding + numTabs*(tabWidth+tm.tabGap) + 4
	return &rect{
		Left:   left,
		Top:    centerY - tm.addBtnSize/2,
		Right:  left + tm.addBtnSize,
		Bottom: centerY + tm.addBtnSize/2,
	}
}

// getMenuButtonRect returns the rectangle for the menu button (right side)
func (tm *TabManager[T]) getMenuButtonRect(totalWidth int32) *rect {
	centerY := tm.titleBarHeight / 2
	return &rect{
		Left:   totalWidth - tm.menuBtnSize - tm.tabPadding,
		Top:    centerY - tm.menuBtnSize/2 + 2,
		Right:  totalWidth - tm.tabPadding,
		Bottom: centerY + tm.menuBtnSize/2 + 2,
	}
}

// hitTest determines what was clicked/hovered
func (tm *TabManager[T]) hitTest(x, y int32, totalWidth int32) (result hitTestResult, tabIndex int) {
	tabIndex = -1

	// Check menu button
	menuRect := tm.getMenuButtonRect(totalWidth)
	if menuRect.inside(x, y) {
		return hitMenuButton, -1
	}

	// Check add button
	addRect := tm.getAddButtonRect(totalWidth)
	if addRect.inside(x, y) {
		return hitAddButton, -1
	}

	// Check each tab
	for i := range tm.tabs {
		tabRect := tm.getTabRect(i, totalWidth)
		if tabRect.inside(x, y) {
			tabIndex = i

			// Check close button within tab
			closeRect := tm.getCloseRect(tabRect)
			if closeRect.inside(x, y) {
				return hitCloseButton, tabIndex
			}
			return hitTab, tabIndex
		}
	}

	return hitNone, -1
}

// HandleMouseMove handles WM_MOUSEMOVE
func (tm *TabManager[T]) HandleMouseMove(x, y int32, totalWidth int32) {
	if y > tm.titleBarHeight {
		if tm.hoverTabIndex != -1 || tm.hoverAddBtn || tm.hoverMenuBtn {
			tm.hoverTabIndex = -1
			tm.hoverAddBtn = false
			tm.hoverMenuBtn = false
			tm.Invalidate()
		}
		return
	}

	oldHoverIndex := tm.hoverTabIndex
	oldHoverAdd := tm.hoverAddBtn
	oldHoverMenu := tm.hoverMenuBtn

	result, tabIndex := tm.hitTest(x, y, totalWidth)

	tm.hoverTabIndex = -1
	tm.hoverAddBtn = false
	tm.hoverMenuBtn = false

	switch result {
	case hitTab:
		tm.hoverTabIndex = tabIndex
	case hitCloseButton:
		tm.hoverTabIndex = tabIndex
	case hitAddButton:
		tm.hoverAddBtn = true
	case hitMenuButton:
		tm.hoverMenuBtn = true
	}

	if oldHoverIndex != tm.hoverTabIndex ||
		oldHoverAdd != tm.hoverAddBtn || oldHoverMenu != tm.hoverMenuBtn {
		tm.Invalidate()
	}
}

// HandleClick handles mouse click
func (tm *TabManager[T]) HandleClick(x, y int32, totalWidth int32) {
	result, tabIndex := tm.hitTest(x, y, totalWidth)

	switch result {
	case hitAddButton:
		if tm.OnNewTab != nil {
			tm.OnNewTab()
		}
	case hitMenuButton:
		if tm.OnMenuClick != nil {
			tm.OnMenuClick()
		}
	case hitCloseButton:
		tm.RemoveTab(tabIndex)
	case hitTab:
		tm.SetActiveTab(tabIndex)
	}
}

// Paint draws the entire title bar with tabs
func (tm *TabManager[T]) Paint(hdc hDc, width int32) {
	// Draw background
	bgBrush := tm.bgBrush
	bgRect := rect{Left: 0, Top: 0, Right: width, Bottom: tm.titleBarHeight}
	fillRect(hdc, &bgRect, bgBrush)

	// Set up drawing
	setBkMode(hdc, TRANSPARENT)
	oldFont := selectObject(hdc, handle(tm.font))

	// Draw each tab
	for i, tab := range tm.tabs {
		tm.drawTab(hdc, i, tab, width)
	}

	// Draw add button
	tm.drawAddButton(hdc, width)

	// Draw menu button
	tm.drawMenuButton(hdc, width)

	// Draw subtle separator line at bottom
	tm.drawBottomLine(hdc, width)

	selectObject(hdc, oldFont)
}

// drawTab draws a single tab with modern styling
func (tm *TabManager[T]) drawTab(hdc hDc, index int, tab *Tab[T], totalWidth int32) {
	tabRect := tm.getTabRect(index, totalWidth)
	isActive := index == tm.activeTabIndex
	isHover := index == tm.hoverTabIndex

	// Determine colors
	var bgColor colorRef
	var textColor colorRef

	if isActive {
		bgColor = tm.tabActiveColor
		textColor = tm.textActiveColor
	} else if isHover {
		bgColor = tm.tabHoverColor
		textColor = tm.textActiveColor
	} else {
		bgColor = tm.tabBgColor
		textColor = tm.textColor
	}

	// Draw tab background for active or hover tabs
	if isActive || isHover {
		tm.drawRoundedTabBackground(hdc, tabRect, bgColor)
	}

	// Draw tab text
	setTextColor(hdc, textColor)
	textRect := &rect{
		Left:   tabRect.Left + tm.tabPadding,
		Top:    tabRect.Top,
		Right:  tabRect.Right - tm.tabPadding,
		Bottom: tabRect.Bottom,
	}

	// Leave room for close button
	if isActive || isHover {
		textRect.Right -= tm.closeSize + 8
	}

	drawText(hdc, tab.Title, textRect, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)

	// Draw close button if applicable
	if isActive || isHover {
		tm.drawCloseButton(hdc, tabRect, isHover)
	}
}

// drawRoundedTabBackground draws a rounded rectangle background for a tab
func (tm *TabManager[T]) drawRoundedTabBackground(hdc hDc, tabRect *rect, color colorRef) {
	brush := createSolidBrush(color)
	pen := createPen(PS_SOLID, 1, color)
	oldBrush := selectObject(hdc, handle(brush))
	oldPen := selectObject(hdc, handle(pen))

	// Draw rounded rectangle for the tab, but don't extend beyond the separator line
	maxBottom := tm.titleBarHeight - 1

	roundRect(hdc, tabRect, tm.cornerRadius*2, tm.cornerRadius*2)

	// Fill bottom part to make only top corners rounded
	bottomRect := &rect{
		Left:   tabRect.Left,
		Top:    tabRect.Bottom - tm.cornerRadius,
		Right:  tabRect.Right,
		Bottom: maxBottom,
	}
	fillRect(hdc, bottomRect, brush)

	selectObject(hdc, oldPen)
	selectObject(hdc, oldBrush)
	deleteObject(handle(pen))
	deleteObject(handle(brush))
}

// drawCloseButton draws the X button for closing a tab
func (tm *TabManager[T]) drawCloseButton(hdc hDc, tabRect *rect, isHover bool) {
	closeRect := tm.getCloseRect(tabRect)

	// Draw hover background (rounded)
	if isHover {
		tm.drawRoundedRect(hdc, closeRect, 8, tm.closeBtnHoverBg)
	}

	// Draw X
	penColor := tm.closeBtnColor
	if isHover {
		penColor = tm.closeBtnHover
	}

	tm.drawX(hdc, closeRect, penColor, 4)
}

// drawAddButton draws the + button for adding tabs
func (tm *TabManager[T]) drawAddButton(hdc hDc, totalWidth int32) {
	rect := tm.getAddButtonRect(totalWidth)

	// Draw hover background
	if tm.hoverAddBtn {
		tm.drawRoundedRect(hdc, rect, 6, tm.tabHoverColor)
	}

	tm.drawPlus(hdc, rect, 5)
}

// drawMenuButton draws the hamburger menu button
func (tm *TabManager[T]) drawMenuButton(hdc hDc, totalWidth int32) {
	rect := tm.getMenuButtonRect(totalWidth)

	// Draw hover background
	if tm.hoverMenuBtn {
		tm.drawRoundedRect(hdc, rect, 6, tm.tabHoverColor)
	}

	tm.drawHamburger(hdc, rect, 7, 4)
}

// drawRoundedRect draws a filled rounded rectangle (helper method)
func (tm *TabManager[T]) drawRoundedRect(hdc hDc, rect *rect, cornerRadius int32, color colorRef) {
	brush := createSolidBrush(color)
	pen := createPen(PS_SOLID, 1, color)
	oldBrush := selectObject(hdc, handle(brush))
	oldPen := selectObject(hdc, handle(pen))

	roundRect(hdc, rect, cornerRadius, cornerRadius)

	selectObject(hdc, oldPen)
	selectObject(hdc, oldBrush)
	deleteObject(handle(pen))
	deleteObject(handle(brush))
}

// drawX draws an X icon (helper method)
func (tm *TabManager[T]) drawX(hdc hDc, rect *rect, color colorRef, padding int32) {
	pen := createPen(PS_SOLID, 1, color)
	oldPen := selectObject(hdc, handle(pen))

	// Draw X lines
	moveToEx(hdc, rect.Left+padding, rect.Top+padding, nil)
	lineTo(hdc, rect.Right-padding+1, rect.Bottom-padding+1)
	moveToEx(hdc, rect.Right-padding, rect.Top+padding, nil)
	lineTo(hdc, rect.Left+padding-1, rect.Bottom-padding+1)

	selectObject(hdc, oldPen)
	deleteObject(handle(pen))
}

// drawPlus draws a + icon (helper method)
func (tm *TabManager[T]) drawPlus(hdc hDc, rect *rect, size int32) {
	oldPen := selectObject(hdc, handle(tm.btnPen))

	centerX := (rect.Left + rect.Right) / 2
	centerY := (rect.Top + rect.Bottom) / 2

	// Horizontal line
	moveToEx(hdc, centerX-size, centerY, nil)
	lineTo(hdc, centerX+size+1, centerY)

	// Vertical line
	moveToEx(hdc, centerX, centerY-size, nil)
	lineTo(hdc, centerX, centerY+size+1)

	selectObject(hdc, oldPen)
}

// drawHamburger draws a hamburger menu icon (helper method)
func (tm *TabManager[T]) drawHamburger(hdc hDc, rect *rect, width, spacing int32) {
	oldPen := selectObject(hdc, handle(tm.btnPen))

	centerX := (rect.Left + rect.Right) / 2
	centerY := (rect.Top + rect.Bottom) / 2

	// Three horizontal lines
	for i := int32(-1); i <= 1; i++ {
		y := centerY + i*spacing
		moveToEx(hdc, centerX-width, y, nil)
		lineTo(hdc, centerX+width+1, y)
	}

	selectObject(hdc, oldPen)
}

// drawBottomLine draws a subtle separator line at the bottom of the tab bar
func (tm *TabManager[T]) drawBottomLine(hdc hDc, totalWidth int32) {
	oldPen := selectObject(hdc, handle(tm.tabBorderPen))
	y := tm.titleBarHeight - 1
	moveToEx(hdc, 0, y, nil)
	lineTo(hdc, totalWidth, y)
	selectObject(hdc, oldPen)
}

// Destroy cleans up all OS resources allocated by the TabManager
// This should be called before the TabManager is discarded to prevent resource leaks
func (tm *TabManager[T]) Destroy() {
	deleteObject(handle(tm.font))
	deleteObject(handle(tm.bgBrush))
	deleteObject(handle(tm.tabBorderPen))
	deleteObject(handle(tm.btnPen))
}
