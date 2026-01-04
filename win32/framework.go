package win32

import (
	"syscall"
	"unsafe"
)

type TabDrawer interface {
	Paint(hdc hDc, width int32)
	HandleMouseMove(x, y, width int32)
	HandleClick(x, y, width int32)
	Invalidate()
	GetHeight() int32
	Destroy()
}

type Window struct {
	hwnd        hWnd
	handlers    map[hWnd]func(msg uint32, wParam, lParam uintptr) (uintptr, bool)
	controls    map[int]ClickController // Map control IDs to controllers
	OnCommand   func(id int, notifyCode int)
	OnResize    func(width, height int32)
	OnMouseMove func(x, y int32) bool // Returns true if handled
	OnMouseDown func(x, y int32) bool // Returns true if handled
	OnMouseUp   func(x, y int32) bool // Returns true if handled
	OnDestroy   func()                // Called when window is being destroyed
	TabManager  TabDrawer
	width       int32
	height      int32
	font        hFont
	monoFont    hFont

	// UI callback queue for thread-safe UI updates
	uiCallbacks    map[uintptr]func()
	nextCallbackID uintptr
}

func NewWindow(title string, width, height int32) *Window {
	hInstance := getModuleHandle(nil)
	className := StringToUTF16Ptr("MyWindowClass")

	w := &Window{
		handlers:       make(map[hWnd]func(msg uint32, wParam, lParam uintptr) (uintptr, bool)),
		controls:       make(map[int]ClickController),
		width:          width,
		height:         height,
		uiCallbacks:    make(map[uintptr]func()),
		nextCallbackID: 1,
	}

	// Load the application icon from resources
	appIcon := loadIcon(hInstance, uintptr(1)) // Resource ID 1 matches app.rc

	wcx := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		Style:      CS_HREDRAW | CS_VREDRAW,
		WndProc:    syscall.NewCallback(w.wndProc),
		Instance:   hInstance,
		Background: hBrush(COLOR_BTNFACE + 1),
		ClassName:  className,
		Cursor:     loadCursor(0, uintptr(IDC_ARROW)),
		Icon:       appIcon, // Large icon (32x32)
		IconSm:     appIcon, // Small icon (16x16) - using same icon, Windows will resize
	}
	registerClassEx(&wcx)

	w.hwnd = createWindowEx(
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
	w.handlers[w.hwnd] = func(msg uint32, wParam, lParam uintptr) (uintptr, bool) {
		switch msg {
		case WM_UI_CALLBACK:
			// Handle UI callback from background thread
			callbackID := wParam
			if callback, ok := w.uiCallbacks[callbackID]; ok {
				callback()
				delete(w.uiCallbacks, callbackID)
			}
			return 0, true

		case WM_COMMAND:
			if w.OnCommand != nil {
				id := int(wParam & 0xFFFF)
				notifyCode := int((wParam >> 16) & 0xFFFF)
				w.OnCommand(id, notifyCode)
				return 0, true
			}

		case WM_NOTIFY:
			// Handle notifications from controls
			// Note: This is a safe conversion. The lParam for WM_NOTIFY messages
			// is guaranteed by Windows to be a valid pointer to an NMHDR structure
			// for the duration of the message handling. This is documented Windows behavior.
			// The unsafeptr analyzer warning is a false positive in this context.
			nmhdr := (*NMHDR)(unsafe.Pointer(lParam))
			if ctrl, ok := w.controls[int(nmhdr.IdFrom)]; ok {
				//TODO: Harmonize: HandleCommand in panels.go needs to be updated to support more controls
				switch nmhdr.Code {
				case NM_DBLCLK:
					// Handle double-click for ListView in-place editing
					if listView, ok := ctrl.(*ListViewControl); ok {
						// Get cursor position
						var pt point
						procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
						// Convert to client coordinates
						procScreenToClient.Call(uintptr(listView.Hwnd), uintptr(unsafe.Pointer(&pt)))
						// Hit test to find the cell
						row, col := listView.HitTestEx(pt.X, pt.Y)
						if row >= 0 && col >= 0 {
							listView.StartEdit(row, col)
							return 0, true
						}
					}
					ctrl.OnDoubleClick()
					return 0, true
				case NM_RCLICK:
					// Handle right-click for TreeView
					if treeView, ok := ctrl.(*TreeViewControl); ok {
						// Get cursor position
						var pt point
						procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
						// Convert to client coordinates
						procScreenToClient.Call(uintptr(treeView.Hwnd), uintptr(unsafe.Pointer(&pt)))
						// Hit test to find the item
						itemHandle := treeView.HitTest(pt.X, pt.Y)
						treeView.OnRightClick(itemHandle)
						return 0, true
					}
				case TCN_SELCHANGE:
					// Handle tab selection change for TabControl
					if tabControl, ok := ctrl.(*TabControlControl); ok {
						tabControl.OnSelChange()
						return 0, true
					}
				case LVN_ITEMCHANGED:
					// Handle selection change for ListView
					if listView, ok := ctrl.(*ListViewControl); ok {
						listView.OnSelChange()
						return 0, true
					}
				}
			}
			return 0, false

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
				var ps paintStruct
				hdc := beginPaint(w.hwnd, &ps)
				w.TabManager.Paint(hdc, w.width)
				endPaint(w.hwnd, &ps)
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
				if y < w.TabManager.GetHeight() {
					w.TabManager.HandleClick(x, y, w.width)
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

		case WM_MOUSELEAVE:
			if w.TabManager != nil {
				w.TabManager.Invalidate()
			}
			return 0, false

		case WM_DESTROY:
			// Call the destroy callback to allow cleanup
			if w.OnDestroy != nil {
				w.OnDestroy()
			}
			w.TabManager.Destroy()
			w.Destroy()
			// Let the default handler process WM_DESTROY
			return 0, false
		}

		return 0, false
	}

	return w
}

func (w *Window) MessageBox(caption, text string) int32 {
	ret, _, _ := procMessageBoxW.Call(
		uintptr(w.hwnd),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(caption))),
		uintptr(MB_OK),
	)
	return int32(ret)
}

