package win32

// Tab represents a single tab in the tab bar
type Tab[T any] struct {
	Title string
	Data  T
}

// HitTestResult represents what was hit in the tab bar
type HitTestResult int

const (
	HitNone HitTestResult = iota
	HitTab
	HitCloseButton
	HitAddButton
	HitMenuButton
)

// TabManager manages a Chrome-style tab bar integrated with title bar
type TabManager[T any] struct {
	parentHwnd     HWND
	tabs           []*Tab[T]
	activeTabIndex int
	hoverTabIndex  int
	hoverCloseBtn  bool
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

	// Colors (Windows 11 style - Mica-like)
	bgColor         COLORREF
	tabBgColor      COLORREF
	tabActiveColor  COLORREF
	tabHoverColor   COLORREF
	textColor       COLORREF
	textActiveColor COLORREF
	closeBtnColor   COLORREF
	closeBtnHover   COLORREF
	closeBtnHoverBg COLORREF
	addBtnColor     COLORREF
	addBtnHover     COLORREF
	borderColor     COLORREF

	// Fonts
	font HFONT

	// Callbacks
	OnTabChanged      func()
	OnTabClosed       func()
	OnBeforeTabChange func()
	OnMenuClick       func()
}

// NewTabManager creates a new tab manager
func NewTabManager[T any](parent HWND) *TabManager[T] {
	tm := &TabManager[T]{
		parentHwnd:     parent,
		tabs:           make([]*Tab[T], 0),
		activeTabIndex: -1,
		hoverTabIndex:  -1,
		hoverCloseBtn:  false,
		hoverAddBtn:    false,
		hoverMenuBtn:   false,

		titleBarHeight: 46,
		tabHeight:      34,
		tabMinWidth:    80,
		tabMaxWidth:    200,
		tabPadding:     12,
		tabGap:         2,
		closeSize:      16,
		addBtnSize:     28,
		cornerRadius:   8,
		menuBtnSize:    38,

		// Windows 11 Mica-inspired colors (light theme)
		bgColor:         rgb(243, 243, 243),
		tabBgColor:      rgb(243, 243, 243),
		tabActiveColor:  rgb(255, 255, 255),
		tabHoverColor:   rgb(235, 235, 235),
		textColor:       rgb(96, 96, 96),
		textActiveColor: rgb(32, 32, 32),
		closeBtnColor:   rgb(128, 128, 128),
		closeBtnHover:   rgb(255, 255, 255),
		closeBtnHoverBg: rgb(196, 43, 28),
		addBtnColor:     rgb(96, 96, 96),
		addBtnHover:     rgb(32, 32, 32),
		borderColor:     rgb(229, 229, 229),
	}

	// Create fonts
	tm.font = createFont(
		-12, 0, 0, 0, FW_NORMAL,
		0, 0, 0, DEFAULT_CHARSET,
		OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS,
		CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE,
		"Segoe UI",
	)

	return tm
}

