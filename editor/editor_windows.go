package editor

import (
	"image"
	"image/color"
	"syscall"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassEx  = user32.NewProc("RegisterClassExW")
	procCreateWindowEx   = user32.NewProc("CreateWindowExW")
	procDefWindowProc    = user32.NewProc("DefWindowProcW")
	procGetMessage       = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procInvalidateRect   = user32.NewProc("InvalidateRect")
	procBeginPaint       = user32.NewProc("BeginPaint")
	procEndPaint         = user32.NewProc("EndPaint")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procSetCapture       = user32.NewProc("SetCapture")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procLoadCursor       = user32.NewProc("LoadCursorW")
	procSetCursor        = user32.NewProc("SetCursor")
	procGetKeyState      = user32.NewProc("GetKeyState")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procSetDIBitsToDevice      = gdi32.NewProc("SetDIBitsToDevice")
	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procCreatePen              = gdi32.NewProc("CreatePen")
	procRectangle              = gdi32.NewProc("Rectangle")
	procRoundRect              = gdi32.NewProc("RoundRect")
	procEllipse                = gdi32.NewProc("Ellipse")
	procMoveToEx               = gdi32.NewProc("MoveToEx")
	procLineTo                 = gdi32.NewProc("LineTo")
	procGetStockObject         = gdi32.NewProc("GetStockObject")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procTextOut                = gdi32.NewProc("TextOutW")
	procCreateFont             = gdi32.NewProc("CreateFontW")
	procFillRect               = user32.NewProc("FillRect")

	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

const (
	WS_POPUP      = 0x80000000
	WS_VISIBLE    = 0x10000000
	WS_EX_LAYERED = 0x00080000
	WS_EX_TOPMOST = 0x00000008

	SW_SHOW = 5

	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_CLOSE       = 0x0010
	WM_KEYDOWN     = 0x0100
	WM_CHAR        = 0x0102
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_MOUSEMOVE   = 0x0200
	WM_MOUSEWHEEL  = 0x020A

	VK_ESCAPE = 0x1B
	VK_RETURN = 0x0D
	VK_BACK   = 0x08
	VK_DELETE = 0x2E

	SRCCOPY     = 0x00CC0020
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	IDC_ARROW = 32512
	IDC_CROSS = 32515
	IDC_IBEAM = 32513
	IDC_HAND  = 32649

	TRANSPARENT = 1
	NULL_BRUSH  = 5
	PS_SOLID    = 0

	FW_NORMAL = 400
	FW_BOLD   = 700

	BI_RGB       = 0
	DIB_RGB_COLORS = 0
)

// Tool types
const (
	TOOL_NONE = iota
	TOOL_TEXT
	TOOL_RECT
	TOOL_ARROW
	TOOL_HIGHLIGHT
	TOOL_FREEDRAW
)

// Custom colors
var (
	colorBg          = rgbToRef(30, 30, 35)
	colorToolbar     = rgbToRef(45, 45, 50)
	colorBtnNormal   = rgbToRef(60, 60, 70)
	colorBtnHover    = rgbToRef(80, 80, 95)
	colorBtnActive   = rgbToRef(100, 130, 255)
	colorBtnText     = rgbToRef(240, 240, 245)
	colorAccent      = rgbToRef(100, 130, 255)
	colorInputBg     = rgbToRef(40, 40, 48)
	colorInputBorder = rgbToRef(70, 70, 80)
	colorWhite       = rgbToRef(255, 255, 255)
	colorTextHint    = rgbToRef(120, 120, 130)
)

// Preset colors for palette
var paletteColors = []color.RGBA{
	{255, 0, 0, 255},     // Red
	{255, 100, 0, 255},   // Orange
	{255, 200, 0, 255},   // Yellow
	{0, 255, 0, 255},     // Green
	{0, 200, 255, 255},   // Cyan
	{0, 100, 255, 255},   // Blue
	{150, 0, 255, 255},   // Purple
	{255, 0, 150, 255},   // Pink
	{255, 255, 255, 255}, // White
	{0, 0, 0, 255},       // Black
}

// Button represents a custom button
type Button struct {
	X, Y, W, H int32
	Text       string
	Icon       string
	ToolID     int
	IsHovered  bool
	IsActive   bool
}

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type wndClassEx struct {
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

type paintStruct struct {
	HDC         uintptr
	FErase      int32
	RcPaint     struct{ Left, Top, Right, Bottom int32 }
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type editorState struct {
	hwnd       uintptr
	editor     *Editor
	winW, winH int32

	// Tools
	currentTool  int
	currentColor color.RGBA
	buttons      []Button
	colorButtons []Button

	// Drawing state
	drawing        bool
	startX, startY int32
	endX, endY     int32

	// Image position
	imgX, imgY int32
	imgW, imgH int32

	// Text input
	textInput   string
	textFocused bool
	inputRect   struct{ X, Y, W, H int32 }
	cursorBlink bool

	// Save/Cancel buttons
	saveBtn, cancelBtn Button

	// Result
	completed bool
	cancelled bool

	// Mouse
	mouseX, mouseY int32

	// Cached image bitmap data for fast rendering
	imageBitmapData []byte
	imageDataDirty  bool
}

// BITMAPINFOHEADER structure
type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type bitmapInfo struct {
	BmiHeader bitmapInfoHeader
}

var state editorState

func EditImageInteractive(img image.Image) (image.Image, bool) {
	state = editorState{
		editor:         NewEditor(img),
		currentTool:    TOOL_NONE,
		currentColor:   color.RGBA{255, 50, 50, 255},
		textInput:      "",
		imageDataDirty: true,
	}

	screenW, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenH, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	imgBounds := img.Bounds()
	state.imgW = int32(imgBounds.Dx())
	state.imgH = int32(imgBounds.Dy())

	// Window size
	state.winW = state.imgW + 60
	state.winH = state.imgH + 160

	maxW := int32(screenW) * 92 / 100
	maxH := int32(screenH) * 92 / 100

	if state.winW > maxW {
		state.winW = maxW
	}
	if state.winH > maxH {
		state.winH = maxH
	}
	if state.winW < 750 {
		state.winW = 750
	}
	if state.winH < 500 {
		state.winH = 500
	}

	// Image position (centered)
	state.imgX = (state.winW - state.imgW) / 2
	state.imgY = 120

	// Make sure image fits
	if state.imgY+state.imgH+10 > state.winH {
		state.imgH = state.winH - state.imgY - 10
	}
	if state.imgX+state.imgW+10 > state.winW {
		state.imgW = state.winW - state.imgX - 10
	}

	winX := (int32(screenW) - state.winW) / 2
	winY := (int32(screenH) - state.winH) / 2

	initButtons()

	hInstance, _, _ := procGetModuleHandle.Call(0)
	className := syscall.StringToUTF16Ptr("CustomScreenshotEditor")

	bgBrush, _, _ := procCreateSolidBrush.Call(colorBg)

	wndClass := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		Style:         3,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       loadCursor(IDC_ARROW),
		HbrBackground: bgBrush,
		LpszClassName: className,
	}

	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wndClass)))

	hwnd, _, _ := procCreateWindowEx.Call(
		WS_EX_TOPMOST,
		uintptr(unsafe.Pointer(className)),
		0,
		WS_POPUP|WS_VISIBLE,
		uintptr(winX), uintptr(winY),
		uintptr(state.winW), uintptr(state.winH),
		0, 0, hInstance, 0,
	)

	if hwnd == 0 {
		return img, false
	}

	state.hwnd = hwnd

	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)

	var m msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 || int32(ret) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))

		if state.completed || state.cancelled {
			break
		}
	}

	procDestroyWindow.Call(hwnd)
	procDeleteObject.Call(bgBrush)

	if state.cancelled {
		return img, false
	}
	return state.editor.GetImage(), true
}

