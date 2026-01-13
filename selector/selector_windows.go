package selector

import (
	"image"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// User32 functions
	procRegisterClassEx            = user32.NewProc("RegisterClassExW")
	procCreateWindowEx             = user32.NewProc("CreateWindowExW")
	procDefWindowProc              = user32.NewProc("DefWindowProcW")
	procGetMessage                 = user32.NewProc("GetMessageW")
	procTranslateMessage           = user32.NewProc("TranslateMessage")
	procDispatchMessage            = user32.NewProc("DispatchMessageW")
	procPostQuitMessage            = user32.NewProc("PostQuitMessage")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procUpdateWindow               = user32.NewProc("UpdateWindow")
	procDestroyWindow              = user32.NewProc("DestroyWindow")
	procGetDC                      = user32.NewProc("GetDC")
	procReleaseDC                  = user32.NewProc("ReleaseDC")
	procInvalidateRect             = user32.NewProc("InvalidateRect")
	procBeginPaint                 = user32.NewProc("BeginPaint")
	procEndPaint                   = user32.NewProc("EndPaint")
	procFillRect                   = user32.NewProc("FillRect")
	procSetCursor                  = user32.NewProc("SetCursor")
	procLoadCursor                 = user32.NewProc("LoadCursorW")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procScreenToClient             = user32.NewProc("ScreenToClient")
	procSetCapture                 = user32.NewProc("SetCapture")
	procReleaseCapture             = user32.NewProc("ReleaseCapture")

	// GDI32 functions
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procRectangle        = gdi32.NewProc("Rectangle")
	procSetROP2          = gdi32.NewProc("SetROP2")
	procGetStockObject   = gdi32.NewProc("GetStockObject")

	// Kernel32 functions
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

// Windows constants
const (
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_POPUP          = 0x80000000
	WS_VISIBLE        = 0x10000000

	SW_SHOW     = 5
	SW_MAXIMIZE = 3

	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_KEYDOWN     = 0x0100
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_MOUSEMOVE   = 0x0200
	WM_SETCURSOR   = 0x0020
	WM_ERASEBKGND  = 0x0014

	VK_ESCAPE = 0x1B

	SM_XVIRTUALSCREEN  = 76
	SM_YVIRTUALSCREEN  = 77
	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79

	LWA_ALPHA    = 0x00000002
	LWA_COLORKEY = 0x00000001

	IDC_CROSS = 32515

	PS_SOLID = 0
	PS_DASH  = 1

	R2_NOT     = 6
	R2_XORPEN  = 7
	NULL_BRUSH = 5

	COLOR_HIGHLIGHT = 13
)

// WNDCLASSEX structure
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// MSG structure
type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// POINT structure
type POINT struct {
	X, Y int32
}

// RECT structure
type RECT struct {
	Left, Top, Right, Bottom int32
}

// PAINTSTRUCT structure
type PAINTSTRUCT struct {
	HDC         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

// Selection state
type selectionState struct {
	hwnd      uintptr
	selecting bool
	startX    int32
	startY    int32
	endX      int32
	endY      int32
	cancelled bool
	completed bool
	virtualX  int32
	virtualY  int32
}

var state selectionState

// SelectRegion opens a fullscreen overlay and lets user select a region with mouse
func SelectRegion() (image.Rectangle, bool) {
	state = selectionState{}

	// Get virtual screen dimensions (all monitors)
	state.virtualX = int32(getSystemMetrics(SM_XVIRTUALSCREEN))
	state.virtualY = int32(getSystemMetrics(SM_YVIRTUALSCREEN))
	virtualW := getSystemMetrics(SM_CXVIRTUALSCREEN)
	virtualH := getSystemMetrics(SM_CYVIRTUALSCREEN)

	// Get module handle
	hInstance, _, _ := procGetModuleHandle.Call(0)

	// Register window class
	className := syscall.StringToUTF16Ptr("ScreenshotSelector")

	wndClass := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         0,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       loadCursor(IDC_CROSS),
		HbrBackground: 0,
		LpszClassName: className,
	}

	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wndClass)))

	// Create layered window covering all monitors
	hwnd, _, _ := procCreateWindowEx.Call(
		WS_EX_LAYERED|WS_EX_TOPMOST|WS_EX_TOOLWINDOW,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Select Region"))),
		WS_POPUP|WS_VISIBLE,
		uintptr(state.virtualX),
		uintptr(state.virtualY),
		uintptr(virtualW),
		uintptr(virtualH),
		0, 0, hInstance, 0,
	)

	if hwnd == 0 {
		return image.Rectangle{}, false
	}

	state.hwnd = hwnd

	// Set window transparency (semi-transparent dark overlay)
	procSetLayeredWindowAttributes.Call(hwnd, 0, 80, LWA_ALPHA)

	// Show window
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)

	// Message loop
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || int32(ret) == -1 {
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))

		if state.completed || state.cancelled {
			break
		}
	}

	// Destroy window
	procDestroyWindow.Call(hwnd)

	if state.cancelled {
		return image.Rectangle{}, false
	}

	// Calculate selected rectangle (handle any drag direction)
	x1, y1 := state.startX, state.startY
	x2, y2 := state.endX, state.endY

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Convert to screen coordinates
	rect := image.Rect(
		int(x1+state.virtualX),
		int(y1+state.virtualY),
		int(x2+state.virtualX),
		int(y2+state.virtualY),
	)

	return rect, true
}

