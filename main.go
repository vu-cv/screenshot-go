package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vucv/screenshot/capture"
	"vucv/screenshot/clipboard"
	"vucv/screenshot/config"
	"vucv/screenshot/editor"
	"vucv/screenshot/hotkey"
	"vucv/screenshot/selector"
)

var (
	version = "1.0.0"
)

func main() {
	// CLI flags
	fullscreen := flag.Bool("full", false, "Capture full screen (all displays)")
	display := flag.Int("display", -1, "Capture specific display (0-indexed)")
	region := flag.String("region", "", "Capture region: x,y,width,height")
	selectRegion := flag.Bool("select", false, "Interactive region selection with mouse")
	editMode := flag.Bool("edit", false, "Open editor after capture to add annotations")
	output := flag.String("output", "", "Output directory")
	format := flag.String("format", "png", "Output format: png or jpg")
	quality := flag.Int("quality", 90, "JPEG quality (1-100)")
	delay := flag.Duration("delay", 0, "Delay before capture (e.g., 3s)")
	copyClip := flag.Bool("copy", true, "Copy to clipboard")
	daemon := flag.Bool("daemon", false, "Run in daemon mode with hotkeys")
	listDisplays := flag.Bool("list-displays", false, "List available displays")
	showVersion := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *showVersion {
		fmt.Printf("Screenshot Tool v%s\n", version)
		return
	}

	if *listDisplays {
		printDisplays()
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Override config with CLI flags
	if *output != "" {
		cfg.OutputDir = *output
	}
	cfg.Format = *format
	cfg.Quality = *quality

	// Create capture options
	opts := &capture.CaptureOptions{
		OutputDir: cfg.OutputDir,
		Format:    cfg.Format,
		Quality:   cfg.Quality,
		Delay:     *delay,
	}

	// Daemon mode - run with hotkeys
	if *daemon {
		runDaemon(cfg)
		return
	}

	// Single capture mode
	var result *capture.Result

	switch {
	case *selectRegion:
		// Interactive region selection
		fmt.Println("Select region with mouse (ESC to cancel)...")
		rect, ok := selector.SelectRegion()
		if !ok {
			fmt.Println("Selection cancelled.")
			return
		}
		result, err = capture.Region(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Open editor if edit mode is enabled
		if *editMode {
			fmt.Println("Opening editor...")
			editedImg, ok := editor.EditImageInteractive(result.Image)
			if ok {
				result.Image = editedImg
			}
		}

	case *region != "":
		var x, y, w, h int
		_, err := fmt.Sscanf(*region, "%d,%d,%d,%d", &x, &y, &w, &h)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid region format. Use: x,y,width,height\n")
			os.Exit(1)
		}
		result, err = capture.Region(x, y, w, h, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case *display >= 0:
		result, err = capture.Display(*display, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case *fullscreen:
		result, err = capture.FullScreen(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		// Default: capture primary display
		result, err = capture.Display(0, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Save the screenshot
	filepath, err := result.Save(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving screenshot: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Screenshot saved: %s\n", filepath)

	// Copy to clipboard if enabled
	if *copyClip {
		if err := clipboard.CopyImage(result.Image); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy to clipboard: %v\n", err)
		} else {
			fmt.Println("Screenshot copied to clipboard")
		}
	}
}

func printHelp() {
	fmt.Println(`Screenshot Tool - Full-featured screenshot utility

Usage:
  screenshot [options]

Options:
  -full              Capture all displays (full screen)
  -display N         Capture specific display (0-indexed)
  -region X,Y,W,H    Capture specific region
  -select            Interactive region selection with mouse drag
  -edit              Open editor after capture to add text/annotations
  -output DIR        Output directory
  -format FORMAT     Output format: png (default) or jpg
  -quality N         JPEG quality 1-100 (default: 90)
  -delay DURATION    Delay before capture (e.g., 3s, 500ms)
  -copy              Copy to clipboard (default: true)
  -daemon            Run in daemon mode with global hotkeys
  -list-displays     List available displays
  -version           Show version
  -help              Show this help

Hotkeys (daemon mode):
  Ctrl+Shift+S       Capture full screen
  Ctrl+Shift+R       Capture region (interactive)
  Ctrl+Shift+W       Capture active window
  Ctrl+Shift+Q       Quit daemon

Examples:
  screenshot -full                    # Capture all screens
  screenshot -display 0               # Capture primary monitor
  screenshot -region 100,100,800,600  # Capture region
  screenshot -select -edit            # Select region and add annotations
  screenshot -delay 3s -full          # Capture after 3 seconds
  screenshot -daemon                  # Run with hotkeys`)
}

func printDisplays() {
	n := capture.GetDisplayCount()
	fmt.Printf("Found %d display(s):\n\n", n)

	for i := 0; i < n; i++ {
		bounds := capture.GetDisplayBounds(i)
		fmt.Printf("  Display %d:\n", i)
		fmt.Printf("    Position: (%d, %d)\n", bounds.Min.X, bounds.Min.Y)
		fmt.Printf("    Size: %dx%d\n", bounds.Dx(), bounds.Dy())
		fmt.Println()
	}

	allBounds := capture.GetAllDisplaysBounds()
	fmt.Printf("  Combined virtual screen:\n")
	fmt.Printf("    Position: (%d, %d)\n", allBounds.Min.X, allBounds.Min.Y)
	fmt.Printf("    Size: %dx%d\n", allBounds.Dx(), allBounds.Dy())
}

func runDaemon(cfg *config.Config) {
	fmt.Println("Screenshot Tool - Daemon Mode")
	fmt.Println("==============================")
	fmt.Println("Hotkeys:")
	fmt.Printf("  %s - Full screen capture\n", cfg.HotkeyFullScreen)
	fmt.Printf("  %s - Region capture\n", cfg.HotkeyRegion)
	fmt.Printf("  %s - Active window capture\n", cfg.HotkeyActiveWindow)
	fmt.Println("  Ctrl+Shift+Q - Quit")
	fmt.Println()
	fmt.Println("Waiting for hotkeys... Press Ctrl+C to exit.")

	hkManager := hotkey.NewManager()

	// Register hotkeys
	opts := &capture.CaptureOptions{
		OutputDir: cfg.OutputDir,
		Format:    cfg.Format,
		Quality:   cfg.Quality,
	}

	// Full screen capture
	hkManager.Register(cfg.HotkeyFullScreen, func() {
		fmt.Println("\n[Hotkey] Capturing full screen...")
		result, err := capture.FullScreen(opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		filepath, err := result.Save(opts)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
			return
		}
		fmt.Printf("Saved: %s\n", filepath)

		if cfg.CopyToClipboard {
			clipboard.CopyImage(result.Image)
			fmt.Println("Copied to clipboard")
		}
	})

	// Region capture with mouse selection
	hkManager.Register(cfg.HotkeyRegion, func() {
		fmt.Println("\n[Hotkey] Select region with mouse...")
		rect, ok := selector.SelectRegion()
		if !ok {
			fmt.Println("Selection cancelled.")
			return
		}

		result, err := capture.Region(rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		filepath, err := result.Save(opts)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
			return
		}
		fmt.Printf("Saved: %s\n", filepath)

		if cfg.CopyToClipboard {
			clipboard.CopyImage(result.Image)
			fmt.Println("Copied to clipboard")
		}
	})

	// Active window (primary display for now)
	hkManager.Register(cfg.HotkeyActiveWindow, func() {
		fmt.Println("\n[Hotkey] Capturing primary display...")
		result, err := capture.Display(0, opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		filepath, err := result.Save(opts)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
			return
		}
		fmt.Printf("Saved: %s\n", filepath)

		if cfg.CopyToClipboard {
			clipboard.CopyImage(result.Image)
			fmt.Println("Copied to clipboard")
		}
	})

	// Quit hotkey
	quitCh := make(chan struct{})
	hkManager.Register("ctrl+shift+q", func() {
		fmt.Println("\n[Hotkey] Quitting...")
		close(quitCh)
	})

	// Start hotkey listener
	if err := hkManager.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting hotkey manager: %v\n", err)
		os.Exit(1)
	}
	defer hkManager.Stop()

	// Wait for quit signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		fmt.Println("\nReceived interrupt signal, exiting...")
	case <-quitCh:
		fmt.Println("Goodbye!")
	}

	// Small delay for cleanup
	time.Sleep(100 * time.Millisecond)
}
