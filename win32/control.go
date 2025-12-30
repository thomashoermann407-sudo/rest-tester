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
	Hwnd HWND
}

type Controler interface {
	GetHwnd() HWND
	Show()
	Hide()
}

func (w *Window) CreateInput(x, y, width, height int32) *Control {
	// Offset Y by tab height if tabs are enabled
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}

	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		WS_CHILD|WS_VISIBLE|WS_BORDER|ES_LEFT|ES_AUTOHSCROLL,
		x, yOffset, width, height,
		w.Hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateLabel creates a static text label
func (w *Window) CreateLabel(text string, x, y, width, height int32) *Control {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}

	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_STATIC),
		StringToUTF16Ptr(text),
		WS_CHILD|WS_VISIBLE|SS_LEFT,
		x, yOffset, width, height,
		w.Hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateGroupBox creates a group box (frame with title)
func (w *Window) CreateGroupBox(text string, x, y, width, height int32) *Control {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}

	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|WS_VISIBLE|0x00000007, // BS_GROUPBOX
		x, yOffset, width, height,
		w.Hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateMultilineEdit creates a multi-line text area
func (w *Window) CreateMultilineEdit(x, y, width, height int32, readonly bool) *Control {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}

	style := uint32(WS_CHILD | WS_VISIBLE | WS_BORDER | WS_VSCROLL | WS_HSCROLL | ES_LEFT | ES_MULTILINE | ES_AUTOVSCROLL | ES_WANTRETURN)
	if readonly {
		style |= ES_READONLY
	}

	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		style,
		x, yOffset, width, height,
		w.Hwnd,
		0,
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &Control{Hwnd: hwnd}
}

// CreateCodeEdit creates a multi-line text area with monospace font for code/JSON
func (w *Window) CreateCodeEdit(x, y, width, height int32, readonly bool) *Control {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}

	style := uint32(WS_CHILD | WS_VISIBLE | WS_BORDER | WS_VSCROLL | WS_HSCROLL | ES_LEFT | ES_MULTILINE | ES_AUTOVSCROLL | ES_WANTRETURN)
	if readonly {
		style |= ES_READONLY
	}

	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_EDIT),
		nil,
		style,
		x, yOffset, width, height,
		w.Hwnd,
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
	w.controls = append(w.controls, hwnd)
	return &Control{Hwnd: hwnd}
}

func (control *Control) GetHwnd() HWND {
	return control.Hwnd
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

func (control *Control) MoveWindow(x, y, width, height int32, repaint bool) bool {
	var repaintVal uintptr
	if repaint {
		repaintVal = 1
	}
	ret, _, _ := procMoveWindow.Call(uintptr(control.Hwnd), uintptr(x), uintptr(y), uintptr(width), uintptr(height), repaintVal)
	return ret != 0
}

type ClickControl struct {
	Control
	ID      int
	OnClick func()
}

func (w *Window) CreateButton(text string, x, y, width, height int32) *ClickControl {
	// Offset Y by tab height if tabs are enabled
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON,
		x, yOffset, width, height,
		w.Hwnd,
		HMENU(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ClickControl{Control: Control{Hwnd: hwnd}, ID: id}
}

// CreateComboBox creates a dropdown combo box
func (w *Window) CreateComboBox(x, y, width, height int32) *ClickControl {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_COMBOBOX),
		nil,
		WS_CHILD|WS_VISIBLE|CBS_DROPDOWNLIST|CBS_HASSTRINGS,
		x, yOffset, width, height,
		w.Hwnd,
		HMENU(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ClickControl{Control: Control{Hwnd: hwnd}, ID: id}
}

// CreateCheckbox creates a checkbox control
func (w *Window) CreateCheckbox(text string, x, y, width, height int32) *ClickControl {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}
	id := nextID()
	hwnd := createWindowEx(
		0,
		StringToUTF16Ptr(WC_BUTTON),
		StringToUTF16Ptr(text),
		WS_CHILD|WS_VISIBLE|0x00000003, // BS_AUTOCHECKBOX
		x, yOffset, width, height,
		w.Hwnd,
		HMENU(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ClickControl{Control: Control{Hwnd: hwnd}, ID: id}
}

// CreateListBox creates a listbox control
func (w *Window) CreateListBox(x, y, width, height int32) *ClickControl {
	yOffset := y
	if w.TabManager != nil {
		yOffset += w.TabManager.GetHeight()
	}
	id := nextID()
	hwnd := createWindowEx(
		WS_EX_CLIENTEDGE,
		StringToUTF16Ptr(WC_LISTBOX),
		nil,
		WS_CHILD|WS_VISIBLE|WS_VSCROLL|LBS_NOTIFY|LBS_HASSTRINGS|LBS_NOINTEGRALHEIGHT,
		x, yOffset, width, height,
		w.Hwnd,
		HMENU(uintptr(id)),
		getModuleHandle(nil),
		nil,
	)
	w.applyFont(hwnd)
	return &ClickControl{Control: Control{Hwnd: hwnd}, ID: id}
}

// ComboBox helper functions
func (control *Control) ComboBoxAddString(text string) int {
	ret := sendMessage(control.Hwnd, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return int(ret)
}

func (control *Control) ComboBoxSetCurSel(index int) int {
	ret := sendMessage(control.Hwnd, CB_SETCURSEL, uintptr(index), 0)
	return int(ret)
}