// wndProc handles window messages
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		// Fill with semi-transparent dark color
		brush, _, _ := procCreateSolidBrush.Call(0x4A4A4A) // Dark gray
		rect := &RECT{
			Left:   ps.RcPaint.Left,
			Top:    ps.RcPaint.Top,
			Right:  ps.RcPaint.Right,
			Bottom: ps.RcPaint.Bottom,
		}
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(rect)), brush)
		procDeleteObject.Call(brush)

		// Draw selection rectangle if selecting
		if state.selecting || state.completed {
			drawSelectionRect(hdc)
		}

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_LBUTTONDOWN:
		state.selecting = true
		state.startX = int32(lParam & 0xFFFF)
		state.startY = int32((lParam >> 16) & 0xFFFF)
		state.endX = state.startX
		state.endY = state.startY
		procSetCapture.Call(hwnd)
		return 0

	case WM_MOUSEMOVE:
		if state.selecting {
			state.endX = int32(lParam & 0xFFFF)
			state.endY = int32((lParam >> 16) & 0xFFFF)
			procInvalidateRect.Call(hwnd, 0, 1)
		}
		return 0

	case WM_LBUTTONUP:
		if state.selecting {
			state.endX = int32(lParam & 0xFFFF)
			state.endY = int32((lParam >> 16) & 0xFFFF)
			state.selecting = false
			state.completed = true
			procReleaseCapture.Call()
			procPostQuitMessage.Call(0)
		}
		return 0

	case WM_KEYDOWN:
		if wParam == VK_ESCAPE {
			state.cancelled = true
			procPostQuitMessage.Call(0)
		}
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// drawSelectionRect draws the selection rectangle
func drawSelectionRect(hdc uintptr) {
	x1, y1 := state.startX, state.startY
	x2, y2 := state.endX, state.endY

	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}

	// Create white pen for border
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, 0xFFFFFF) // White
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	// Use hollow brush
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)

	// Draw rectangle
	procRectangle.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))

	// Create highlight pen for inner border
	pen2, _, _ := procCreatePen.Call(PS_SOLID, 1, 0x00FF00) // Green
	procSelectObject.Call(hdc, pen2)
	procRectangle.Call(hdc, uintptr(x1+2), uintptr(y1+2), uintptr(x2-2), uintptr(y2-2))

	// Restore old objects
	procSelectObject.Call(hdc, oldPen)
	procSelectObject.Call(hdc, oldBrush)
	procDeleteObject.Call(pen)
	procDeleteObject.Call(pen2)
}

func getSystemMetrics(index int) int {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(ret)
}

func loadCursor(cursorID int) uintptr {
	cursor, _, _ := procLoadCursor.Call(0, uintptr(cursorID))
	return cursor
}
