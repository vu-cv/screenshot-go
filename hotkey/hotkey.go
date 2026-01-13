package hotkey

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey      = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey    = user32.NewProc("UnregisterHotKey")
	procGetMessage          = user32.NewProc("GetMessageW")
	procPostThreadMessage   = user32.NewProc("PostThreadMessageW")
)

const (
	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	WM_HOTKEY    = 0x0312
	WM_QUIT      = 0x0012
)

// MSG structure for Windows messages
type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// Handler is a function that handles a hotkey press
type Handler func()

// Hotkey represents a keyboard shortcut
type Hotkey struct {
	ID        int
	Modifiers uint32
	KeyCode   uint32
	Handler   Handler
}

// Manager manages global hotkeys
type Manager struct {
	hotkeys  map[int]*Hotkey
	nextID   int
	running  bool
	stopCh   chan struct{}
	threadID uint32
	mu       sync.RWMutex
}

// NewManager creates a new hotkey manager
func NewManager() *Manager {
	return &Manager{
		hotkeys: make(map[int]*Hotkey),
		nextID:  1,
		stopCh:  make(chan struct{}),
	}
}

// keyNameToVK converts key name to virtual key code
func keyNameToVK(key string) uint32 {
	keyMap := map[string]uint32{
		"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45,
		"f": 0x46, "g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A,
		"k": 0x4B, "l": 0x4C, "m": 0x4D, "n": 0x4E, "o": 0x4F,
		"p": 0x50, "q": 0x51, "r": 0x52, "s": 0x53, "t": 0x54,
		"u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58, "y": 0x59,
		"z": 0x5A,
		"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
		"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
		"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73,
		"f5": 0x74, "f6": 0x75, "f7": 0x76, "f8": 0x77,
		"f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
		"printscreen": 0x2C, "escape": 0x1B, "space": 0x20,
		"enter": 0x0D, "backspace": 0x08, "tab": 0x09,
		"insert": 0x2D, "delete": 0x2E, "home": 0x24, "end": 0x23,
		"pageup": 0x21, "pagedown": 0x22,
		"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
	}
	if vk, ok := keyMap[strings.ToLower(key)]; ok {
		return vk
	}
	return 0
}

// ParseHotkey parses a hotkey string like "ctrl+shift+s"
func ParseHotkey(hotkeyStr string) (*Hotkey, error) {
	parts := strings.Split(strings.ToLower(hotkeyStr), "+")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid hotkey string: %s", hotkeyStr)
	}

	hk := &Hotkey{}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if i == len(parts)-1 {
			// Last part is the main key
			vk := keyNameToVK(part)
			if vk == 0 {
				return nil, fmt.Errorf("unknown key: %s", part)
			}
			hk.KeyCode = vk
		} else {
			// Other parts are modifiers
			switch part {
			case "ctrl", "control":
				hk.Modifiers |= MOD_CONTROL
			case "alt":
				hk.Modifiers |= MOD_ALT
			case "shift":
				hk.Modifiers |= MOD_SHIFT
			case "win", "cmd", "super":
				hk.Modifiers |= MOD_WIN
			default:
				return nil, fmt.Errorf("unknown modifier: %s", part)
			}
		}
	}

	return hk, nil
}

// Register registers a hotkey with a handler
func (m *Manager) Register(hotkeyStr string, handler Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hk, err := ParseHotkey(hotkeyStr)
	if err != nil {
		return err
	}
	hk.Handler = handler
	hk.ID = m.nextID
	m.nextID++

	m.hotkeys[hk.ID] = hk
	return nil
}

// Unregister removes a hotkey by ID
func (m *Manager) Unregister(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.hotkeys, id)
}

// registerHotkeys registers all hotkeys with Windows
func (m *Manager) registerHotkeys() error {
	for _, hk := range m.hotkeys {
		ret, _, err := procRegisterHotKey.Call(
			0,
			uintptr(hk.ID),
			uintptr(hk.Modifiers),
			uintptr(hk.KeyCode),
		)
		if ret == 0 {
			return fmt.Errorf("failed to register hotkey ID %d: %v", hk.ID, err)
		}
	}
	return nil
}

// unregisterHotkeys unregisters all hotkeys from Windows
func (m *Manager) unregisterHotkeys() {
	for _, hk := range m.hotkeys {
		procUnregisterHotKey.Call(0, uintptr(hk.ID))
	}
}

// Start starts listening for hotkeys
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("hotkey manager is already running")
	}
	m.running = true
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	// Register hotkeys
	if err := m.registerHotkeys(); err != nil {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
		return err
	}

	// Get current thread ID for message loop
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getCurrentThreadId := kernel32.NewProc("GetCurrentThreadId")

	go func() {
		// Store thread ID for stopping
		tid, _, _ := getCurrentThreadId.Call()
		m.mu.Lock()
		m.threadID = uint32(tid)
		m.mu.Unlock()

		var msg MSG
		for {
			select {
			case <-m.stopCh:
				m.unregisterHotkeys()
				return
			default:
				ret, _, _ := procGetMessage.Call(
					uintptr(unsafe.Pointer(&msg)),
					0, 0, 0,
				)
				if ret == 0 || int32(ret) == -1 {
					return
				}

				if msg.Message == WM_HOTKEY {
					m.mu.RLock()
					if hk, ok := m.hotkeys[int(msg.WParam)]; ok {
						go hk.Handler()
					}
					m.mu.RUnlock()
				}
			}
		}
	}()

	return nil
}

// Stop stops the hotkey listener
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	close(m.stopCh)
	m.running = false

	// Post quit message to break GetMessage loop
	if m.threadID != 0 {
		procPostThreadMessage.Call(
			uintptr(m.threadID),
			WM_QUIT,
			0, 0,
		)
	}
}

// IsRunning returns whether the manager is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
