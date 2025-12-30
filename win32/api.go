package win32

import (
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
	uxtheme  = syscall.NewLazyDLL("uxtheme.dll")

	procGetModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	procLoadCursorW          = user32.NewProc("LoadCursorW")
	procLoadIconW            = user32.NewProc("LoadIconW")
	procRegisterClassExW     = user32.NewProc("RegisterClassExW")
	procCreateWindowExW      = user32.NewProc("CreateWindowExW")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procDefWindowProcW       = user32.NewProc("DefWindowProcW")
	procMessageBoxW          = user32.NewProc("MessageBoxW")
	procBeginPaint           = user32.NewProc("BeginPaint")
	procEndPaint             = user32.NewProc("EndPaint")
	procGetClientRect        = user32.NewProc("GetClientRect")
	procInvalidateRect       = user32.NewProc("InvalidateRect")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procScreenToClient       = user32.NewProc("ScreenToClient")
	procMoveWindow           = user32.NewProc("MoveWindow")
	procFillRect             = user32.NewProc("FillRect")
	procDrawTextW            = user32.NewProc("DrawTextW")
	procSetBkMode            = gdi32.NewProc("SetBkMode")
	procSetTextColor         = gdi32.NewProc("SetTextColor")
	procCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject         = gdi32.NewProc("DeleteObject")
	procSelectObject         = gdi32.NewProc("SelectObject")
	procCreateFontW          = gdi32.NewProc("CreateFontW")
	procCreatePen            = gdi32.NewProc("CreatePen")
	procRoundRect            = gdi32.NewProc("RoundRect")
	procMoveToEx             = gdi32.NewProc("MoveToEx")
	procLineTo               = gdi32.NewProc("LineTo")
	procSendMessageW         = user32.NewProc("SendMessageW")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procSetWindowTextW       = user32.NewProc("SetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW     = comdlg32.NewProc("GetSaveFileNameW")
	procCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	procAppendMenuW          = user32.NewProc("AppendMenuW")
	procTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	procDestroyMenu          = user32.NewProc("DestroyMenu")
	procInitCommonControlsEx = comctl32.NewProc("InitCommonControlsEx")
	procSetWindowTheme       = uxtheme.NewProc("SetWindowTheme")
)

func getModuleHandle(name *uint16) HINSTANCE {
	ret, _, _ := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(name)))
	return HINSTANCE(ret)
}

func loadCursor(instance HINSTANCE, cursorName uintptr) HCURSOR {
	ret, _, _ := procLoadCursorW.Call(uintptr(instance), cursorName)
	return HCURSOR(ret)
}

func loadIcon(instance HINSTANCE, iconName uintptr) HICON {
	ret, _, _ := procLoadIconW.Call(uintptr(instance), iconName)
	return HICON(ret)
}

func registerClassEx(wcx *WNDCLASSEX) uint16 {
	ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wcx)))
	return uint16(ret)
}

func createWindowEx(exStyle uint32, className, windowName *uint16, style uint32, x, y, width, height int32, parent HWND, menu HMENU, instance HINSTANCE, param unsafe.Pointer) HWND {
	ret, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(param),
	)
	return HWND(ret)
}

func updateWindow(hwnd HWND) bool {
	ret, _, _ := procUpdateWindow.Call(uintptr(hwnd))
	return ret != 0
}

func getMessage(msg *MSG, hwnd HWND, msgFilterMin, msgFilterMax uint32) int32 {
	ret, _, _ := procGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return int32(ret)
}

func translateMessage(msg *MSG) bool {
	ret, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return ret != 0
}

func dispatchMessage(msg *MSG) uintptr {
	ret, _, _ := procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return ret
}

func postQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func defWindowProc(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

func MessageBox(hwnd HWND, text, caption string, type_ uint32) int32 {
	ret, _, _ := procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(caption))),
		uintptr(type_),
	)
	return int32(ret)
}

func beginPaint(hwnd HWND, ps *PAINTSTRUCT) HDC {
	ret, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return HDC(ret)
}

func endPaint(hwnd HWND, ps *PAINTSTRUCT) bool {
	ret, _, _ := procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return ret != 0
}

func getClientRect(hwnd HWND, rect *RECT) bool {
	ret, _, _ := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return ret != 0
}

func invalidateRect(hwnd HWND, rect *RECT, erase bool) bool {
	var eraseVal uintptr
	if erase {
		eraseVal = 1
	}
	ret, _, _ := procInvalidateRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)), eraseVal)
	return ret != 0
}

func getCursorPos(point *POINT) bool {
	ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(point)))
	return ret != 0
}

func screenToClient(hwnd HWND, point *POINT) bool {
	ret, _, _ := procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(point)))
	return ret != 0
}

