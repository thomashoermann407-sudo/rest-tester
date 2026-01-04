package win32

import (
	"syscall"
	"unsafe"
)

var hwndIdCounter = 100

// nextID returns the next available control ID and increments the counter.
func nextID() int {
	hwndIdCounter++
	return hwndIdCounter
}

type Control struct {
	Hwnd hWnd
}

type Controller interface {
	Show()
	Hide()
	ID() int
}

type ClickControl struct {
	Control
	onClick       func()
	onDoubleClick func()
}

type ClickController interface {
	OnClick()
	OnDoubleClick()
}
type ControlFactory interface {
	CreateButton(text string, onClick func()) *ButtonControl
	CreateCheckbox(text string) *CheckBoxControl
	CreateComboBox() *ComboBoxControl
	CreateEditableComboBox() *ComboBoxControl
	CreateListBox(onDoubleClick func(*ListBoxControl)) *ListBoxControl
	CreateListView() *ListViewControl
	CreateTreeView(onDoubleClick func(*TreeViewControl)) *TreeViewControl
	CreateTabControl() *TabControlControl
	CreateInput() *Control
	CreateLabel(text string) *Control
	CreateGroupBox(text string) *Control
	CreateMultilineEdit(readonly bool) *Control
	CreateCodeEdit(readonly bool) *Control
	MessageBox(title, message string) int32
	MessageBoxYesNo(title, message string) int32
	OpenFileDialog(title, filter, defaultExt string) (string, bool)
	SaveFileDialog(title, filter, defaultExt, defaultName string) (string, bool)
	CreatePopupMenu() *PopupMenu
	PostUICallback(callback func())
}

func (w *Window) CreateInput() *Control {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		WS_CHILD|WS_BORDER|ES_LEFT|ES_AUTOHSCROLL,
		0, 0, 0, 0,
		w.hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateLabel creates a static text label
func (w *Window) CreateLabel(text string) *Control {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_STATIC),
		StringToUTF16Ptr(text),
		WS_CHILD|SS_LEFT,
		0, 0, 0, 0,
		w.hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateGroupBox creates a group box (frame with title)
func (w *Window) CreateGroupBox(text string) *Control {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|BS_GROUPBOX,
		0, 0, 0, 0,
		w.hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateMultilineEdit creates a multi-line text area
func (w *Window) CreateMultilineEdit(readonly bool) *Control {
	style := uint32(WS_CHILD | WS_BORDER | WS_VSCROLL | WS_HSCROLL | ES_LEFT | ES_MULTILINE | ES_AUTOVSCROLL | ES_WANTRETURN)
	if readonly {
		style |= ES_READONLY
	}

	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		style,
		0, 0, 0, 0,
		w.hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateCodeEdit creates a multi-line text area with monospace font for code/JSON
func (w *Window) CreateCodeEdit(readonly bool) *Control {
	style := uint32(WS_CHILD | WS_BORDER | WS_VSCROLL | WS_HSCROLL | ES_LEFT | ES_MULTILINE | ES_AUTOVSCROLL | ES_WANTRETURN)
	if readonly {
		style |= ES_READONLY
	}

	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		style,
		0, 0, 0, 0,
		w.hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	// Apply monospace font
	if w.monoFont != 0 {
		sendMessage(hwnd, WM_SETFONT, uintptr(w.monoFont), 1)
	}
	// Enable modern visual styles
	setWindowTheme(hwnd, "", "")
	return &Control{Hwnd: hwnd}
}

func (control *Control) Show() {
	procShowWindow.Call(uintptr(control.Hwnd), SW_SHOW)
}

func (control *Control) Hide() {
	procShowWindow.Call(uintptr(control.Hwnd), SW_HIDE)
}

func (control *Control) ID() int {
	return int(getWindowLongPtr(control.Hwnd, GWLP_ID))
}

func (control *Control) GetText() string {
	return getText(control.Hwnd)
}

func getText(hwnd hWnd) string {
	length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), length+1)
	return syscall.UTF16ToString(buf)
}

func (control *Control) SetText(text string) bool {
	ret, _, _ := procSetWindowTextW.Call(uintptr(control.Hwnd), uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return ret != 0
}

func (control *Control) SetReadOnly(readonly bool) {
	var wParam uintptr
	if readonly {
		wParam = 1
	} else {
		wParam = 0
	}
	sendMessage(control.Hwnd, EM_SETREADONLY, wParam, 0)
}

func (control *Control) MoveWindow(x, y, width, height int32) bool {
	ret, _, _ := procMoveWindow.Call(uintptr(control.Hwnd), uintptr(x), uintptr(y), uintptr(width), uintptr(height), 1)
	return ret != 0
}

func (cc *ClickControl) OnClick() {
	if cc.onClick != nil {
		cc.onClick()
	}
}
func (cc *ClickControl) OnDoubleClick() {
	if cc.onDoubleClick != nil {
		cc.onDoubleClick()
	}
}

type ButtonControl struct {
	ClickControl
}

func (w *Window) CreateButton(text string, onClick func()) *ButtonControl {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|BS_PUSHBUTTON,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(nextID())),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ButtonControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}, onClick: onClick}}
}