func initButtons() {
	x := int32(20)
	y := int32(15)
	btnW := int32(90)
	btnH := int32(36)
	gap := int32(8)

	// Tool buttons
	tools := []struct {
		text string
		tool int
	}{
		{"✏ Text", TOOL_TEXT},
		{"▢ Rect", TOOL_RECT},
		{"➔ Arrow", TOOL_ARROW},
		{"▨ Mark", TOOL_HIGHLIGHT},
		{"✎ Draw", TOOL_FREEDRAW},
	}

	for _, t := range tools {
		state.buttons = append(state.buttons, Button{
			X: x, Y: y, W: btnW, H: btnH,
			Text: t.text, ToolID: t.tool,
		})
		x += btnW + gap
	}

	// Undo button
	state.buttons = append(state.buttons, Button{
		X: x, Y: y, W: 70, H: btnH,
		Text: "↺ Undo", ToolID: -1,
	})

	// Text input area
	state.inputRect.X = 20
	state.inputRect.Y = y + btnH + 12
	state.inputRect.W = 350
	state.inputRect.H = 32

	// Color palette
	colorY := state.inputRect.Y + 3
	colorX := state.inputRect.X + state.inputRect.W + 20
	colorSize := int32(26)

	for i, c := range paletteColors {
		state.colorButtons = append(state.colorButtons, Button{
			X:      colorX + int32(i)*(colorSize+4),
			Y:      colorY,
			W:      colorSize,
			H:      colorSize,
			ToolID: i,
		})
		_ = c
	}

	// Save & Cancel
	state.saveBtn = Button{
		X: state.winW - 180, Y: state.inputRect.Y,
		W: 80, H: 32, Text: "✓ Save",
	}
	state.cancelBtn = Button{
		X: state.winW - 90, Y: state.inputRect.Y,
		W: 80, H: 32, Text: "✕ Cancel",
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps paintStruct
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		drawUI(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_MOUSEMOVE:
		state.mouseX = int32(lParam & 0xFFFF)
		state.mouseY = int32((lParam >> 16) & 0xFFFF)
		handleMouseMove()
		return 0

	case WM_LBUTTONDOWN:
		handleMouseDown()
		return 0

	case WM_LBUTTONUP:
		handleMouseUp()
		return 0

	case WM_KEYDOWN:
		handleKeyDown(int(wParam))
		return 0

	case WM_CHAR:
		handleChar(rune(wParam))
		return 0

	case WM_CLOSE, WM_DESTROY:
		state.cancelled = true
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func handleMouseMove() {
	// Update button hover states
	needsRepaint := false
	for i := range state.buttons {
		wasHovered := state.buttons[i].IsHovered
		state.buttons[i].IsHovered = pointInRect(state.mouseX, state.mouseY,
			state.buttons[i].X, state.buttons[i].Y,
			state.buttons[i].W, state.buttons[i].H)
		if wasHovered != state.buttons[i].IsHovered {
			needsRepaint = true
		}
	}

	wasHovered := state.saveBtn.IsHovered
	state.saveBtn.IsHovered = pointInRect(state.mouseX, state.mouseY,
		state.saveBtn.X, state.saveBtn.Y, state.saveBtn.W, state.saveBtn.H)
	if wasHovered != state.saveBtn.IsHovered {
		needsRepaint = true
	}

	wasHovered = state.cancelBtn.IsHovered
	state.cancelBtn.IsHovered = pointInRect(state.mouseX, state.mouseY,
		state.cancelBtn.X, state.cancelBtn.Y, state.cancelBtn.W, state.cancelBtn.H)
	if wasHovered != state.cancelBtn.IsHovered {
		needsRepaint = true
	}

	// Drawing preview
	if state.drawing {
		state.endX = state.mouseX - state.imgX
		state.endY = state.mouseY - state.imgY
		needsRepaint = true
	}

	if needsRepaint {
		procInvalidateRect.Call(state.hwnd, 0, 0)
	}
}

func handleMouseDown() {
	x, y := state.mouseX, state.mouseY

	// Check tool buttons
	for i := range state.buttons {
		if state.buttons[i].IsHovered {
			if state.buttons[i].ToolID == -1 {
				// Undo
				state.editor.Reset()
				state.imageDataDirty = true
			} else {
				// Select tool
				state.currentTool = state.buttons[i].ToolID
				for j := range state.buttons {
					state.buttons[j].IsActive = (state.buttons[j].ToolID == state.currentTool)
				}
			}
			state.textFocused = false
			procInvalidateRect.Call(state.hwnd, 0, 0)
			return
		}
	}

	// Check color buttons
	for i := range state.colorButtons {
		if pointInRect(x, y, state.colorButtons[i].X, state.colorButtons[i].Y,
			state.colorButtons[i].W, state.colorButtons[i].H) {
			state.currentColor = paletteColors[i]
			procInvalidateRect.Call(state.hwnd, 0, 0)
			return
		}
	}

	// Check text input
	if pointInRect(x, y, state.inputRect.X, state.inputRect.Y,
		state.inputRect.W, state.inputRect.H) {
		state.textFocused = true
		procInvalidateRect.Call(state.hwnd, 0, 0)
		return
	}
	state.textFocused = false

	// Check save/cancel
	if state.saveBtn.IsHovered {
		state.completed = true
		procPostQuitMessage.Call(0)
		return
	}
	if state.cancelBtn.IsHovered {
		state.cancelled = true
		procPostQuitMessage.Call(0)
		return
	}

	// Check image area for drawing
	imgX := x - state.imgX
	imgY := y - state.imgY

	if imgX >= 0 && imgY >= 0 && imgX < state.imgW && imgY < state.imgH {
		if state.currentTool == TOOL_TEXT {
			// Add text at click position
			if state.textInput != "" {
				state.editor.AddTextWithBackground(state.textInput, int(imgX), int(imgY),
					ColorWhite, color.RGBA{0, 0, 0, 200})
				state.imageDataDirty = true
				procInvalidateRect.Call(state.hwnd, 0, 0)
			}
		} else if state.currentTool != TOOL_NONE {
			state.drawing = true
			state.startX = imgX
			state.startY = imgY
			state.endX = imgX
			state.endY = imgY
			procSetCapture.Call(state.hwnd)
		}
	}
}

func handleMouseUp() {
	if !state.drawing {
		return
	}

	state.drawing = false
	procReleaseCapture.Call()

	x1, y1 := int(state.startX), int(state.startY)
	x2, y2 := int(state.endX), int(state.endY)

	// Ensure within bounds
	bounds := state.editor.Image.Bounds()
	clamp := func(v, max int) int {
		if v < 0 {
			return 0
		}
		if v > max {
			return max
		}
		return v
	}
	x1 = clamp(x1, bounds.Dx())
	y1 = clamp(y1, bounds.Dy())
	x2 = clamp(x2, bounds.Dx())
	y2 = clamp(y2, bounds.Dy())

	switch state.currentTool {
	case TOOL_RECT:
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		state.editor.DrawRectangle(x1, y1, x2, y2, state.currentColor, 3)
	case TOOL_ARROW:
		state.editor.DrawArrow(x1, y1, x2, y2, state.currentColor, 3)
	case TOOL_HIGHLIGHT:
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		state.editor.Highlight(x1, y1, x2, y2, state.currentColor)
	case TOOL_FREEDRAW:
		state.editor.DrawLine(x1, y1, x2, y2, state.currentColor, 3)
	}

	state.imageDataDirty = true
	procInvalidateRect.Call(state.hwnd, 0, 0)
}

func handleKeyDown(key int) {
	switch key {
	case VK_ESCAPE:
		state.cancelled = true
		procPostQuitMessage.Call(0)
	case VK_RETURN:
		if state.textFocused && state.textInput != "" {
			// Confirm text entry
			state.textFocused = false
		}
	case VK_BACK:
		if state.textFocused && len(state.textInput) > 0 {
			state.textInput = state.textInput[:len(state.textInput)-1]
			procInvalidateRect.Call(state.hwnd, 0, 0)
		}
	}
}

func handleChar(ch rune) {
	if state.textFocused && ch >= 32 && ch != 127 {
		state.textInput += string(ch)
		procInvalidateRect.Call(state.hwnd, 0, 0)
	}
}

func drawUI(hdc uintptr) {
	// Draw toolbar background
	drawFilledRect(hdc, 0, 0, state.winW, 110, colorToolbar)

	// Draw title
	drawText(hdc, "SCREENSHOT EDITOR", 20, state.winH-30, colorTextHint, false)

	// Draw tool buttons
	for _, btn := range state.buttons {
		bgColor := colorBtnNormal
		if btn.IsActive {
			bgColor = colorBtnActive
		} else if btn.IsHovered {
			bgColor = colorBtnHover
		}
		drawRoundedButton(hdc, btn.X, btn.Y, btn.W, btn.H, btn.Text, bgColor)
	}

	// Draw text input
	drawTextInput(hdc)

	// Draw color palette
	drawColorPalette(hdc)

	// Draw save/cancel
	saveBg := colorBtnNormal
	if state.saveBtn.IsHovered {
		saveBg = rgbToRef(50, 180, 100)
	}
	drawRoundedButton(hdc, state.saveBtn.X, state.saveBtn.Y,
		state.saveBtn.W, state.saveBtn.H, state.saveBtn.Text, saveBg)

	cancelBg := colorBtnNormal
	if state.cancelBtn.IsHovered {
		cancelBg = rgbToRef(200, 60, 60)
	}
	drawRoundedButton(hdc, state.cancelBtn.X, state.cancelBtn.Y,
		state.cancelBtn.W, state.cancelBtn.H, state.cancelBtn.Text, cancelBg)

	// Draw image
	drawImage(hdc)

	// Draw preview overlay
	if state.drawing {
		drawPreview(hdc)
	}

	// Draw status
	status := getStatusText()
	drawText(hdc, status, state.imgX, state.imgY+state.imgH+5, colorTextHint, false)
}

func drawRoundedButton(hdc uintptr, x, y, w, h int32, text string, bgColor uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(bgColor)
	pen, _, _ := procCreatePen.Call(PS_SOLID, 1, bgColor)

	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	procRoundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), 8, 8)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)

	// Draw text centered
	textX := x + (w-int32(len(text)*7))/2
	textY := y + (h-14)/2
	drawText(hdc, text, textX, textY, colorWhite, false)
}