func (tm *TabManager[T]) AddTab(title string, data T) {
	tab := &Tab[T]{
		Title: title,
		Data:  data,
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
			// Trigger tab changed callback for new active tab
			if tm.OnTabChanged != nil {
				tm.OnTabChanged()
			}
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
func (tm *TabManager[T]) SetActiveTab(tabIndex int) {
	if tabIndex >= 0 && tabIndex < len(tm.tabs) {
		if tm.activeTabIndex != tabIndex {
			// Call before change callback to allow saving state
			if tm.OnBeforeTabChange != nil && tm.activeTabIndex != -1 {
				tm.OnBeforeTabChange()
			}

			tm.activeTabIndex = tabIndex
			if tm.OnTabChanged != nil {
				tm.OnTabChanged()
			}
			tm.Invalidate()
		}
	}
}

// GetActiveTab returns the currently active tab
func (tm *TabManager[T]) GetActiveTab() *Tab[T] {
	if tm.activeTabIndex < 0 || tm.activeTabIndex >= len(tm.tabs) {
		return nil
	}
	return tm.tabs[tm.activeTabIndex]
}

// GetTab returns a tab by Index
func (tm *TabManager[T]) GetTab(tabIndex int) *Tab[T] {
	if tabIndex >= 0 && tabIndex < len(tm.tabs) {
		return tm.tabs[tabIndex]
	}
	return nil
}

// GetTabCount returns the number of tabs
func (tm *TabManager[T]) GetTabCount() int {
	return len(tm.tabs)
}

// GetHeight returns the total title bar height
func (tm *TabManager[T]) GetHeight() int32 {
	return tm.titleBarHeight
}

// GetContentOffset returns the Y offset where content should start
func (tm *TabManager[T]) GetContentOffset() int32 {
	return tm.titleBarHeight
}

// Invalidate triggers a repaint of the tab bar area
func (tm *TabManager[T]) Invalidate() {
	if tm.parentHwnd != 0 {
		// Get the actual client rect width for proper invalidation
		var clientRect RECT
		getClientRect(tm.parentHwnd, &clientRect)

		rect := RECT{
			Left:   0,
			Top:    0,
			Right:  clientRect.Right, // Use actual window width
			Bottom: tm.titleBarHeight,
		}
		invalidateRect(tm.parentHwnd, &rect, true)
		// Force immediate repaint to ensure tabs are redrawn when added/removed
		updateWindow(tm.parentHwnd)
	}
}

// getTabRect calculates the rectangle for a tab at the given index
func (tm *TabManager[T]) getTabRect(index int, totalWidth int32) RECT {
	numTabs := int32(len(tm.tabs))
	if numTabs == 0 {
		return RECT{}
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
	return RECT{
		Left:   left,
		Top:    topMargin,
		Right:  left + tabWidth,
		Bottom: topMargin + tm.tabHeight,
	}
}

// getCloseRect calculates the close button rectangle for a tab
func (tm *TabManager[T]) getCloseRect(tabRect RECT) RECT {
	padding := int32(8)
	centerY := (tabRect.Top + tabRect.Bottom) / 2
	return RECT{
		Left:   tabRect.Right - tm.closeSize - padding,
		Top:    centerY - tm.closeSize/2,
		Right:  tabRect.Right - padding,
		Bottom: centerY + tm.closeSize/2,
	}
}

// getAddButtonRect returns the rectangle for the add button
func (tm *TabManager[T]) getAddButtonRect(totalWidth int32) RECT {
	numTabs := int32(len(tm.tabs))
	rightReserved := tm.menuBtnSize + tm.tabPadding*2
	availableWidth := totalWidth - rightReserved - tm.addBtnSize - tm.tabPadding*2

	tabWidth := max(min((availableWidth-(numTabs-1)*tm.tabGap)/max(numTabs, 1), tm.tabMaxWidth), tm.tabMinWidth)

	topMargin := (tm.titleBarHeight-tm.tabHeight)/2 + 2
	centerY := topMargin + tm.tabHeight/2

	left := tm.tabPadding + numTabs*(tabWidth+tm.tabGap) + 4
	return RECT{
		Left:   left,
		Top:    centerY - tm.addBtnSize/2,
		Right:  left + tm.addBtnSize,
		Bottom: centerY + tm.addBtnSize/2,
	}
}

// getMenuButtonRect returns the rectangle for the menu button (right side)
func (tm *TabManager[T]) getMenuButtonRect(totalWidth int32) RECT {
	centerY := tm.titleBarHeight / 2
	return RECT{
		Left:   totalWidth - tm.menuBtnSize - tm.tabPadding,
		Top:    centerY - tm.menuBtnSize/2 + 2,
		Right:  totalWidth - tm.tabPadding,
		Bottom: centerY + tm.menuBtnSize/2 + 2,
	}
}

// HitTest determines what was clicked/hovered
func (tm *TabManager[T]) HitTest(x, y int32, totalWidth int32) (result HitTestResult, tabIndex int) {
	tabIndex = -1

	// Check menu button
	menuRect := tm.getMenuButtonRect(totalWidth)
	if x >= menuRect.Left && x <= menuRect.Right && y >= menuRect.Top && y <= menuRect.Bottom {
		return HitMenuButton, -1
	}

	// Check add button
	addRect := tm.getAddButtonRect(totalWidth)
	if x >= addRect.Left && x <= addRect.Right && y >= addRect.Top && y <= addRect.Bottom {
		return HitAddButton, -1
	}

	// Check each tab
	for i := range tm.tabs {
		tabRect := tm.getTabRect(i, totalWidth)
		if x >= tabRect.Left && x <= tabRect.Right && y >= tabRect.Top && y <= tabRect.Bottom {
			tabIndex = i

			// Check close button within tab
			closeRect := tm.getCloseRect(tabRect)
			if x >= closeRect.Left && x <= closeRect.Right && y >= closeRect.Top && y <= closeRect.Bottom {
				return HitCloseButton, tabIndex
			}
			return HitTab, tabIndex
		}
	}

	return HitNone, -1
}

// HandleMouseMove handles WM_MOUSEMOVE
func (tm *TabManager[T]) HandleMouseMove(x, y int32, totalWidth int32) {
	if y > tm.titleBarHeight {
		if tm.hoverTabIndex != -1 || tm.hoverAddBtn || tm.hoverMenuBtn {
			tm.hoverTabIndex = -1
			tm.hoverCloseBtn = false
			tm.hoverAddBtn = false
			tm.hoverMenuBtn = false
			tm.Invalidate()
		}
		return
	}

	oldHoverIndex := tm.hoverTabIndex
	oldHoverClose := tm.hoverCloseBtn
	oldHoverAdd := tm.hoverAddBtn
	oldHoverMenu := tm.hoverMenuBtn

	result, tabIndex := tm.HitTest(x, y, totalWidth)

	tm.hoverTabIndex = -1
	tm.hoverCloseBtn = false
	tm.hoverAddBtn = false
	tm.hoverMenuBtn = false

	switch result {
	case HitTab:
		tm.hoverTabIndex = tabIndex
	case HitCloseButton:
		tm.hoverTabIndex = tabIndex
		tm.hoverCloseBtn = true
	case HitAddButton:
		tm.hoverAddBtn = true
	case HitMenuButton:
		tm.hoverMenuBtn = true
	}

	if oldHoverIndex != tm.hoverTabIndex || oldHoverClose != tm.hoverCloseBtn ||
		oldHoverAdd != tm.hoverAddBtn || oldHoverMenu != tm.hoverMenuBtn {
		tm.Invalidate()
	}
}

// HandleMouseLeave handles WM_MOUSELEAVE
func (tm *TabManager[T]) HandleMouseLeave() {
	if tm.hoverTabIndex != -1 || tm.hoverAddBtn || tm.hoverMenuBtn {
		tm.hoverTabIndex = -1
		tm.hoverCloseBtn = false
		tm.hoverAddBtn = false
		tm.hoverMenuBtn = false
		tm.Invalidate()
	}
}

// HandleClick handles mouse click
func (tm *TabManager[T]) HandleClick(x, y int32, totalWidth int32) bool {
	if y > tm.titleBarHeight {
		return false
	}

	result, tabIndex := tm.HitTest(x, y, totalWidth)

	switch result {
	case HitAddButton:
		//TODO tm.AddTab("New Tab", nil)
		return true

	case HitMenuButton:
		if tm.OnMenuClick != nil {
			tm.OnMenuClick()
		}
		return true

	case HitCloseButton:
		if tabIndex >= 0 && tabIndex < len(tm.tabs) {
			tm.RemoveTab(tabIndex)
		}
		return true

	case HitTab:
		if tabIndex >= 0 && tabIndex < len(tm.tabs) {
			tm.SetActiveTab(tabIndex)
		}
		return true
	}

	return false
}

// Paint draws the entire title bar with tabs
func (tm *TabManager[T]) Paint(hdc HDC, width int32) {
	// Draw background
	bgColor := tm.bgColor
	bgBrush := createSolidBrush(bgColor)
	bgRect := RECT{Left: 0, Top: 0, Right: width, Bottom: tm.titleBarHeight}
	fillRect(hdc, &bgRect, bgBrush)
	deleteObject(HANDLE(bgBrush))

	// Set up drawing
	setBkMode(hdc, TRANSPARENT)
	oldFont := selectObject(hdc, HANDLE(tm.font))

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
func (tm *TabManager[T]) drawTab(hdc HDC, index int, tab *Tab[T	], totalWidth int32) {
	rect := tm.getTabRect(index, totalWidth)
	isActive := index == tm.activeTabIndex
	isHover := index == tm.hoverTabIndex && !tm.hoverCloseBtn

	// Determine colors
	var bgColor COLORREF
	var textColor COLORREF

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
		brush := createSolidBrush(bgColor)
		pen := createPen(PS_SOLID, 1, bgColor)
		oldBrush := selectObject(hdc, HANDLE(brush))
		oldPen := selectObject(hdc, HANDLE(pen))

		// Draw rounded rectangle for the tab, but don't extend beyond the separator line
		// The separator line is at titleBarHeight - 1, so limit the bottom to that
		maxBottom := tm.titleBarHeight - 1
		roundedBottom := min(rect.Bottom+tm.cornerRadius, maxBottom)

		roundRect(hdc, rect.Left, rect.Top, rect.Right, roundedBottom, tm.cornerRadius*2, tm.cornerRadius*2)

		// Fill bottom part to make only top corners rounded, but respect the separator line
		bottomRect := RECT{
			Left:   rect.Left,
			Top:    rect.Bottom - tm.cornerRadius,
			Right:  rect.Right,
			Bottom: maxBottom,
		}
		fillRect(hdc, &bottomRect, brush)

		selectObject(hdc, oldPen)
		selectObject(hdc, oldBrush)
		deleteObject(HANDLE(pen))
		deleteObject(HANDLE(brush))
	}

	// Draw tab text
	setTextColor(hdc, textColor)
	textRect := RECT{
		Left:   rect.Left + tm.tabPadding,
		Top:    rect.Top,
		Right:  rect.Right - tm.tabPadding,
		Bottom: rect.Bottom,
	}

	// Leave room for close button
	if isActive || isHover {
		textRect.Right -= tm.closeSize + 4
	}

	drawText(hdc, tab.Title, &textRect, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS|DT_NOPREFIX)

	// Draw close button if applicable
	if isActive || isHover {
		tm.drawCloseButton(hdc, rect, tm.hoverCloseBtn && index == tm.hoverTabIndex)
	}
}

// drawCloseButton draws the X button for closing a tab
func (tm *TabManager[T]) drawCloseButton(hdc HDC, tabRect RECT, isHover bool) {
	closeRect := tm.getCloseRect(tabRect)

	// Draw hover background (rounded)
	if isHover {
		brush := createSolidBrush(tm.closeBtnHoverBg)
		pen := createPen(PS_SOLID, 1, tm.closeBtnHoverBg)
		oldBrush := selectObject(hdc, HANDLE(brush))
		oldPen := selectObject(hdc, HANDLE(pen))

		// Draw circular background
		roundRect(hdc, closeRect.Left-3, closeRect.Top-3, closeRect.Right+3, closeRect.Bottom+3, 8, 8)

		selectObject(hdc, oldPen)
		selectObject(hdc, oldBrush)
		deleteObject(HANDLE(pen))
		deleteObject(HANDLE(brush))
	}

	// Draw X
	var penColor COLORREF
	if isHover {
		penColor = tm.closeBtnHover
	} else {
		penColor = tm.closeBtnColor
	}

	pen := createPen(PS_SOLID, 1, penColor)
	oldPen := selectObject(hdc, HANDLE(pen))

	padding := int32(4)
	// Draw X lines
	moveToEx(hdc, closeRect.Left+padding, closeRect.Top+padding, nil)
	lineTo(hdc, closeRect.Right-padding+1, closeRect.Bottom-padding+1)
	moveToEx(hdc, closeRect.Right-padding, closeRect.Top+padding, nil)
	lineTo(hdc, closeRect.Left+padding-1, closeRect.Bottom-padding+1)

	selectObject(hdc, oldPen)
	deleteObject(HANDLE(pen))
}

// drawAddButton draws the + button for adding tabs
func (tm *TabManager[T]) drawAddButton(hdc HDC, totalWidth int32) {
	rect := tm.getAddButtonRect(totalWidth)

	// Draw hover background
	if tm.hoverAddBtn {
		brush := createSolidBrush(tm.tabHoverColor)
		pen := createPen(PS_SOLID, 1, tm.tabHoverColor)
		oldBrush := selectObject(hdc, HANDLE(brush))
		oldPen := selectObject(hdc, HANDLE(pen))

		roundRect(hdc, rect.Left, rect.Top, rect.Right, rect.Bottom, 6, 6)

		selectObject(hdc, oldPen)
		selectObject(hdc, oldBrush)
		deleteObject(HANDLE(pen))
		deleteObject(HANDLE(brush))
	}

	// Draw + sign
	var penColor COLORREF
	if tm.hoverAddBtn {
		penColor = tm.addBtnHover
	} else {
		penColor = tm.addBtnColor
	}

	pen := createPen(PS_SOLID, 1, penColor)
	oldPen := selectObject(hdc, HANDLE(pen))

	centerX := (rect.Left + rect.Right) / 2
	centerY := (rect.Top + rect.Bottom) / 2
	size := int32(5)

	// Horizontal line
	moveToEx(hdc, centerX-size, centerY, nil)
	lineTo(hdc, centerX+size+1, centerY)

	// Vertical line
	moveToEx(hdc, centerX, centerY-size, nil)
	lineTo(hdc, centerX, centerY+size+1)

	selectObject(hdc, oldPen)
	deleteObject(HANDLE(pen))
}

// drawMenuButton draws the hamburger menu button
func (tm *TabManager[T]) drawMenuButton(hdc HDC, totalWidth int32) {
	rect := tm.getMenuButtonRect(totalWidth)

	// Draw hover background
	if tm.hoverMenuBtn {
		brush := createSolidBrush(tm.tabHoverColor)
		pen := createPen(PS_SOLID, 1, tm.tabHoverColor)
		oldBrush := selectObject(hdc, HANDLE(brush))
		oldPen := selectObject(hdc, HANDLE(pen))

		roundRect(hdc, rect.Left, rect.Top, rect.Right, rect.Bottom, 6, 6)

		selectObject(hdc, oldPen)
		selectObject(hdc, oldBrush)
		deleteObject(HANDLE(pen))
		deleteObject(HANDLE(brush))
	}

	// Draw three horizontal lines (hamburger)
	var penColor COLORREF
	if tm.hoverMenuBtn {
		penColor = tm.addBtnHover
	} else {
		penColor = tm.addBtnColor
	}

	pen := createPen(PS_SOLID, 1, penColor)
	oldPen := selectObject(hdc, HANDLE(pen))

	centerX := (rect.Left + rect.Right) / 2
	centerY := (rect.Top + rect.Bottom) / 2
	width := int32(7)
	spacing := int32(4)

	// Three horizontal lines
	for i := int32(-1); i <= 1; i++ {
		y := centerY + i*spacing
		moveToEx(hdc, centerX-width, y, nil)
		lineTo(hdc, centerX+width+1, y)
	}

	selectObject(hdc, oldPen)
	deleteObject(HANDLE(pen))
}

// drawBottomLine draws a subtle separator line at the bottom of the tab bar
func (tm *TabManager[T]) drawBottomLine(hdc HDC, totalWidth int32) {
	pen := createPen(PS_SOLID, 1, tm.borderColor)
	oldPen := selectObject(hdc, HANDLE(pen))

	// Find active tab rect to skip drawing line under it
	var activeRect *RECT
	for i := range tm.tabs {
		if i == tm.activeTabIndex {
			r := tm.getTabRect(i, totalWidth)
			activeRect = &r
			break
		}
	}

	y := tm.titleBarHeight - 1

	if activeRect == nil {
		// No active tab, draw full line
		moveToEx(hdc, 0, y, nil)
		lineTo(hdc, totalWidth, y)
	} else {
		// Draw line with gap for active tab
		if activeRect.Left > 0 {
			moveToEx(hdc, 0, y, nil)
			lineTo(hdc, activeRect.Left, y)
		}
		if activeRect.Right < totalWidth {
			moveToEx(hdc, activeRect.Right, y, nil)
			lineTo(hdc, totalWidth, y)
		}
	}

	selectObject(hdc, oldPen)
	deleteObject(HANDLE(pen))
}