type CheckBoxControl struct {
	ClickControl
}

// CreateComboBox creates a dropdown combo box
func (w *Window) CreateComboBox() *ComboBoxControl {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_COMBOBOX),
		nil,
		WS_CHILD|CBS_DROPDOWNLIST|CBS_HASSTRINGS,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(nextID())),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ComboBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}}}
}

// CreateEditableComboBox creates an editable dropdown combo box
func (w *Window) CreateEditableComboBox() *ComboBoxControl {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_COMBOBOX),
		nil,
		WS_CHILD|CBS_DROPDOWN|CBS_HASSTRINGS|ES_AUTOHSCROLL,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(nextID())),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ComboBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}}}
}

// CheckboxIsChecked returns true if the checkbox is checked
func (c *CheckBoxControl) IsChecked() bool {
	ret := sendMessage(c.Hwnd, BM_GETCHECK, 0, 0)
	return ret == BST_CHECKED
}

// CheckboxSetChecked sets the checkbox state
func (c *CheckBoxControl) SetChecked(checked bool) {
	val := uintptr(0)
	if checked {
		val = BST_CHECKED
	}
	sendMessage(c.Hwnd, BM_SETCHECK, val, 0)
}

type ComboBoxControl struct {
	ClickControl
}

// CreateCheckbox creates a checkbox control
func (w *Window) CreateCheckbox(text string) *CheckBoxControl {
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|BS_AUTOCHECKBOX,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(nextID())),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &CheckBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}}}
}

// ComboBox helper functions
func (control *ComboBoxControl) AddString(text string) int {
	ret := sendMessage(control.Hwnd, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return int(ret)
}

func (control *ComboBoxControl) Clear() {
	sendMessage(control.Hwnd, CB_RESETCONTENT, 0, 0)
}

func (control *ComboBoxControl) SetCurSel(index int) int {
	ret := sendMessage(control.Hwnd, CB_SETCURSEL, uintptr(index), 0)
	return int(ret)
}

func (control *ComboBoxControl) GetCurSel() int {
	ret := sendMessage(control.Hwnd, CB_GETCURSEL, 0, 0)
	return int(ret)
}

func (c *ComboBoxControl) GetText() string {
	// For editable combo boxes, get the text from the edit control
	length := sendMessage(c.Hwnd, WM_GETTEXTLENGTH, 0, 0)
	if length > 0 {
		buf := make([]uint16, length+1)
		sendMessage(c.Hwnd, WM_GETTEXT, uintptr(length+1), uintptr(unsafe.Pointer(&buf[0])))
		return syscall.UTF16ToString(buf)
	}

	// Fallback: get selected item text from list
	index := c.GetCurSel()
	if index < 0 {
		return ""
	}
	length = sendMessage(c.Hwnd, CB_GETLBTEXTLEN, uintptr(index), 0)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	sendMessage(c.Hwnd, CB_GETLBTEXT, uintptr(index), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func (c *ComboBoxControl) SetText(text string) {
	sendMessage(c.Hwnd, WM_SETTEXT, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
}

type ListBoxControl struct {
	ClickControl
	onDoubleClick func(*ListBoxControl)
}

// CreateListBox creates a listbox control
func (w *Window) CreateListBox(onDoubleClick func(*ListBoxControl)) *ListBoxControl {
	id := nextID()
	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_LISTBOX),
		nil,
		WS_CHILD|WS_VSCROLL|LBS_NOTIFY|LBS_HASSTRINGS|LBS_NOINTEGRALHEIGHT,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	listBox := &ListBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}}, onDoubleClick: onDoubleClick}
	w.controls[id] = listBox
	return listBox
}