func drawTextInput(hdc uintptr) {
	r := state.inputRect

	// Background
	brush, _, _ := procCreateSolidBrush.Call(colorInputBg)
	borderColor := colorInputBorder
	if state.textFocused {
		borderColor = colorAccent
	}
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, borderColor)

	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	procRoundRect.Call(hdc, uintptr(r.X), uintptr(r.Y), uintptr(r.X+r.W), uintptr(r.Y+r.H), 6, 6)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)

	// Text or placeholder
	textX := r.X + 10
	textY := r.Y + 8
	if state.textInput == "" {
		drawText(hdc, "Type text here, then click on image...", textX, textY, colorTextHint, false)
	} else {
		text := state.textInput
		if state.textFocused {
			text += "|" // Cursor
		}
		drawText(hdc, text, textX, textY, colorWhite, false)
	}
}

func drawColorPalette(hdc uintptr) {
	for i, btn := range state.colorButtons {
		c := paletteColors[i]
		colorRef := rgbToRef(c.R, c.G, c.B)

		// Selection ring
		if c == state.currentColor {
			ringPen, _, _ := procCreatePen.Call(PS_SOLID, 3, colorWhite)
			oldPen, _, _ := procSelectObject.Call(hdc, ringPen)
			nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
			oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)
			procRoundRect.Call(hdc, uintptr(btn.X-2), uintptr(btn.Y-2),
				uintptr(btn.X+btn.W+2), uintptr(btn.Y+btn.H+2), 6, 6)
			procSelectObject.Call(hdc, oldPen)
			procSelectObject.Call(hdc, oldBrush)
			procDeleteObject.Call(ringPen)
		}

		// Color square
		brush, _, _ := procCreateSolidBrush.Call(colorRef)
		pen, _, _ := procCreatePen.Call(PS_SOLID, 1, rgbToRef(80, 80, 90))

		oldBrush, _, _ := procSelectObject.Call(hdc, brush)
		oldPen, _, _ := procSelectObject.Call(hdc, pen)

		procRoundRect.Call(hdc, uintptr(btn.X), uintptr(btn.Y),
			uintptr(btn.X+btn.W), uintptr(btn.Y+btn.H), 4, 4)

		procSelectObject.Call(hdc, oldBrush)
		procSelectObject.Call(hdc, oldPen)
		procDeleteObject.Call(brush)
		procDeleteObject.Call(pen)
	}
}

