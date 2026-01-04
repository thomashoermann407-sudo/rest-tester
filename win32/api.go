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
	procDestroyWindow        = user32.NewProc("DestroyWindow")
	procShowWindow           = user32.NewProc("ShowWindow")
	procUpdateWindow         = user32.NewProc("UpdateWindow")
	procGetMessageW          = user32.NewProc("GetMessageW")
	procTranslateMessage     = user32.NewProc("TranslateMessage")
	procDispatchMessageW     = user32.NewProc("DispatchMessageW")
	procPostQuitMessage      = user32.NewProc("PostQuitMessage")
	procPostMessageW         = user32.NewProc("PostMessageW")
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
	procGetWindowLongPtr     = user32.NewProc("GetWindowLongPtrW")
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

func getModuleHandle(name *uint16) hInstance {
	ret, _, _ := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(name)))
	return hInstance(ret)
}

func loadCursor(instance hInstance, cursorName uintptr) hCursor {
	ret, _, _ := procLoadCursorW.Call(uintptr(instance), cursorName)
	return hCursor(ret)
}

func loadIcon(instance hInstance, iconName uintptr) hIcon {
	ret, _, _ := procLoadIconW.Call(uintptr(instance), iconName)
	return hIcon(ret)
}

func registerClassEx(wcx *wndClassEx) uint16 {
	ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wcx)))
	return uint16(ret)
}

func createWindowEx(exStyle uint32, className, windowName *uint16, style uint32, x, y, width, height int32, parent hWnd, menu hMenu, instance hInstance, param unsafe.Pointer) hWnd {
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
	return hWnd(ret)
}

func destroyWindow(hwnd hWnd) bool {
	ret, _, _ := procDestroyWindow.Call(uintptr(hwnd))
	return ret != 0
}

func updateWindow(hwnd hWnd) bool {
	ret, _, _ := procUpdateWindow.Call(uintptr(hwnd))
	return ret != 0
}

func getMessage(msg *msg, hwnd hWnd, msgFilterMin, msgFilterMax uint32) int32 {
	ret, _, _ := procGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return int32(ret)
}

func translateMessage(msg *msg) bool {
	ret, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return ret != 0
}

func dispatchMessage(msg *msg) uintptr {
	ret, _, _ := procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return ret
}

func postQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func postMessage(hwnd hWnd, msg uint32, wParam, lParam uintptr) bool {
	ret, _, _ := procPostMessageW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret != 0
}

func defWindowProc(hwnd hWnd, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

func beginPaint(hwnd hWnd, ps *paintStruct) hDc {
	ret, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return hDc(ret)
}

func endPaint(hwnd hWnd, ps *paintStruct) bool {
	ret, _, _ := procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return ret != 0
}

func getClientRect(hwnd hWnd, rect *rect) bool {
	ret, _, _ := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return ret != 0
}

func invalidateRect(hwnd hWnd, rect *rect, erase bool) bool {
	var eraseVal uintptr
	if erase {
		eraseVal = 1
	}
	ret, _, _ := procInvalidateRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)), eraseVal)
	return ret != 0
}

func fillRect(hdc hDc, rect *rect, brush hBrush) int32 {
	ret, _, _ := procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush))
	return int32(ret)
}

func drawText(hdc hDc, text string, rect *rect, format uint32) int32 {
	textPtr := StringToUTF16Ptr(text)
	ret, _, _ := procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(textPtr)), uintptr(len(text)), uintptr(unsafe.Pointer(rect)), uintptr(format))
	return int32(ret)
}

func getWindowLongPtr(hwnd hWnd, index int32) uintptr {
	ret, _, _ := procGetWindowLongPtr.Call(uintptr(hwnd), uintptr(index))
	return ret
}

func setBkMode(hdc hDc, mode int32) int32 {
	ret, _, _ := procSetBkMode.Call(uintptr(hdc), uintptr(mode))
	return int32(ret)
}

func setTextColor(hdc hDc, color colorRef) colorRef {
	ret, _, _ := procSetTextColor.Call(uintptr(hdc), uintptr(color))
	return colorRef(ret)
}

func createSolidBrush(color colorRef) hBrush {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return hBrush(ret)
}

func deleteObject(obj handle) bool {
	if obj == 0 {
		return false
	}
	ret, _, _ := procDeleteObject.Call(uintptr(obj))
	return ret != 0
}

func selectObject(hdc hDc, obj handle) handle {
	ret, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return handle(ret)
}

func createFont(height, width, escapement, orientation, weight int32, italic, underline, strikeOut, charSet, outputPrecision, clipPrecision, quality, pitchAndFamily uint32, faceName string) hFont {
	ret, _, _ := procCreateFontW.Call(
		uintptr(height), uintptr(width), uintptr(escapement), uintptr(orientation), uintptr(weight),
		uintptr(italic), uintptr(underline), uintptr(strikeOut), uintptr(charSet),
		uintptr(outputPrecision), uintptr(clipPrecision), uintptr(quality), uintptr(pitchAndFamily),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(faceName))),
	)
	return hFont(ret)
}

func createPen(style, width int32, color colorRef) hPen {
	ret, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return hPen(ret)
}

func roundRect(hdc hDc, rect *rect, width, height int32) bool {
	ret, _, _ := procRoundRect.Call(uintptr(hdc), uintptr(rect.Left), uintptr(rect.Top), uintptr(rect.Right), uintptr(rect.Bottom), uintptr(width), uintptr(height))
	return ret != 0
}

func moveToEx(hdc hDc, x, y int32, point *point) bool {
	ret, _, _ := procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(point)))
	return ret != 0
}

func lineTo(hdc hDc, x, y int32) bool {
	ret, _, _ := procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(y))
	return ret != 0
}

// rgb creates a colorref from red, green, blue values
func rgb(r, g, b byte) colorRef {
	return colorRef(uint32(r) | uint32(g)<<8 | uint32(b)<<16)
}

func sendMessage(hwnd hWnd, msg uint32, wParam, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
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

func init() {
	initCommonControls(
		ICC_WIN95_CLASSES |
			ICC_STANDARD_CLASSES |
			ICC_BAR_CLASSES |
			ICC_TAB_CLASSES |
			ICC_UPDOWN_CLASS |
			ICC_PROGRESS_CLASS |
			ICC_HOTKEY_CLASS |
			ICC_ANIMATE_CLASS |
			ICC_DATE_CLASSES |
			ICC_USEREX_CLASSES |
			ICC_COOL_CLASSES |
			ICC_INTERNET_CLASSES |
			ICC_PAGESCROLLER_CLASS |
			ICC_NATIVEFNTCTL_CLASS |
			ICC_LINK_CLASS |
			ICC_LISTVIEW_CLASSES |
			ICC_TREEVIEW_CLASSES,
	)
}

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
func setWindowTheme(hwnd hWnd, appName, idList string) error {
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
