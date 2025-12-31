package win32

import "unsafe"

// Popup menu constants
const (
	MF_STRING       = 0x00000000
	MF_SEPARATOR    = 0x00000800
	TPM_LEFTALIGN   = 0x0000
	TPM_TOPALIGN    = 0x0000
	TPM_RETURNCMD   = 0x0100
	TPM_RIGHTBUTTON = 0x0002
)

// PopupMenu represents a popup context menu
type PopupMenu struct {
	handle HANDLE
}

// CreatePopupMenu creates a new popup menu
func CreatePopupMenu() *PopupMenu {
	ret, _, _ := procCreatePopupMenu.Call()
	if ret == 0 {
		return nil
	}
	return &PopupMenu{handle: HANDLE(ret)}
}

// AddItem adds a menu item
func (m *PopupMenu) AddItem(id int, text string) {
	procAppendMenuW.Call(
		uintptr(m.handle),
		MF_STRING,
		uintptr(id),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(text))),
	)
}

// AddSeparator adds a separator line
func (m *PopupMenu) AddSeparator() {
	procAppendMenuW.Call(
		uintptr(m.handle),
		MF_SEPARATOR,
		0,
		0,
	)
}

// Show displays the menu at cursor position and returns selected item ID (0 if cancelled)
func (m *PopupMenu) Show(hwnd HWND) int {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	ret, _, _ := procTrackPopupMenu.Call(
		uintptr(m.handle),
		TPM_LEFTALIGN|TPM_TOPALIGN|TPM_RETURNCMD|TPM_RIGHTBUTTON,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(hwnd),
		0,
	)
	return int(ret)
}

// ShowAt displays the menu at specified position
func (m *PopupMenu) ShowAt(hwnd HWND, x, y int32) int {
	ret, _, _ := procTrackPopupMenu.Call(
		uintptr(m.handle),
		TPM_LEFTALIGN|TPM_TOPALIGN|TPM_RETURNCMD|TPM_RIGHTBUTTON,
		uintptr(x),
		uintptr(y),
		0,
		uintptr(hwnd),
		0,
	)
	return int(ret)
}

// Destroy destroys the menu
func (m *PopupMenu) Destroy() {
	procDestroyMenu.Call(uintptr(m.handle))
}