func drawImage(hdc uintptr) {
	img := state.editor.Image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Draw border
	borderPen, _, _ := procCreatePen.Call(PS_SOLID, 2, colorInputBorder)
	oldPen, _, _ := procSelectObject.Call(hdc, borderPen)
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)
	procRectangle.Call(hdc, uintptr(state.imgX-2), uintptr(state.imgY-2),
		uintptr(state.imgX+int32(width)+2), uintptr(state.imgY+int32(height)+2))
	procSelectObject.Call(hdc, oldPen)
	procSelectObject.Call(hdc, oldBrush)
	procDeleteObject.Call(borderPen)

	// Update cached bitmap data if needed
	if state.imageDataDirty || state.imageBitmapData == nil {
		updateImageCache()
	}

	// Draw using SetDIBitsToDevice - much faster than SetPixel!
	bmi := bitmapInfo{
		BmiHeader: bitmapInfoHeader{
			BiSize:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			BiWidth:       int32(width),
			BiHeight:      -int32(height), // Negative for top-down DIB
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: BI_RGB,
		},
	}

	procSetDIBitsToDevice.Call(
		hdc,
		uintptr(state.imgX), uintptr(state.imgY),
		uintptr(width), uintptr(height),
		0, 0,
		0, uintptr(height),
		uintptr(unsafe.Pointer(&state.imageBitmapData[0])),
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
	)
}