func (l *ListBoxControl) OnDoubleClick() {
	if l.onDoubleClick != nil {
		l.onDoubleClick(l)
	}
}

// ListBox helper functions
func (l *ListBoxControl) AddString(text string) int {
	ret := sendMessage(l.Hwnd, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return int(ret)
}

func (l *ListBoxControl) GetCurSel() int {
	ret := sendMessage(l.Hwnd, LB_GETCURSEL, 0, 0)
	return int(ret)
}

func (l *ListBoxControl) SetCurSel(index int) int {
	ret := sendMessage(l.Hwnd, LB_SETCURSEL, uintptr(index), 0)
	return int(ret)
}

func (l *ListBoxControl) GetCount() int {
	ret := sendMessage(l.Hwnd, LB_GETCOUNT, 0, 0)
	return int(ret)
}

func (l *ListBoxControl) ResetContent() {
	sendMessage(l.Hwnd, LB_RESETCONTENT, 0, 0)
}

type ListViewControl struct {
	ClickControl
	onSelChange  func(*ListViewControl)
	editControl  hWnd
	callbackPtr  uintptr
	editingRow   int
	editingCol   int
	isEditing    bool
	onEditEnd    func(row, col int, newText string)
	parentWindow *Window
}

// CreateListView creates a listview control in report (details) mode
func (w *Window) CreateListView() *ListViewControl {
	id := nextID()
	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_LISTVIEW),
		nil,
		WS_CHILD|WS_VISIBLE|LVS_REPORT|LVS_SINGLESEL|LVS_SHOWSELALWAYS,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)

	// Set extended styles for better appearance
	sendMessage(hwnd, LVM_SETEXTENDEDLISTVIEWSTYLE, 0, LVS_EX_FULLROWSELECT|LVS_EX_GRIDLINES)

	listView := &ListViewControl{
		ClickControl: ClickControl{Control: Control{Hwnd: hwnd}},
		editingRow:   -1,
		editingCol:   -1,
		isEditing:    false,
		parentWindow: w,
	}
	w.controls[id] = listView
	return listView
}

func (lv *ListViewControl) SetOnSelChange(handler func(*ListViewControl)) {
	lv.onSelChange = handler
}

func (lv *ListViewControl) OnSelChange() {
	if lv.onSelChange != nil {
		lv.onSelChange(lv)
	}
}

// InsertColumn adds a column to the listview
func (lv *ListViewControl) InsertColumn(index int, text string, width int32) int {
	textPtr := StringToUTF16Ptr(text)
	col := LVCOLUMNW{
		Mask:    LVCF_TEXT | LVCF_WIDTH,
		PszText: textPtr,
		Cx:      width,
	}
	ret := sendMessage(lv.Hwnd, LVM_INSERTCOLUMNW, uintptr(index), uintptr(unsafe.Pointer(&col)))
	return int(ret)
}

// InsertItem adds a new row to the listview
func (lv *ListViewControl) InsertItem(index int, text string, lParam uintptr) int {
	textPtr := StringToUTF16Ptr(text)
	item := LVITEMW{
		Mask:    LVIF_TEXT | LVIF_PARAM,
		IItem:   int32(index),
		PszText: textPtr,
		LParam:  lParam,
	}
	ret := sendMessage(lv.Hwnd, LVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&item)))
	return int(ret)
}

// SetItemText sets the text of a subitem (column)
func (lv *ListViewControl) SetItemText(row, col int, text string) {
	textPtr := StringToUTF16Ptr(text)
	item := LVITEMW{
		ISubItem: int32(col),
		PszText:  textPtr,
	}
	sendMessage(lv.Hwnd, LVM_SETITEMTEXTW, uintptr(row), uintptr(unsafe.Pointer(&item)))
}

