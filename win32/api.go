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
