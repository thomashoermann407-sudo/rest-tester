package win32

import "syscall"

type handle uintptr
type hWnd handle
type hInstance handle
type hIcon handle
type hCursor handle
type hBrush handle
type hMenu handle
type hDc handle
type hFont handle
type hPen handle
type colorRef uint32

const (
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	WS_OVERLAPPED       = 0x00000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER           = 0x00800000
	WS_CLIPCHILDREN     = 0x02000000

	WM_CREATE        = 0x0001
	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_ACTIVATE      = 0x0006
	WM_KILLFOCUS     = 0x0008
	WM_PAINT         = 0x000F
	WM_CLOSE         = 0x0010
	WM_ERASEBKGND    = 0x0014
	WM_SETFONT       = 0x0030
	WM_SETCURSOR     = 0x0020
	WM_NCCALCSIZE    = 0x0083
	WM_NCHITTEST     = 0x0084
	WM_NCMOUSEMOVE   = 0x00A0
	WM_NCLBUTTONDOWN = 0x00A1
	WM_NOTIFY        = 0x004E
	WM_COMMAND       = 0x0111
	WM_CONTEXTMENU   = 0x007B
	WM_KEYDOWN       = 0x0100
	WM_MOUSEMOVE     = 0x0200
	WM_LBUTTONDOWN   = 0x0201
	WM_LBUTTONUP     = 0x0202
	WM_RBUTTONDOWN   = 0x0204
	WM_RBUTTONUP     = 0x0205
	WM_MOUSELEAVE    = 0x02A3
	WM_USER          = 0x0400
	WM_APP           = 0x8000

	// Custom application messages
	WM_UI_CALLBACK = WM_APP + 1

	// Non-client hit test values
	HTCLIENT      = 1
	HTCAPTION     = 2
	HTSYSMENU     = 3
	HTMINBUTTON   = 8
	HTMAXBUTTON   = 9
	HTCLOSE       = 20
	HTTOPLEFT     = 13
	HTTOP         = 12
	HTTOPRIGHT    = 14
	HTLEFT        = 10
	HTRIGHT       = 11
	HTBOTTOMLEFT  = 16
	HTBOTTOM      = 15
	HTBOTTOMRIGHT = 17
	HTNOWHERE     = 0

	CW_USEDEFAULT = -2147483648

	SW_SHOW = 5
	SW_HIDE = 0

	IDC_ARROW  = 32512
	IDC_HAND   = 32649
	IDC_SIZENS = 32645 // Vertical resize cursor (north-south)

	// Resource IDs
	IDI_APPLICATION = 32512 // Default application icon
	RT_ICON         = 3     // Icon resource type

	COLOR_WINDOW  = 5
	COLOR_BTNFACE = 15

	BS_PUSHBUTTON    = 0x00000000
	BS_DEFPUSHBUTTON = 0x00000001
	BS_AUTOCHECKBOX  = 0x00000003
	BS_GROUPBOX      = 0x00000007

	ES_LEFT        = 0x0000
	ES_AUTOHSCROLL = 0x0080

	MB_OK    = 0x00000000
	MB_YESNO = 0x00000004
	ID_YES   = 6
	ID_NO    = 7

	// Drawing constants
	TRANSPARENT = 1
	OPAQUE      = 2

	// DrawText format flags
	DT_LEFT         = 0x00000000
	DT_CENTER       = 0x00000001
	DT_RIGHT        = 0x00000002
	DT_VCENTER      = 0x00000004
	DT_SINGLELINE   = 0x00000020
	DT_END_ELLIPSIS = 0x00008000
	DT_NOPREFIX     = 0x00000800

	// Pen styles
	PS_SOLID = 0

	// Stock objects
	NULL_BRUSH  = 5
	DC_BRUSH    = 18
	DC_PEN      = 19
	SYSTEM_FONT = 13

	// Font weights
	FW_NORMAL = 400
	FW_BOLD   = 700

	// Character sets
	DEFAULT_CHARSET = 1

	// Output precision
	OUT_DEFAULT_PRECIS = 0

	// Clip precision
	CLIP_DEFAULT_PRECIS = 0

	// Quality
	DEFAULT_QUALITY   = 0
	CLEARTYPE_QUALITY = 5

	// Pitch and family
	DEFAULT_PITCH = 0
	FF_DONTCARE   = 0

	// Track mouse event flags
	TME_LEAVE = 0x00000002

	// ComboBox styles
	CBS_DROPDOWN     = 0x0002
	CBS_DROPDOWNLIST = 0x0003
	CBS_HASSTRINGS   = 0x0200

	// ComboBox messages
	CB_ADDSTRING    = 0x0143
	CB_GETCURSEL    = 0x0147
	CB_SETCURSEL    = 0x014E
	CB_GETLBTEXT    = 0x0148
	CB_GETLBTEXTLEN = 0x0149
	CB_RESETCONTENT = 0x014B

	// Edit control styles
	ES_MULTILINE   = 0x0004
	ES_AUTOVSCROLL = 0x0040
	ES_READONLY    = 0x0800
	ES_WANTRETURN  = 0x1000

	// Window styles extended
	WS_VSCROLL       = 0x00200000
	WS_HSCROLL       = 0x00100000
	WS_EX_CLIENTEDGE = 0x00000200

	// Static control styles
	SS_LEFT = 0x0000

	// Window messages for text
	WM_SETTEXT       = 0x000C
	WM_GETTEXT       = 0x000D
	WM_GETTEXTLENGTH = 0x000E

	// ListBox styles
	LBS_NOTIFY           = 0x0001
	LBS_HASSTRINGS       = 0x0040
	LBS_NOINTEGRALHEIGHT = 0x0100

	// ListBox messages
	LB_ADDSTRING    = 0x0180
	LB_DELETESTRING = 0x0182
	LB_GETCURSEL    = 0x0188
	LB_SETCURSEL    = 0x0186
	LB_GETTEXT      = 0x0189
	LB_GETTEXTLEN   = 0x018A
	LB_GETCOUNT     = 0x018B
	LB_RESETCONTENT = 0x0184

	// ListBox notifications
	LBN_SELCHANGE = 1
	LBN_DBLCLK    = 2

	// Button notifications
	BN_CLICKED = 0

	// Get/SetWindowLongPtr indices
	GWLP_ID      = -12
	GWLP_WNDPROC = -4

	// Checkbox helper functions
	BM_GETCHECK = 0x00F0
	BM_SETCHECK = 0x00F1
	BST_CHECKED = 1

	// TreeView constants
	TVS_HASBUTTONS    = 0x0001
	TVS_HASLINES      = 0x0002
	TVS_LINESATROOT   = 0x0004
	TVS_SHOWSELALWAYS = 0x0020
	TVS_FULLROWSELECT = 0x1000
	TVM_INSERTITEMW   = 0x1132
	TVM_DELETEITEM    = 0x1101
	TVM_GETNEXTITEM   = 0x110A
	TVM_EXPAND        = 0x1102
	TVM_SELECTITEM    = 0x110B
	TVM_GETITEMW      = 0x113E
	TVM_SETITEMW      = 0x113F
	TVM_HITTEST       = 0x1111
	TVGN_ROOT         = 0x0000
	TVGN_NEXT         = 0x0001
	TVGN_CHILD        = 0x0004
	TVGN_PARENT       = 0x0003
	TVGN_CARET        = 0x0009
	TVE_COLLAPSE      = 0x0001
	TVE_EXPAND        = 0x0002
	TVIF_TEXT         = 0x0001
	TVIF_STATE        = 0x0008
	TVIF_PARAM        = 0x0004
	TVIF_HANDLE       = 0x0010
	TVIS_EXPANDED     = 0x0020
	TVI_ROOT          = ^uintptr(0xFFFF) // -0x10000 sign-extended
	TVI_FIRST         = ^uintptr(0xFFFE) // -0x0FFFF sign-extended
	TVI_LAST          = ^uintptr(0xFFFD) // -0x0FFFE sign-extended
	TVI_SORT          = ^uintptr(0xFFFC) // -0x0FFFD sign-extended
	NM_DBLCLK         = ^uint32(3) + 1   // -3 as unsigned
	NM_RCLICK         = ^uint32(5) + 1   // -5 as unsigned
	TVN_SELCHANGEDW   = ^uint32(401) + 1 // -401 as unsigned

	// TabControl constants
	TCS_TABS           = 0x0000
	TCS_MULTILINE      = 0x0200
	TCM_INSERTITEMW    = 0x133E
	TCM_DELETEITEM     = 0x1308
	TCM_DELETEALLITEMS = 0x1309
	TCM_GETCURSEL      = 0x130B
	TCM_SETCURSEL      = 0x130C
	TCM_GETITEMCOUNT   = 0x1304
	TCIF_TEXT          = 0x0001
	TCIF_PARAM         = 0x0008
	TCN_SELCHANGE      = ^uint32(550) + 1 // -550 as unsigned

	// ListView constants
	LVS_REPORT                   = 0x0001
	LVS_SINGLESEL                = 0x0004
	LVS_SHOWSELALWAYS            = 0x0008
	LVS_NOSORTHEADER             = 0x8000
	LVM_INSERTCOLUMNW            = 0x1061
	LVM_INSERTITEMW              = 0x104D
	LVM_SETITEMW                 = 0x104C
	LVM_GETITEMW                 = 0x104B
	LVM_DELETEITEM               = 0x1008
	LVM_DELETEALLITEMS           = 0x1009
	LVM_GETNEXTITEM              = 0x100C
	LVM_GETITEMCOUNT             = 0x1004
	LVM_ENSUREVISIBLE            = 0x1013
	LVM_SETITEMTEXTW             = 0x1074
	LVM_SETEXTENDEDLISTVIEWSTYLE = 0x1036
	LVM_SUBITEMHITTEST           = 0x1039
	LVM_GETITEMRECT              = 0x100E
	LVM_GETSUBITEMRECT           = 0x1038
	LVM_SETITEMSTATE             = 0x102B
	LVS_EX_FULLROWSELECT         = 0x00000020
	LVS_EX_GRIDLINES             = 0x00000001
	LVIF_TEXT                    = 0x0001
	LVIF_PARAM                   = 0x0004
	LVIF_STATE                   = 0x0008
	LVCF_TEXT                    = 0x0004
	LVCF_WIDTH                   = 0x0002
	LVNI_SELECTED                = 0x0002
	LVIS_SELECTED                = 0x0002
	LVIS_FOCUSED                 = 0x0001
	LVN_ITEMCHANGED              = ^uint32(100) + 1 // -100 as unsigned
	LVIR_BOUNDS                  = 0
	LVIR_LABEL                   = 2
	LVM_GETITEMTEXTW             = 0x1073

	// Edit control messages
	EM_SETSEL      = 0x00B1
	EM_SETREADONLY = 0x00CF

	// Virtual key codes
	VK_RETURN = 0x0D
	VK_ESCAPE = 0x1B
)

