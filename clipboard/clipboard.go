package clipboard

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os/exec"
	"runtime"
)

// CopyImage copies an image to the system clipboard
func CopyImage(img image.Image) error {
	switch runtime.GOOS {
	case "windows":
		return copyImageWindows(img)
	case "darwin":
		return copyImageMacOS(img)
	case "linux":
		return copyImageLinux(img)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// copyImageWindows copies image to clipboard on Windows using PowerShell
func copyImageWindows(img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Use PowerShell to copy image to clipboard
	script := `
Add-Type -AssemblyName System.Windows.Forms
$ms = New-Object System.IO.MemoryStream
$input | ForEach-Object { $ms.WriteByte($_) }
$ms.Position = 0
$bitmap = [System.Drawing.Bitmap]::FromStream($ms)
[System.Windows.Forms.Clipboard]::SetImage($bitmap)
$bitmap.Dispose()
$ms.Dispose()
`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.Stdin = &buf
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w, output: %s", err, string(output))
	}

	return nil
}

// copyImageMacOS copies image to clipboard on macOS
func copyImageMacOS(img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	cmd := exec.Command("osascript", "-e", `set the clipboard to (read (POSIX file "/dev/stdin") as «class PNGf»)`)
	cmd.Stdin = &buf

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w, output: %s", err, string(output))
	}

	return nil
}

// copyImageLinux copies image to clipboard on Linux using xclip
func copyImageLinux(img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Try xclip first, then xsel
	cmd := exec.Command("xclip", "-selection", "clipboard", "-t", "image/png")
	cmd.Stdin = &buf

	if err := cmd.Run(); err != nil {
		// Try wl-copy for Wayland
		cmd = exec.Command("wl-copy", "--type", "image/png")
		cmd.Stdin = bytes.NewReader(buf.Bytes())
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to copy to clipboard (install xclip or wl-clipboard): %w", err)
		}
	}

	return nil
}

// CopyText copies text to the system clipboard
func CopyText(text string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard", "-Value", text)
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = bytes.NewBufferString(text)
		return cmd.Run()
	case "linux":
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("wl-copy")
			cmd.Stdin = bytes.NewBufferString(text)
			return cmd.Run()
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