func updateImageCache() {
	img := state.editor.Image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Allocate buffer for BGRA data (4 bytes per pixel)
	dataSize := width * height * 4
	if len(state.imageBitmapData) != dataSize {
		state.imageBitmapData = make([]byte, dataSize)
	}

	// Convert RGBA to BGRA (Windows DIB format)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.RGBAAt(x+bounds.Min.X, y+bounds.Min.Y)
			i := (y*width + x) * 4
			state.imageBitmapData[i+0] = c.B   // Blue
			state.imageBitmapData[i+1] = c.G   // Green
			state.imageBitmapData[i+2] = c.R   // Red
			state.imageBitmapData[i+3] = c.A   // Alpha
		}
	}

	state.imageDataDirty = false
}

func drawPreview(hdc uintptr) {
	x1 := state.startX + state.imgX
	y1 := state.startY + state.imgY
	x2 := state.endX + state.imgX
	y2 := state.endY + state.imgY

	colorRef := rgbToRef(state.currentColor.R, state.currentColor.G, state.currentColor.B)
	pen, _, _ := procCreatePen.Call(PS_SOLID, 2, colorRef)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	nullBrush, _, _ := procGetStockObject.Call(NULL_BRUSH)
	oldBrush, _, _ := procSelectObject.Call(hdc, nullBrush)

	switch state.currentTool {
	case TOOL_RECT, TOOL_HIGHLIGHT:
		procRectangle.Call(hdc, uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
	case TOOL_ARROW, TOOL_FREEDRAW:
		procMoveToEx.Call(hdc, uintptr(x1), uintptr(y1), 0)
		procLineTo.Call(hdc, uintptr(x2), uintptr(y2))
	}

	procSelectObject.Call(hdc, oldPen)
	procSelectObject.Call(hdc, oldBrush)
	procDeleteObject.Call(pen)
}

func drawFilledRect(hdc uintptr, x, y, w, h int32, colorRef uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(colorRef)
	rect := struct{ Left, Top, Right, Bottom int32 }{x, y, x + w, y + h}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rect)), brush)
	procDeleteObject.Call(brush)
}

