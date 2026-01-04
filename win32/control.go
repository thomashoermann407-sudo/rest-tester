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
	CreateTreeView(onDoubleClick func(*TreeViewControl)) *TreeViewControl
	CreateTabControl() *TabControlControl
	CreateInput() *Control
	CreateLabel(text string) *Control
	CreateGroupBox(text string) *Control
	CreateMultilineEdit(readonly bool) *Control
	CreateCodeEdit(readonly bool) *Control
	MessageBox(title, message string) int32
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
	length, _, _ := procGetWindowTextLengthW.Call(uintptr(control.Hwnd))
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(control.Hwnd), uintptr(unsafe.Pointer(&buf[0])), length+1)
	return syscall.UTF16ToString(buf)
}

func (control *Control) SetText(text string) bool {
	ret, _, _ := procSetWindowTextW.Call(uintptr(control.Hwnd), uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return ret != 0
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
	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_LISTBOX),
		nil,
		WS_CHILD|WS_VSCROLL|LBS_NOTIFY|LBS_HASSTRINGS|LBS_NOINTEGRALHEIGHT,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(nextID())),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ListBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}}, onDoubleClick: onDoubleClick}
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

type NMHDR struct {
	HwndFrom hWnd
	IdFrom   uintptr
	Code     uint32
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
		hMenu(uintptr(nextID())),
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
