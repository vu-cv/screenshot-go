package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds application configuration
type Config struct {
	OutputDir      string        `json:"output_dir"`
	Format         string        `json:"format"`          // "png" or "jpg"
	Quality        int           `json:"quality"`         // JPEG quality 1-100
	DefaultDelay   time.Duration `json:"default_delay"`   // Default delay before capture
	CopyToClipboard bool         `json:"copy_to_clipboard"`
	PlaySound      bool          `json:"play_sound"`
	ShowNotification bool        `json:"show_notification"`
	
	// Hotkeys
	HotkeyFullScreen   string `json:"hotkey_fullscreen"`    // e.g., "ctrl+shift+f"
	HotkeyRegion       string `json:"hotkey_region"`        // e.g., "ctrl+shift+r"
	HotkeyActiveWindow string `json:"hotkey_active_window"` // e.g., "ctrl+shift+w"
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	picturesDir := filepath.Join(homeDir, "Pictures", "Screenshots")

	return &Config{
		OutputDir:         picturesDir,
		Format:            "png",
		Quality:           90,
		DefaultDelay:      0,
		CopyToClipboard:   true,
		PlaySound:         true,
		ShowNotification:  true,
		HotkeyFullScreen:  "ctrl+shift+s",
		HotkeyRegion:      "ctrl+shift+r",
		HotkeyActiveWindow: "ctrl+shift+w",
	}
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "screenshot-go", "config.json")
}

// Load loads configuration from file
func Load() (*Config, error) {
	configPath := GetConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()
	
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// ToCaptureOptions converts config to capture options
func (c *Config) ToCaptureOptions() map[string]interface{} {
	return map[string]interface{}{
		"output_dir": c.OutputDir,
		"format":     c.Format,
		"quality":    c.Quality,
		"delay":      c.DefaultDelay,
	}
}