// GetSelectedIndex returns the index of the selected item (-1 if none)
func (lv *ListViewControl) GetSelectedIndex() int {
	ret := sendMessage(lv.Hwnd, LVM_GETNEXTITEM, ^uintptr(0), LVNI_SELECTED)
	return int(int32(ret))
}

// SetCurSel selects an item by index
func (lv *ListViewControl) SetCurSel(index int) {
	// First, deselect all items
	item := LVITEMW{
		StateMask: LVIS_SELECTED | LVIS_FOCUSED,
		State:     0,
	}
	sendMessage(lv.Hwnd, LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&item)))

	// Then select and focus the specified item
	item.State = LVIS_SELECTED | LVIS_FOCUSED
	sendMessage(lv.Hwnd, LVM_SETITEMSTATE, uintptr(index), uintptr(unsafe.Pointer(&item)))

	// Ensure the item is visible
	sendMessage(lv.Hwnd, LVM_ENSUREVISIBLE, uintptr(index), 0)
}

// GetItemCount returns the number of items in the listview
func (lv *ListViewControl) GetItemCount() int {
	ret := sendMessage(lv.Hwnd, LVM_GETITEMCOUNT, 0, 0)
	return int(ret)
}

// GetItemLParam gets the user data associated with an item
func (lv *ListViewControl) GetItemLParam(index int) uintptr {
	item := LVITEMW{
		Mask:  LVIF_PARAM,
		IItem: int32(index),
	}
	sendMessage(lv.Hwnd, LVM_GETITEMW, 0, uintptr(unsafe.Pointer(&item)))
	return item.LParam
}

// DeleteItem removes an item by index
func (lv *ListViewControl) DeleteItem(index int) bool {
	ret := sendMessage(lv.Hwnd, LVM_DELETEITEM, uintptr(index), 0)
	return ret != 0
}

// DeleteAllItems removes all items
func (lv *ListViewControl) DeleteAllItems() bool {
	ret := sendMessage(lv.Hwnd, LVM_DELETEALLITEMS, 0, 0)
	return ret != 0
}

// SetOnEditEnd sets the callback for when editing is complete
func (lv *ListViewControl) SetOnEditEnd(handler func(row, col int, newText string)) {
	lv.onEditEnd = handler
}

// GetItemText retrieves the text of a cell
func (lv *ListViewControl) GetItemText(row, col int) string {
	buf := make([]uint16, 260)
	item := LVITEMW{
		ISubItem:   int32(col),
		PszText:    &buf[0],
		CchTextMax: 260,
	}
	sendMessage(lv.Hwnd, LVM_GETITEMTEXTW, uintptr(row), uintptr(unsafe.Pointer(&item)))
	return syscall.UTF16ToString(buf)
}

// HitTestEx performs a hit test to determine which cell was clicked
func (lv *ListViewControl) HitTestEx(x, y int32) (row, col int) {
	var hti LVHITTESTINFO
	hti.Pt.X = x
	hti.Pt.Y = y
	sendMessage(lv.Hwnd, LVM_SUBITEMHITTEST, 0, uintptr(unsafe.Pointer(&hti)))
	return int(hti.IItem), int(hti.ISubItem)
}

// GetSubItemRect gets the rectangle for a specific cell
func (lv *ListViewControl) GetSubItemRect(row, col int) rect {
	var rc rect
	rc.Top = int32(col)
	rc.Left = LVIR_LABEL
	sendMessage(lv.Hwnd, LVM_GETSUBITEMRECT, uintptr(row), uintptr(unsafe.Pointer(&rc)))
	return rc
}