// getCursorPosClient returns cursor position in client coordinates
func getCursorPosClient(hwnd HWND) (int32, int32) {
	var pt POINT
	if getCursorPos(&pt) {
		screenToClient(hwnd, &pt)
		return pt.X, pt.Y
	}
	return 0, 0
}

func fillRect(hdc HDC, rect *RECT, brush HBRUSH) int32 {
	ret, _, _ := procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush))
	return int32(ret)
}

func drawText(hdc HDC, text string, rect *RECT, format uint32) int32 {
	textPtr := StringToUTF16Ptr(text)
	ret, _, _ := procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(textPtr)), uintptr(len(text)), uintptr(unsafe.Pointer(rect)), uintptr(format))
	return int32(ret)
}

func setBkMode(hdc HDC, mode int32) int32 {
	ret, _, _ := procSetBkMode.Call(uintptr(hdc), uintptr(mode))
	return int32(ret)
}

func setTextColor(hdc HDC, color COLORREF) COLORREF {
	ret, _, _ := procSetTextColor.Call(uintptr(hdc), uintptr(color))
	return COLORREF(ret)
}

func createSolidBrush(color COLORREF) HBRUSH {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return HBRUSH(ret)
}

func deleteObject(obj HANDLE) bool {
	ret, _, _ := procDeleteObject.Call(uintptr(obj))
	return ret != 0
}

func selectObject(hdc HDC, obj HANDLE) HANDLE {
	ret, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return HANDLE(ret)
}

func createFont(height, width, escapement, orientation, weight int32, italic, underline, strikeOut, charSet, outputPrecision, clipPrecision, quality, pitchAndFamily uint32, faceName string) HFONT {
	ret, _, _ := procCreateFontW.Call(
		uintptr(height), uintptr(width), uintptr(escapement), uintptr(orientation), uintptr(weight),
		uintptr(italic), uintptr(underline), uintptr(strikeOut), uintptr(charSet),
		uintptr(outputPrecision), uintptr(clipPrecision), uintptr(quality), uintptr(pitchAndFamily),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(faceName))),
	)
	return HFONT(ret)
}

func createPen(style, width int32, color COLORREF) HPEN {
	ret, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return HPEN(ret)
}

func roundRect(hdc HDC, left, top, right, bottom, width, height int32) bool {
	ret, _, _ := procRoundRect.Call(uintptr(hdc), uintptr(left), uintptr(top), uintptr(right), uintptr(bottom), uintptr(width), uintptr(height))
	return ret != 0
}

func moveToEx(hdc HDC, x, y int32, point *POINT) bool {
	ret, _, _ := procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(point)))
	return ret != 0
}

func lineTo(hdc HDC, x, y int32) bool {
	ret, _, _ := procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(y))
	return ret != 0
}