func (w *Window) MessageBoxYesNo(caption, text string) int32 {
	ret, _, _ := procMessageBoxW.Call(
		uintptr(w.hwnd),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(caption))),
		uintptr(MB_YESNO),
	)
	return int32(ret)
}

// OpenFileDialog shows a file open dialog and returns the selected file path
func (w *Window) OpenFileDialog(title, filter, defaultExt string) (string, bool) {
	var ofn openFilename
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
	ofn.Owner = w.hwnd
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
func (w *Window) SaveFileDialog(title, filter, defaultExt, defaultName string) (string, bool) {
	var ofn openFilename
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
	ofn.Owner = w.hwnd
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

func Run() {
	var msg msg
	for getMessage(&msg, 0, 0, 0) > 0 {
		translateMessage(&msg)
		dispatchMessage(&msg)
	}
}

// Destroy cleans up all OS resources (fonts, window handle, etc.)
// This should be called before the application exits to prevent resource leaks
func (w *Window) Destroy() {
	// Delete GDI font objects
	if w.font != 0 {
		deleteObject(handle(w.font))
		w.font = 0
	}

	if w.monoFont != 0 {
		deleteObject(handle(w.monoFont))
		w.monoFont = 0
	}

	// Remove the window's message handler
	if w.hwnd != 0 {
		delete(w.handlers, w.hwnd)
	}

	// Destroy the window handle (this also destroys all child controls)
	if w.hwnd != 0 {
		destroyWindow(w.hwnd)
		w.hwnd = 0
	}
}

func (w *Window) GetWidth() int32 {
	return w.width
}
func (w *Window) GetHeight() int32 {
	return w.height
}

func (w *Window) wndProc(hwnd hWnd, msg uintptr, wParam, lParam uintptr) uintptr {
	if handler, ok := w.handlers[hwnd]; ok {
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

// applyFont applies the modern font to a control and enables visual styles
func (w *Window) applyFont(hwnd hWnd) {
	if w.font != 0 {
		sendMessage(hwnd, WM_SETFONT, uintptr(w.font), 1)
	}
	// Enable modern visual styles for the control
	setWindowTheme(hwnd, "", "")
}

// PostUICallback posts a callback function to be executed on the UI thread
// This is thread-safe and should be used when updating UI from background goroutines
func (w *Window) PostUICallback(callback func()) {
	callbackID := w.nextCallbackID
	w.nextCallbackID++
	w.uiCallbacks[callbackID] = callback
	postMessage(w.hwnd, WM_UI_CALLBACK, callbackID, 0)
}