func drawText(hdc uintptr, text string, x, y int32, colorRef uintptr, bold bool) {
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, colorRef)

	weight := int32(FW_NORMAL)
	if bold {
		weight = FW_BOLD
	}

	font, _, _ := procCreateFont.Call(
		14, 0, 0, 0, uintptr(weight),
		0, 0, 0, 0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Segoe UI"))),
	)
	oldFont, _, _ := procSelectObject.Call(hdc, font)

	textPtr := syscall.StringToUTF16Ptr(text)
	procTextOut.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(textPtr)), uintptr(len(text)))

	procSelectObject.Call(hdc, oldFont)
	procDeleteObject.Call(font)
}

func getStatusText() string {
	switch state.currentTool {
	case TOOL_TEXT:
		return "TEXT: Click on image to place text"
	case TOOL_RECT:
		return "RECTANGLE: Drag to draw"
	case TOOL_ARROW:
		return "ARROW: Drag to draw"
	case TOOL_HIGHLIGHT:
		return "HIGHLIGHT: Drag to mark area"
	case TOOL_FREEDRAW:
		return "FREEDRAW: Draw freely"
	default:
		return "Select a tool to start editing | ESC to cancel"
	}
}

func pointInRect(px, py, x, y, w, h int32) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}

func rgbToRef(r, g, b uint8) uintptr {
	return uintptr(uint32(b)<<16 | uint32(g)<<8 | uint32(r))
}

func loadCursor(id int) uintptr {
	cursor, _, _ := procLoadCursor.Call(0, uintptr(id))
	return cursor
}