// rgb creates a COLORREF from red, green, blue values
func rgb(r, g, b byte) COLORREF {
	return COLORREF(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

func sendMessage(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func comboBoxGetCurSel(hwnd HWND) int {
	ret := sendMessage(hwnd, CB_GETCURSEL, 0, 0)
	return int(ret)
}

func ComboBoxGetText(hwnd HWND) string {
	index := comboBoxGetCurSel(hwnd)
	if index < 0 {
		return ""
	}
	length := sendMessage(hwnd, CB_GETLBTEXTLEN, uintptr(index), 0)
	if length == 0 {
		return ""
	}
	buf := make([]uint16, length+1)
	sendMessage(hwnd, CB_GETLBTEXT, uintptr(index), uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

// OpenFileDialog shows a file open dialog and returns the selected file path
func OpenFileDialog(owner HWND, title, filter, defaultExt string) (string, bool) {
	var ofn OPENFILENAME
	fileNameBuf := make([]uint16, 260)

	// Convert filter string: "Description\0*.ext\0\0"
	filterBuf := make([]uint16, len(filter)+2)
	copy(filterBuf, syscall.StringToUTF16(filter))
	// Replace | with null
	for i := range filterBuf {
		if filterBuf[i] == '|' {
			filterBuf[i] = 0
		}
	}

	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	ofn.Owner = owner
	ofn.Filter = &filterBuf[0]
	ofn.File = &fileNameBuf[0]
	ofn.MaxFile = uint32(len(fileNameBuf))
	ofn.Title = StringToUTF16Ptr(title)
	ofn.DefExt = StringToUTF16Ptr(defaultExt)
	ofn.Flags = OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_EXPLORER

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", false
	}

	return syscall.UTF16ToString(fileNameBuf), true
}

// SaveFileDialog shows a file save dialog and returns the selected file path
func SaveFileDialog(owner HWND, title, filter, defaultExt, defaultName string) (string, bool) {
	var ofn OPENFILENAME
	fileNameBuf := make([]uint16, 260)

	// Copy default name to buffer
	if defaultName != "" {
		defaultNameUTF16 := syscall.StringToUTF16(defaultName)
		copy(fileNameBuf, defaultNameUTF16)
	}

	// Convert filter string: "Description\0*.ext\0\0"
	filterBuf := make([]uint16, len(filter)+2)
	copy(filterBuf, syscall.StringToUTF16(filter))
	// Replace | with null
	for i := range filterBuf {
		if filterBuf[i] == '|' {
			filterBuf[i] = 0
		}
	}

	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	ofn.Owner = owner
	ofn.Filter = &filterBuf[0]
	ofn.File = &fileNameBuf[0]
	ofn.MaxFile = uint32(len(fileNameBuf))
	ofn.Title = StringToUTF16Ptr(title)
	ofn.DefExt = StringToUTF16Ptr(defaultExt)
	ofn.Flags = OFN_OVERWRITEPROMPT | OFN_PATHMUSTEXIST | OFN_EXPLORER

	ret, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", false
	}

	return syscall.UTF16ToString(fileNameBuf), true
}

// Checkbox helper functions
const (
	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1
)

// CheckboxIsChecked returns true if the checkbox is checked
func CheckboxIsChecked(hwnd HWND) bool {
	ret := sendMessage(hwnd, BM_GETCHECK, 0, 0)
	return ret == BST_CHECKED
}

// CheckboxSetChecked sets the checkbox state
func CheckboxSetChecked(hwnd HWND, checked bool) {
	val := uintptr(0)
	if checked {
		val = BST_CHECKED
	}
	sendMessage(hwnd, BM_SETCHECK, val, 0)
}

// ListBox helper functions
func ListBoxAddString(hwnd HWND, text string) int {
	ret := sendMessage(hwnd, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(StringToUTF16Ptr(text))))
	return int(ret)
}

func ListBoxGetCurSel(hwnd HWND) int {
	ret := sendMessage(hwnd, LB_GETCURSEL, 0, 0)
	return int(ret)
}

func ListBoxSetCurSel(hwnd HWND, index int) int {
	ret := sendMessage(hwnd, LB_SETCURSEL, uintptr(index), 0)
	return int(ret)
}

func ListBoxGetCount(hwnd HWND) int {
	ret := sendMessage(hwnd, LB_GETCOUNT, 0, 0)
	return int(ret)
}

func ListBoxResetContent(hwnd HWND) {
	sendMessage(hwnd, LB_RESETCONTENT, 0, 0)
}

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

// InitCommonControlsEx structure
type INITCOMMONCONTROLSEX struct {
	Size uint32
	ICC  uint32
}

// Common control classes
const (
	ICC_LISTVIEW_CLASSES   = 0x00000001
	ICC_TREEVIEW_CLASSES   = 0x00000002
	ICC_BAR_CLASSES        = 0x00000004
	ICC_TAB_CLASSES        = 0x00000008
	ICC_UPDOWN_CLASS       = 0x00000010
	ICC_PROGRESS_CLASS     = 0x00000020
	ICC_HOTKEY_CLASS       = 0x00000040
	ICC_ANIMATE_CLASS      = 0x00000080
	ICC_WIN95_CLASSES      = 0x000000FF
	ICC_DATE_CLASSES       = 0x00000100
	ICC_USEREX_CLASSES     = 0x00000200
	ICC_COOL_CLASSES       = 0x00000400
	ICC_INTERNET_CLASSES   = 0x00000800
	ICC_PAGESCROLLER_CLASS = 0x00001000
	ICC_NATIVEFNTCTL_CLASS = 0x00002000
	ICC_STANDARD_CLASSES   = 0x00004000
	ICC_LINK_CLASS         = 0x00008000
)

// initCommonControls initializes common controls
func initCommonControls(classes uint32) bool {
	icc := INITCOMMONCONTROLSEX{
		Size: uint32(unsafe.Sizeof(INITCOMMONCONTROLSEX{})),
		ICC:  classes,
	}
	ret, _, _ := procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&icc)))
	return ret != 0
}

// SetWindowTheme sets the visual theme for a window/control
// Pass empty strings to use default theme, or specific theme names
func setWindowTheme(hwnd HWND, appName, idList string) error {
	var appNamePtr, idListPtr *uint16
	if appName != "" {
		appNamePtr = StringToUTF16Ptr(appName)
	}
	if idList != "" {
		idListPtr = StringToUTF16Ptr(idList)
	}
	ret, _, _ := procSetWindowTheme.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(appNamePtr)),
		uintptr(unsafe.Pointer(idListPtr)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}
