package capture

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/kbinani/screenshot"
)

// CaptureOptions holds configuration for screenshot capture
type CaptureOptions struct {
	OutputDir    string
	Format       string // "png", "jpg"
	Quality      int    // JPEG quality (1-100)
	Delay        time.Duration
	FilenameFunc func() string
}

// DefaultOptions returns default capture options
func DefaultOptions() *CaptureOptions {
	return &CaptureOptions{
		OutputDir: ".",
		Format:    "png",
		Quality:   90,
		Delay:     0,
		FilenameFunc: func() string {
			return fmt.Sprintf("screenshot_%s", time.Now().Format("20060102_150405"))
		},
	}
}

// Result holds the capture result
type Result struct {
	Image    image.Image
	Bounds   image.Rectangle
	FilePath string
}

// GetDisplayCount returns the number of active displays
func GetDisplayCount() int {
	return screenshot.NumActiveDisplays()
}

// GetDisplayBounds returns the bounds of a specific display
func GetDisplayBounds(displayIndex int) image.Rectangle {
	return screenshot.GetDisplayBounds(displayIndex)
}

// GetAllDisplaysBounds returns combined bounds of all displays
func GetAllDisplaysBounds() image.Rectangle {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return image.Rectangle{}
	}

	var minX, minY, maxX, maxY int
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		if i == 0 {
			minX, minY = bounds.Min.X, bounds.Min.Y
			maxX, maxY = bounds.Max.X, bounds.Max.Y
		} else {
			if bounds.Min.X < minX {
				minX = bounds.Min.X
			}
			if bounds.Min.Y < minY {
				minY = bounds.Min.Y
			}
			if bounds.Max.X > maxX {
				maxX = bounds.Max.X
			}
			if bounds.Max.Y > maxY {
				maxY = bounds.Max.Y
			}
		}
	}
	return image.Rect(minX, minY, maxX, maxY)
}

// FullScreen captures all displays
func FullScreen(opts *CaptureOptions) (*Result, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if opts.Delay > 0 {
		time.Sleep(opts.Delay)
	}

	bounds := GetAllDisplaysBounds()
	img, err := screenshot.Capture(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
	if err != nil {
		return nil, fmt.Errorf("failed to capture screen: %w", err)
	}

	return &Result{
		Image:  img,
		Bounds: bounds,
	}, nil
}

// Display captures a specific display by index
func Display(displayIndex int, opts *CaptureOptions) (*Result, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	n := screenshot.NumActiveDisplays()
	if displayIndex < 0 || displayIndex >= n {
		return nil, fmt.Errorf("invalid display index %d (available: 0-%d)", displayIndex, n-1)
	}

	if opts.Delay > 0 {
		time.Sleep(opts.Delay)
	}

	bounds := screenshot.GetDisplayBounds(displayIndex)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, fmt.Errorf("failed to capture display %d: %w", displayIndex, err)
	}

	return &Result{
		Image:  img,
		Bounds: bounds,
	}, nil
}

// Region captures a specific region of the screen
func Region(x, y, width, height int, opts *CaptureOptions) (*Result, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid region size: width=%d, height=%d", width, height)
	}

	if opts.Delay > 0 {
		time.Sleep(opts.Delay)
	}

	img, err := screenshot.Capture(x, y, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to capture region: %w", err)
	}

	return &Result{
		Image:  img,
		Bounds: image.Rect(x, y, x+width, y+height),
	}, nil
}

// Save saves the captured image to a file
func (r *Result) Save(opts *CaptureOptions) (string, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Ensure FilenameFunc is set
	if opts.FilenameFunc == nil {
		opts.FilenameFunc = func() string {
			return fmt.Sprintf("screenshot_%s", time.Now().Format("20060102_150405"))
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filename
	filename := opts.FilenameFunc()
	ext := opts.Format
	if ext == "jpg" {
		ext = "jpg"
	} else {
		ext = "png"
	}
	fullPath := filepath.Join(opts.OutputDir, fmt.Sprintf("%s.%s", filename, ext))

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Encode and save
	switch opts.Format {
	case "jpg", "jpeg":
		err = jpeg.Encode(file, r.Image, &jpeg.Options{Quality: opts.Quality})
	default:
		err = png.Encode(file, r.Image)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	r.FilePath = fullPath
	return fullPath, nil
}
