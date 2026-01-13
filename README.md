# Screenshot Tool

A full-featured screenshot application written in Go for Windows.

## Features

- 📸 **Multiple Capture Modes**
  - Full screen capture
  - Specific display capture (multi-monitor support)
  - Custom region capture
  - Mouse drag selection

- ✏️ **Built-in Editor**
  - Add text with background
  - Draw rectangles
  - Draw arrows
  - Highlight areas
  - Free draw
  - Custom dark-themed UI

- ⌨️ **Global Hotkeys** (Daemon mode)
  - `Ctrl+Shift+S` - Fullscreen screenshot
  - `Ctrl+Shift+R` - Region selection with editor

- 📋 **Clipboard Integration**
  - Auto-copy screenshots to clipboard

- ⚙️ **Configuration**
  - JSON config file
  - Customizable output directory
  - Multiple formats (PNG, JPEG, GIF)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/screenshot.git
cd screenshot

# Build
go build -o screenshot.exe .
```

## Usage

### Basic Commands

```bash
# Full screen capture
.\screenshot.exe -full -output .

# Capture specific display (0-indexed)
.\screenshot.exe -display 0 -output .

# Capture specific region
.\screenshot.exe -region 100,100,800,600 -output .

# Mouse drag region selection
.\screenshot.exe -select -output .

# Capture with editor
.\screenshot.exe -select -edit -output .

# Full screen with editor
.\screenshot.exe -full -edit -output .
```

### Options

| Flag | Description |
|------|-------------|
| `-full` | Capture entire screen (all displays) |
| `-display N` | Capture specific display (0-indexed) |
| `-region X,Y,W,H` | Capture specific region |
| `-select` | Mouse drag to select region |
| `-edit` | Open editor after capture |
| `-output PATH` | Output directory (default: current) |
| `-format FORMAT` | Image format: png, jpg, gif (default: png) |
| `-quality N` | JPEG quality 1-100 (default: 90) |
| `-delay N` | Delay in seconds before capture |
| `-copy` | Copy to clipboard |
| `-daemon` | Run in background with hotkeys |
| `-list-displays` | List available displays |

### Daemon Mode

Run in background with global hotkeys:

```bash
.\screenshot.exe -daemon
```

Hotkeys:
- `Ctrl+Shift+S` - Capture fullscreen
- `Ctrl+Shift+R` - Select region + edit

### Editor Controls

| Tool | Description |
|------|-------------|
| ✏ Text | Click to place text on image |
| ▢ Rect | Drag to draw rectangle |
| ➔ Arrow | Drag to draw arrow |
| ▨ Mark | Drag to highlight area |
| ✎ Draw | Free draw line |
| ↺ Undo | Reset all changes |

- **Color Palette**: 10 preset colors
- **Text Input**: Type text, then click on image to place
- **Save**: Confirm changes
- **Cancel** / `ESC`: Discard changes

## Configuration

Config file: `screenshot_config.json`

```json
{
  "output_dir": ".",
  "format": "png",
  "quality": 90,
  "copy_to_clipboard": true,
  "hotkeys": {
    "fullscreen": "ctrl+shift+s",
    "region": "ctrl+shift+r"
  }
}
```

## Dependencies

- [github.com/kbinani/screenshot](https://github.com/kbinani/screenshot) - Screen capture
- [golang.org/x/image](https://golang.org/x/image) - Font rendering

## Project Structure

```
screenshot/
├── main.go              # CLI entry point
├── capture/
│   └── capture.go       # Screen capture logic
├── config/
│   └── config.go        # Configuration management
├── clipboard/
│   └── clipboard.go     # Clipboard operations
├── hotkey/
│   └── hotkey.go        # Global hotkey registration
├── selector/
│   └── selector_windows.go  # Mouse region selection
├── editor/
│   ├── editor.go        # Image manipulation
│   └── editor_windows.go    # GUI editor (Windows)
├── go.mod
├── go.sum
└── README.md
```

## Requirements

- Windows 10/11
- Go 1.21+

## License

MIT License