// StartEdit begins in-place editing of a cell.
// Note: We destroy and recreate the edit control for each edit session rather than
// reusing a single control. This approach:
// - Ensures proper cleanup of subclassing/callbacks
// - Avoids potential issues with text/state not being properly reset
// - Is simpler than managing show/hide/reposition logic
// - Has negligible performance impact (editing is infrequent)
func (lv *ListViewControl) StartEdit(row, col int) {
	if lv.isEditing {
		lv.EndEdit(true) // Save current edit
	}

	if row < 0 || col < 0 {
		return
	}

	// Get the cell rectangle
	cellRect := lv.GetSubItemRect(row, col)

	// Get current text
	currentText := lv.GetItemText(row, col)

	// Create edit control if it doesn't exist
	if lv.editControl == 0 {
		id := nextID()
		lv.editControl = createWindowEx(
			0,
			StringToUTF16Ptr(WC_EDIT),
			StringToUTF16Ptr(currentText),
			WS_CHILD|ES_LEFT|ES_AUTOHSCROLL|WS_VISIBLE,
			cellRect.Left+2, cellRect.Top+2, cellRect.Right-cellRect.Left-4, cellRect.Bottom-cellRect.Top-2,
			lv.Hwnd,
			hMenu(uintptr(id)),
			getModuleHandle(nil),
			nil,
		)
		lv.parentWindow.applyFont(lv.editControl)
		callback := func(hwnd hWnd, msg uintptr, wParam, lParam uintptr) uintptr {
			result := callWindowProc(lv.callbackPtr, hwnd, msg, wParam, lParam)
			switch msg {
			case WM_KEYDOWN:
				lv.HandleKeyDown(wParam)
			case WM_KILLFOCUS:
				lv.EndEdit(true)
			}
			return result
		}
		lv.callbackPtr = setWindowLongPtr(lv.editControl, GWLP_WNDPROC, syscall.NewCallback(callback))

	}

	// Set focus to edit control and select all text
	procSetFocus.Call(uintptr(lv.editControl))
	sendMessage(lv.editControl, EM_SETSEL, 0, ^uintptr(0))

	lv.editingRow = row
	lv.editingCol = col
	lv.isEditing = true
}

// EndEdit completes the editing and updates the cell
func (lv *ListViewControl) EndEdit(save bool) {
	if !lv.isEditing || lv.editControl == 0 {
		return
	}

	if save {
		newText := getText(lv.editControl)
		lv.SetItemText(lv.editingRow, lv.editingCol, newText)

		// Call the callback if set
		if lv.onEditEnd != nil {
			lv.onEditEnd(lv.editingRow, lv.editingCol, newText)
		}
	}

	destroyWindow(lv.editControl)
	lv.editControl = 0
	lv.isEditing = false
	lv.editingRow = -1
	lv.editingCol = -1

	// Return focus to ListView
	procSetFocus.Call(uintptr(lv.Hwnd))
}

// IsEditing returns whether the ListView is currently in edit mode
func (lv *ListViewControl) IsEditing() bool {
	return lv.isEditing
}

// HandleKeyDown handles key events during editing.
// This is called from the edit control's subclassed window procedure,
// not from the main window message loop.
// Note: The isEditing check is technically redundant since this is only
// called when the edit control exists and has focus, but kept for safety.
func (lv *ListViewControl) HandleKeyDown(key uintptr) bool {
	if !lv.isEditing {
		return false
	}

	switch key {
	case VK_RETURN:
		lv.EndEdit(true)
		return true
	case VK_ESCAPE:
		lv.EndEdit(false)
		return true
	}
	return false
}

type NMHDR struct {
	HwndFrom hWnd
	IdFrom   uintptr
	Code     uint32
}

type LVCOLUMNW struct {
	Mask       uint32
	Fmt        int32
	Cx         int32
	PszText    *uint16
	CchTextMax int32
	ISubItem   int32
	IImage     int32
	IOrder     int32
}

type LVITEMW struct {
	Mask       uint32
	IItem      int32
	ISubItem   int32
	State      uint32
	StateMask  uint32
	PszText    *uint16
	CchTextMax int32
	IImage     int32
	LParam     uintptr
	IIndent    int32
}

type LVHITTESTINFO struct {
	Pt       point
	Flags    uint32
	IItem    int32
	ISubItem int32
	IGroup   int32
}

type TVHITTESTINFO struct {
	Pt    point
	Flags uint32
	HItem uintptr
}

