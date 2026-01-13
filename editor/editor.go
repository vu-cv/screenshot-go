package editor

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// TextAnnotation represents a text overlay on the image
type TextAnnotation struct {
	Text            string
	X, Y            int
	Color           color.Color
	FontSize        int
	BackgroundColor color.Color
	HasBackground   bool
}

// Editor provides image editing capabilities
type Editor struct {
	Image       *image.RGBA
	Original    image.Image
	Annotations []TextAnnotation
}

// NewEditor creates a new editor from an image
func NewEditor(img image.Image) *Editor {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	return &Editor{
		Image:       rgba,
		Original:    img,
		Annotations: make([]TextAnnotation, 0),
	}
}

// AddText adds a text annotation to the image
func (e *Editor) AddText(text string, x, y int, col color.Color) {
	annotation := TextAnnotation{
		Text:  text,
		X:     x,
		Y:     y,
		Color: col,
	}
	e.Annotations = append(e.Annotations, annotation)
	e.drawText(annotation)
}

// AddTextWithBackground adds text with a background box
func (e *Editor) AddTextWithBackground(text string, x, y int, textColor, bgColor color.Color) {
	annotation := TextAnnotation{
		Text:            text,
		X:               x,
		Y:               y,
		Color:           textColor,
		BackgroundColor: bgColor,
		HasBackground:   true,
	}
	e.Annotations = append(e.Annotations, annotation)
	e.drawTextWithBackground(annotation)
}

// drawText draws text on the image using basic font
func (e *Editor) drawText(ann TextAnnotation) {
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  e.Image,
		Src:  image.NewUniform(ann.Color),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(ann.X), Y: fixed.I(ann.Y)},
	}
	d.DrawString(ann.Text)
}

// drawTextWithBackground draws text with a background rectangle
func (e *Editor) drawTextWithBackground(ann TextAnnotation) {
	face := basicfont.Face7x13

	// Calculate text dimensions
	textWidth := font.MeasureString(face, ann.Text).Ceil()
	textHeight := face.Metrics().Height.Ceil()
	padding := 4

	// Draw background rectangle
	bgRect := image.Rect(
		ann.X-padding,
		ann.Y-textHeight-padding+3,
		ann.X+textWidth+padding,
		ann.Y+padding,
	)
	draw.Draw(e.Image, bgRect, image.NewUniform(ann.BackgroundColor), image.Point{}, draw.Src)

	// Draw text
	d := &font.Drawer{
		Dst:  e.Image,
		Src:  image.NewUniform(ann.Color),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(ann.X), Y: fixed.I(ann.Y)},
	}
	d.DrawString(ann.Text)
}

// DrawRectangle draws a rectangle outline on the image
func (e *Editor) DrawRectangle(x1, y1, x2, y2 int, col color.Color, thickness int) {
	// Top line
	for t := 0; t < thickness; t++ {
		for x := x1; x <= x2; x++ {
			e.Image.Set(x, y1+t, col)
		}
	}
	// Bottom line
	for t := 0; t < thickness; t++ {
		for x := x1; x <= x2; x++ {
			e.Image.Set(x, y2-t, col)
		}
	}
	// Left line
	for t := 0; t < thickness; t++ {
		for y := y1; y <= y2; y++ {
			e.Image.Set(x1+t, y, col)
		}
	}
	// Right line
	for t := 0; t < thickness; t++ {
		for y := y1; y <= y2; y++ {
			e.Image.Set(x2-t, y, col)
		}
	}
}