// ComCtl32 Common Control Class Names
const (
	WC_BUTTON          = "Button"   // Standard button (User32, but works with ComCtl32)
	WC_EDIT            = "Edit"     // Standard edit (User32, but works with ComCtl32)
	WC_STATIC          = "Static"   // Standard static (User32, but works with ComCtl32)
	WC_LISTBOX         = "ListBox"  // Standard listbox (User32, but works with ComCtl32)
	WC_COMBOBOX        = "ComboBox" // Standard combobox (User32, but works with ComCtl32)
	TOOLBARCLASSNAME   = "ToolbarWindow32"
	STATUSCLASSNAME    = "msctls_statusbar32"
	TRACKBAR_CLASS     = "msctls_trackbar32"
	UPDOWN_CLASS       = "msctls_updown32"
	PROGRESS_CLASS     = "msctls_progress32"
	HOTKEY_CLASS       = "msctls_hotkey32"
	WC_LISTVIEW        = "SysListView32"
	WC_TREEVIEW        = "SysTreeView32"
	WC_TABCONTROL      = "SysTabControl32"
	ANIMATE_CLASS      = "SysAnimate32"
	WC_HEADER          = "SysHeader32"
	MONTHCAL_CLASS     = "SysMonthCal32"
	DATETIMEPICK_CLASS = "SysDateTimePick32"
	WC_IPADDRESS       = "SysIPAddress32"
	WC_PAGESCROLLER    = "SysPager"
	WC_NATIVEFONTCTL   = "NativeFontCtl"
	WC_LINK            = "SysLink"
)

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   hInstance
	Icon       hIcon
	Cursor     hCursor
	Background hBrush
	MenuName   *uint16
	ClassName  *uint16
	IconSm     hIcon
}

type point struct {
	X, Y int32
}

type msg struct {
	Hwnd    hWnd
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type rect struct {
	Left, Top, Right, Bottom int32
}

func (r *rect) inside(x, y int32) bool {
	return x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom
}

type paintStruct struct {
	Hdc         hDc
	Erase       int32
	RcPaint     rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

// openFilename structure for file dialogs
type openFilename struct {
	StructSize      uint32
	Owner           hWnd
	Instance        hInstance
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustData        uintptr
	FnHook          uintptr
	TemplateName    *uint16
	PvReserved      uintptr
	DwReserved      uint32
	FlagsEx         uint32
}

// File dialog flags
const (
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_PATHMUSTEXIST   = 0x00000800
	OFN_OVERWRITEPROMPT = 0x00000002
	OFN_NOCHANGEDIR     = 0x00000008
	OFN_EXPLORER        = 0x00080000
)

// Helper to convert string to *uint16
func StringToUTF16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}