type TVITEMW struct {
	Mask           uint32
	HItem          uintptr
	State          uint32
	StateMask      uint32
	PszText        *uint16
	CchTextMax     int32
	IImage         int32
	ISelectedImage int32
	CChildren      int32
	LParam         uintptr
}

type TVINSERTSTRUCTW struct {
	HParent      uintptr
	HInsertAfter uintptr
	Item         TVITEMW
}

type TreeViewControl struct {
	ClickControl
	onDoubleClick func(*TreeViewControl)
	onSelChange   func(*TreeViewControl)
	onRightClick  func(*TreeViewControl, uintptr) // Called with item handle
}

// CreateTreeView creates a tree view control
func (w *Window) CreateTreeView(onDoubleClick func(*TreeViewControl)) *TreeViewControl {
	id := nextID()
	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_TREEVIEW),
		nil,
		WS_CHILD|WS_VISIBLE|WS_VSCROLL|TVS_HASBUTTONS|TVS_HASLINES|TVS_LINESATROOT|TVS_SHOWSELALWAYS,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)), // Use the same ID for the control
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	treeView := &TreeViewControl{
		ClickControl:  ClickControl{Control: Control{Hwnd: hwnd}},
		onDoubleClick: onDoubleClick,
	}
	// Register control for event handling
	w.controls[id] = treeView
	return treeView
}

// SetOnRightClick sets the right-click handler
func (t *TreeViewControl) SetOnRightClick(handler func(*TreeViewControl, uintptr)) {
	t.onRightClick = handler
}

func (t *TreeViewControl) OnDoubleClick() {
	if t.onDoubleClick != nil {
		t.onDoubleClick(t)
	}
}

func (t *TreeViewControl) OnRightClick(hItem uintptr) {
	if t.onRightClick != nil {
		t.onRightClick(t, hItem)
	}
}

func (t *TreeViewControl) OnSelChange() {
	if t.onSelChange != nil {
		t.onSelChange(t)
	}
}

// InsertItem inserts a new item into the tree view
// parentItem: handle to parent item (use TVI_ROOT for root items)
// insertAfter: where to insert (use TVI_LAST to append)
// text: display text for the item
// lParam: user data associated with the item
func (t *TreeViewControl) InsertItem(parentItem, insertAfter uintptr, text string, lParam uintptr) uintptr {
	textPtr := StringToUTF16Ptr(text)
	tvins := TVINSERTSTRUCTW{
		HParent:      parentItem,
		HInsertAfter: insertAfter,
		Item: TVITEMW{
			Mask:    TVIF_TEXT | TVIF_PARAM,
			PszText: textPtr,
			LParam:  lParam,
		},
	}
	return sendMessage(t.Hwnd, TVM_INSERTITEMW, 0, uintptr(unsafe.Pointer(&tvins)))
}

// DeleteItem deletes an item from the tree view
func (t *TreeViewControl) DeleteItem(hItem uintptr) bool {
	ret := sendMessage(t.Hwnd, TVM_DELETEITEM, 0, hItem)
	return ret != 0
}

// DeleteAllItems deletes all items from the tree view
func (t *TreeViewControl) DeleteAllItems() bool {
	return t.DeleteItem(TVI_ROOT)
}

// GetSelection returns the currently selected item handle
func (t *TreeViewControl) GetSelection() uintptr {
	return sendMessage(t.Hwnd, TVM_GETNEXTITEM, TVGN_CARET, 0)
}

// GetRoot returns the first root item
func (t *TreeViewControl) GetRoot() uintptr {
	return sendMessage(t.Hwnd, TVM_GETNEXTITEM, TVGN_ROOT, 0)
}

// GetChild returns the first child of an item
func (t *TreeViewControl) GetChild(hItem uintptr) uintptr {
	return sendMessage(t.Hwnd, TVM_GETNEXTITEM, TVGN_CHILD, hItem)
}

// GetNextSibling returns the next sibling of an item
func (t *TreeViewControl) GetNextSibling(hItem uintptr) uintptr {
	return sendMessage(t.Hwnd, TVM_GETNEXTITEM, TVGN_NEXT, hItem)
}

