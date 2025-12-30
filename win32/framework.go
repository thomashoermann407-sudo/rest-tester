package win32

import (
	"syscall"
	"unsafe"
)

var (
	handlers   = make(map[HWND]func(msg uint32, wParam, lParam uintptr) (uintptr, bool))
	registered = false
	comCtlInit = false
)

func WndProc(hwnd HWND, msg uintptr, wParam, lParam uintptr) uintptr {
	if handler, ok := handlers[hwnd]; ok {
		ret, handled := handler(uint32(msg), wParam, lParam)
		if handled {
			return ret
		}
	}

	switch uint32(msg) {
	case WM_DESTROY:
		postQuitMessage(0)
		return 0
	case WM_CLOSE:
		postQuitMessage(0)
		return 0
	default:
		return defWindowProc(hwnd, uint32(msg), wParam, lParam)
	}
}

var DefaultWndProc = syscall.NewCallback(WndProc)

type Window struct {
	Hwnd        HWND
	OnCommand   func(id int, notifyCode int)
	OnResize    func(width, height int32)
	OnMouseMove func(x, y int32) bool // Returns true if handled
	OnMouseDown func(x, y int32) bool // Returns true if handled
	OnMouseUp   func(x, y int32) bool // Returns true if handled
	OnSetCursor func(x, y int32) bool // Returns true if cursor was set
	TabManager  *TabManager
	width       int32
	height      int32
	font        HFONT
	monoFont    HFONT
	controls    []HWND
}

func NewWindow(title string, width, height int32) *Window {
	// Initialize ALL common controls once with modern visual styles
	if !comCtlInit {
		// Initialize all control classes for full ComCtl32 support
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
		comCtlInit = true
	}

	hInstance := getModuleHandle(nil)
	className := StringToUTF16Ptr("MyWindowClass")

	if !registered {
		// Load the application icon from resources
		appIcon := loadIcon(hInstance, uintptr(1)) // Resource ID 1 matches app.rc

		wcx := WNDCLASSEX{
			Size:       uint32(unsafe.Sizeof(WNDCLASSEX{})),
			Style:      CS_HREDRAW | CS_VREDRAW,
			WndProc:    DefaultWndProc,
			Instance:   hInstance,
			Background: HBRUSH(COLOR_BTNFACE + 1),
			ClassName:  className,
			Cursor:     loadCursor(0, uintptr(IDC_ARROW)),
			Icon:       appIcon, // Large icon (32x32)
			IconSm:     appIcon, // Small icon (16x16) - using same icon, Windows will resize
		}
		registerClassEx(&wcx)
		registered = true
	}

	hwnd := createWindowEx(
		0,
		className,
		StringToUTF16Ptr(title),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE|WS_CLIPCHILDREN,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		width,
		height,
		0,
		0,
		hInstance,
		nil,
	)

	w := &Window{
		Hwnd:     hwnd,
		width:    width,
		height:   height,
		controls: make([]HWND, 0),
	}

	// Create modern Segoe UI font
	w.font = createFont(
		-14, 0, 0, 0, FW_NORMAL,
		0, 0, 0, DEFAULT_CHARSET,
		OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS,
		CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE,
		"Segoe UI",
	)

	// Create monospace font for code areas
	w.monoFont = createFont(
		-13, 0, 0, 0, FW_NORMAL,
		0, 0, 0, DEFAULT_CHARSET,
		OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS,
		CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE,
		"Cascadia Code",
	)

	// Register handler for this window
	handlers[hwnd] = func(msg uint32, wParam, lParam uintptr) (uintptr, bool) {
		switch msg {
		case WM_COMMAND:
			if w.OnCommand != nil {
				id := int(wParam & 0xFFFF)
				notifyCode := int((wParam >> 16) & 0xFFFF)
				w.OnCommand(id, notifyCode)
				return 0, true
			}

		case WM_SIZE:
			w.width = int32(lParam & 0xFFFF)
			w.height = int32((lParam >> 16) & 0xFFFF)
			if w.OnResize != nil {
				w.OnResize(w.width, w.height)
			}
			// Also invalidate tab bar to repaint on resize
			if w.TabManager != nil {
				w.TabManager.Invalidate()
			}
			return 0, false

		case WM_PAINT:
			if w.TabManager != nil {
				var ps PAINTSTRUCT
				hdc := beginPaint(hwnd, &ps)
				w.TabManager.Paint(hdc, w.width)
				endPaint(hwnd, &ps)
				return 0, true
			}

		case WM_MOUSEMOVE:
			x := int32(lParam & 0xFFFF)
			y := int32((lParam >> 16) & 0xFFFF)

			// Let tab manager handle it first (for tab bar area)
			if w.TabManager != nil {
				w.TabManager.HandleMouseMove(x, y, w.width)
			}

			// Then let application handle mouse move
			if w.OnMouseMove != nil && w.OnMouseMove(x, y) {
				return 0, true
			}
			return 0, false

		case WM_LBUTTONDOWN:
			x := int32(lParam & 0xFFFF)
			y := int32((lParam >> 16) & 0xFFFF)

			// Let tab manager handle it first (for tab bar area)
			if w.TabManager != nil {
				if w.TabManager.HandleClick(x, y, w.width) {
					return 0, true
				}
			}

			// Then let application handle mouse down
			if w.OnMouseDown != nil && w.OnMouseDown(x, y) {
				return 0, true
			}
			return 0, false

		case WM_LBUTTONUP:
			x := int32(lParam & 0xFFFF)
			y := int32((lParam >> 16) & 0xFFFF)

			// Let application handle mouse up
			if w.OnMouseUp != nil && w.OnMouseUp(x, y) {
				return 0, true
			}
			return 0, false

		case WM_SETCURSOR:
			// Let application set custom cursor
			if w.OnSetCursor != nil {
				x, y := getCursorPosClient(w.Hwnd)
				if w.OnSetCursor(x, y) {
					return 1, true // TRUE - cursor was set
				}
			}
			return 0, false

		case WM_MOUSELEAVE:
			if w.TabManager != nil {
				w.TabManager.HandleMouseLeave()
			}
			return 0, false
		}

		return 0, false
	}

	return w
}

// EnableTabs creates and attaches a TabManager to the window
func (w *Window) EnableTabs() *TabManager {
	w.TabManager = NewTabManager(w.Hwnd)
	return w.TabManager
}

func (w *Window) Run() {
	var msg MSG
	for getMessage(&msg, 0, 0, 0) > 0 {
		translateMessage(&msg)
		dispatchMessage(&msg)
	}
}

// applyFont applies the modern font to a control and enables visual styles
func (w *Window) applyFont(hwnd HWND) {
	if w.font != 0 {
		sendMessage(hwnd, WM_SETFONT, uintptr(w.font), 1)
	}
	// Enable modern visual styles for the control
	setWindowTheme(hwnd, "", "")
	w.controls = append(w.controls, hwnd)
}
