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
}

type ClickControl struct {
	Control
	id            int
	onClick       func()
	onDoubleClick func()
}

type ClickController interface {
	ID() int
	OnClick()
	OnDoubleClick()
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
		WS_CHILD|0x00000007, // BS_GROUPBOX
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

func (cc *ClickControl) ID() int {
	return cc.id
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
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|BS_PUSHBUTTON,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ButtonControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}, id: id, onClick: onClick}}
}

type CheckBoxControl struct {
	ClickControl
}

// CreateComboBox creates a dropdown combo box
func (w *Window) CreateComboBox() *ComboBoxControl {
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_COMBOBOX),
		nil,
		WS_CHILD|CBS_DROPDOWNLIST|CBS_HASSTRINGS,
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ComboBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}, id: id}}
}

// Checkbox helper functions
const (
	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1
)

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
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|0x00000003, // BS_AUTOCHECKBOX
		0, 0, 0, 0,
		w.hwnd,
		hMenu(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &CheckBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}, id: id}}
}

// ComboBox helper functions
func (control *ComboBoxControl) AddString(text string) int {
	ret := sendMessage(control.Hwnd, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return int(ret)
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
	index := c.GetCurSel()
	if index < 0 {
		return ""
	}
	length := sendMessage(c.Hwnd, CB_GETLBTEXTLEN, uintptr(index), 0)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	sendMessage(c.Hwnd, CB_GETLBTEXT, uintptr(index), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
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
	return &ListBoxControl{ClickControl: ClickControl{Control: Control{Hwnd: hwnd}, id: id}, onDoubleClick: onDoubleClick}
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