// GetParent returns the parent of an item
func (t *TreeViewControl) GetParent(hItem uintptr) uintptr {
	return sendMessage(t.Hwnd, TVM_GETNEXTITEM, TVGN_PARENT, hItem)
}

// GetItemLParam retrieves the user data (lParam) associated with an item
func (t *TreeViewControl) GetItemLParam(hItem uintptr) uintptr {
	var item TVITEMW
	item.Mask = TVIF_PARAM | TVIF_HANDLE
	item.HItem = hItem
	sendMessage(t.Hwnd, TVM_GETITEMW, 0, uintptr(unsafe.Pointer(&item)))
	return item.LParam
}

// GetItemText retrieves the text of an item
func (t *TreeViewControl) GetItemText(hItem uintptr) string {
	buf := make([]uint16, 260)
	var item TVITEMW
	item.Mask = TVIF_TEXT | TVIF_HANDLE
	item.HItem = hItem
	item.PszText = &buf[0]
	item.CchTextMax = 260
	sendMessage(t.Hwnd, TVM_GETITEMW, 0, uintptr(unsafe.Pointer(&item)))
	return syscall.UTF16ToString(buf)
}

// HitTest returns the item handle at the specified coordinates
func (t *TreeViewControl) HitTest(x, y int32) uintptr {
	var hti TVHITTESTINFO
	hti.Pt.X = x
	hti.Pt.Y = y
	sendMessage(t.Hwnd, TVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))
	return hti.HItem
}

type TCITEMW struct {
	Mask        uint32
	DwState     uint32
	DwStateMask uint32
	PszText     *uint16
	CchTextMax  int32
	IImage      int32
	LParam      uintptr
}

type TabControlControl struct {
	ClickControl
	onSelChange func(*TabControlControl)
}

// CreateTabControl creates a tab control
func (w *Window) CreateTabControl() *TabControlControl {
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_TABCONTROL),
		nil,
		WS_CHILD|WS_VISIBLE|TCS_TABS,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	tabControl := &TabControlControl{
		ClickControl: ClickControl{Control: Control{Hwnd: hwnd}},
	}
	// Register control for event handling
	w.controls[id] = tabControl
	return tabControl
}

// SetOnSelChange sets the selection change handler
func (tc *TabControlControl) SetOnSelChange(handler func(*TabControlControl)) {
	tc.onSelChange = handler
}

func (tc *TabControlControl) OnSelChange() {
	if tc.onSelChange != nil {
		tc.onSelChange(tc)
	}
}

// InsertItem inserts a new tab item
func (tc *TabControlControl) InsertItem(index int, text string, lParam uintptr) int {
	textPtr := StringToUTF16Ptr(text)
	item := TCITEMW{
		Mask:    TCIF_TEXT | TCIF_PARAM,
		PszText: textPtr,
		LParam:  lParam,
	}
	ret := sendMessage(tc.Hwnd, TCM_INSERTITEMW, uintptr(index), uintptr(unsafe.Pointer(&item)))
	return int(ret)
}

// DeleteItem deletes a tab item by index
func (tc *TabControlControl) DeleteItem(index int) bool {
	ret := sendMessage(tc.Hwnd, TCM_DELETEITEM, uintptr(index), 0)
	return ret != 0
}

// DeleteAllItems deletes all tab items
func (tc *TabControlControl) DeleteAllItems() bool {
	ret := sendMessage(tc.Hwnd, TCM_DELETEALLITEMS, 0, 0)
	return ret != 0
}

// GetCurSel returns the currently selected tab index
func (tc *TabControlControl) GetCurSel() int {
	ret := sendMessage(tc.Hwnd, TCM_GETCURSEL, 0, 0)
	return int(ret)
}

// SetCurSel sets the currently selected tab
func (tc *TabControlControl) SetCurSel(index int) int {
	ret := sendMessage(tc.Hwnd, TCM_SETCURSEL, uintptr(index), 0)
	return int(ret)
}

// GetItemCount returns the number of tabs
func (tc *TabControlControl) GetItemCount() int {
	ret := sendMessage(tc.Hwnd, TCM_GETITEMCOUNT, 0, 0)
	return int(ret)
}