// DrawLine draws a line on the image using Bresenham's algorithm
func (e *Editor) DrawLine(x1, y1, x2, y2 int, col color.Color, thickness int) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		// Draw thick point
		for tx := -thickness / 2; tx <= thickness/2; tx++ {
			for ty := -thickness / 2; ty <= thickness/2; ty++ {
				e.Image.Set(x1+tx, y1+ty, col)
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

// DrawArrow draws an arrow from (x1,y1) to (x2,y2)
func (e *Editor) DrawArrow(x1, y1, x2, y2 int, col color.Color, thickness int) {
	// Draw main line
	e.DrawLine(x1, y1, x2, y2, col, thickness)

	// Calculate arrow head
	arrowLength := 15.0
	arrowAngle := 0.5 // radians (~30 degrees)

	// Direction vector
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	length := sqrt(dx*dx + dy*dy)
	if length == 0 {
		return
	}
	dx /= length
	dy /= length

	// Arrow head points
	ax1 := int(float64(x2) - arrowLength*(dx*cos(arrowAngle)+dy*sin(arrowAngle)))
	ay1 := int(float64(y2) - arrowLength*(dy*cos(arrowAngle)-dx*sin(arrowAngle)))
	ax2 := int(float64(x2) - arrowLength*(dx*cos(arrowAngle)-dy*sin(arrowAngle)))
	ay2 := int(float64(y2) - arrowLength*(dy*cos(arrowAngle)+dx*sin(arrowAngle)))

	e.DrawLine(x2, y2, ax1, ay1, col, thickness)
	e.DrawLine(x2, y2, ax2, ay2, col, thickness)
}

// Highlight draws a semi-transparent highlight rectangle
func (e *Editor) Highlight(x1, y1, x2, y2 int, col color.Color) {
	// Get alpha from color
	r, g, b, _ := col.RGBA()
	highlightColor := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 100}

	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			orig := e.Image.RGBAAt(x, y)
			blended := blendColors(orig, highlightColor)
			e.Image.Set(x, y, blended)
		}
	}
}

// GetImage returns the edited image
func (e *Editor) GetImage() image.Image {
	return e.Image
}

// Reset resets to original image
func (e *Editor) Reset() {
	bounds := e.Original.Bounds()
	draw.Draw(e.Image, bounds, e.Original, bounds.Min, draw.Src)
	e.Annotations = nil
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

func cos(x float64) float64 {
	// Taylor series approximation
	x2 := x * x
	return 1 - x2/2 + x2*x2/24
}

func sin(x float64) float64 {
	// Taylor series approximation
	x2 := x * x
	return x - x*x2/6 + x*x2*x2/120
}

func blendColors(base, overlay color.RGBA) color.RGBA {
	alpha := float64(overlay.A) / 255.0
	return color.RGBA{
		R: uint8(float64(base.R)*(1-alpha) + float64(overlay.R)*alpha),
		G: uint8(float64(base.G)*(1-alpha) + float64(overlay.G)*alpha),
		B: uint8(float64(base.B)*(1-alpha) + float64(overlay.B)*alpha),
		A: 255,
	}
}

// Predefined colors
var (
	ColorRed    = color.RGBA{255, 0, 0, 255}
	ColorGreen  = color.RGBA{0, 255, 0, 255}
	ColorBlue   = color.RGBA{0, 0, 255, 255}
	ColorYellow = color.RGBA{255, 255, 0, 255}
	ColorWhite  = color.RGBA{255, 255, 255, 255}
	ColorBlack  = color.RGBA{0, 0, 0, 255}
	ColorOrange = color.RGBA{255, 165, 0, 255}
	ColorPink   = color.RGBA{255, 105, 180, 255}
	ColorCyan   = color.RGBA{0, 255, 255, 255}
)

// ParseColor parses a color string (name or hex)
func ParseColor(s string) (color.Color, error) {
	switch s {
	case "red":
		return ColorRed, nil
	case "green":
		return ColorGreen, nil
	case "blue":
		return ColorBlue, nil
	case "yellow":
		return ColorYellow, nil
	case "white":
		return ColorWhite, nil
	case "black":
		return ColorBlack, nil
	case "orange":
		return ColorOrange, nil
	case "pink":
		return ColorPink, nil
	case "cyan":
		return ColorCyan, nil
	default:
		// Try to parse as hex color #RRGGBB
		if len(s) == 7 && s[0] == '#' {
			var r, g, b uint8
			_, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
			if err == nil {
				return color.RGBA{r, g, b, 255}, nil
			}
		}
		return ColorRed, fmt.Errorf("unknown color: %s", s)
	}
}
