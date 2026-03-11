//go:build windows

package platform

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/dambrisco/tsikuri/backend"
)

var (
	procSendInput       = user32.NewProc("SendInput")
	procSetCursorPos    = user32.NewProc("SetCursorPos")
	procGetCursorPos    = user32.NewProc("GetCursorPos")
	procOpenClipboard   = user32.NewProc("OpenClipboard")
	procCloseClipboard  = user32.NewProc("CloseClipboard")
	procEmptyClipboard  = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGlobalAlloc     = kernel32.NewProc("GlobalAlloc")
	procGlobalLock      = kernel32.NewProc("GlobalLock")
	procGlobalUnlock    = kernel32.NewProc("GlobalUnlock")
)

const (
	inputMouse    = 0
	inputKeyboard = 1

	mousefMove       = 0x0001
	mousefLeftDown   = 0x0002
	mousefLeftUp     = 0x0004
	mousefRightDown  = 0x0008
	mousefRightUp    = 0x0010
	mousefMiddleDown = 0x0020
	mousefMiddleUp   = 0x0040
	mousefWheel      = 0x0800
	mousefHWheel     = 0x1000
	mousefAbsolute   = 0x8000

	keybdfKeyUp  = 0x0002
	keybdfUnicode = 0x0004

	wheelDelta = 120

	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

type mouseInput struct {
	typ uint32
	mi  struct {
		dx, dy    int32
		mouseData uint32
		dwFlags   uint32
		time      uint32
		dwExtraInfo uintptr
	}
}

type keybdInput struct {
	typ uint32
	ki  struct {
		wVk         uint16
		wScan       uint16
		dwFlags     uint32
		time        uint32
		dwExtraInfo uintptr
	}
}

// WindowsInput implements InputBackend using Win32 SendInput.
type WindowsInput struct{}

// NewInput creates a new WindowsInput backend.
func NewInput() *WindowsInput {
	return &WindowsInput{}
}

func (w *WindowsInput) MouseMove(x, y int) error {
	ret, _, _ := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("tsikuri: SetCursorPos failed")
	}
	return nil
}

func (w *WindowsInput) MouseClick(x, y int, button backend.MouseButton) error {
	if err := w.MouseMove(x, y); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	down, up := buttonFlags(button)
	input := mouseInput{typ: inputMouse}
	input.mi.dwFlags = down

	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))

	time.Sleep(10 * time.Millisecond)

	input.mi.dwFlags = up
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))

	return nil
}

func (w *WindowsInput) MouseDoubleClick(x, y int, button backend.MouseButton) error {
	if err := w.MouseClick(x, y, button); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return w.MouseClick(x, y, button)
}

func (w *WindowsInput) MouseDown(button backend.MouseButton) error {
	down, _ := buttonFlags(button)
	input := mouseInput{typ: inputMouse}
	input.mi.dwFlags = down
	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))
	return nil
}

func (w *WindowsInput) MouseUp(button backend.MouseButton) error {
	_, up := buttonFlags(button)
	input := mouseInput{typ: inputMouse}
	input.mi.dwFlags = up
	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))
	return nil
}

func (w *WindowsInput) Scroll(x, y int, direction backend.ScrollDirection, amount int) error {
	if err := w.MouseMove(x, y); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	input := mouseInput{typ: inputMouse}
	switch direction {
	case backend.ScrollUp:
		input.mi.dwFlags = mousefWheel
		input.mi.mouseData = uint32(amount * wheelDelta)
	case backend.ScrollDown:
		input.mi.dwFlags = mousefWheel
		input.mi.mouseData = uint32(-amount * wheelDelta)
	case backend.ScrollLeft:
		input.mi.dwFlags = mousefHWheel
		input.mi.mouseData = uint32(-amount * wheelDelta)
	case backend.ScrollRight:
		input.mi.dwFlags = mousefHWheel
		input.mi.mouseData = uint32(amount * wheelDelta)
	}

	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))
	return nil
}

func (w *WindowsInput) DragDrop(fromX, fromY, toX, toY int) error {
	if err := w.MouseMove(fromX, fromY); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	if err := w.MouseDown(backend.ButtonLeft); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	// Move in steps for smooth dragging
	steps := 20
	for i := 1; i <= steps; i++ {
		x := fromX + (toX-fromX)*i/steps
		y := fromY + (toY-fromY)*i/steps
		if err := w.MouseMove(x, y); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	return w.MouseUp(backend.ButtonLeft)
}

func (w *WindowsInput) TypeText(text string) error {
	for _, ch := range text {
		encoded := utf16.Encode([]rune{ch})
		for _, u := range encoded {
			// Key down
			input := keybdInput{typ: inputKeyboard}
			input.ki.wScan = u
			input.ki.dwFlags = keybdfUnicode
			size := unsafe.Sizeof(input)
			procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))

			// Key up
			input.ki.dwFlags = keybdfUnicode | keybdfKeyUp
			procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))

			time.Sleep(5 * time.Millisecond)
		}
	}
	return nil
}

func (w *WindowsInput) KeyDown(key string) error {
	vk := keyToVK(key)
	if vk == 0 {
		return fmt.Errorf("tsikuri: unknown key %q", key)
	}
	input := keybdInput{typ: inputKeyboard}
	input.ki.wVk = vk
	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))
	return nil
}

func (w *WindowsInput) KeyUp(key string) error {
	vk := keyToVK(key)
	if vk == 0 {
		return fmt.Errorf("tsikuri: unknown key %q", key)
	}
	input := keybdInput{typ: inputKeyboard}
	input.ki.wVk = vk
	input.ki.dwFlags = keybdfKeyUp
	size := unsafe.Sizeof(input)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(size))
	return nil
}

func (w *WindowsInput) KeyPress(key string, modifiers ...string) error {
	for _, mod := range modifiers {
		if err := w.KeyDown(mod); err != nil {
			return err
		}
	}

	if err := w.KeyDown(key); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := w.KeyUp(key); err != nil {
		return err
	}

	for i := len(modifiers) - 1; i >= 0; i-- {
		if err := w.KeyUp(modifiers[i]); err != nil {
			return err
		}
	}
	return nil
}

func (w *WindowsInput) SetClipboard(text string) error {
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("tsikuri: OpenClipboard failed")
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	// Convert to UTF-16 with null terminator
	utf16Text := utf16.Encode([]rune(text + "\x00"))
	size := len(utf16Text) * 2

	hMem, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(size))
	if hMem == 0 {
		return fmt.Errorf("tsikuri: GlobalAlloc failed")
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return fmt.Errorf("tsikuri: GlobalLock failed")
	}

	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(utf16Text))
	copy(dst, utf16Text)

	procGlobalUnlock.Call(hMem)
	procSetClipboardData.Call(cfUnicodeText, hMem)

	return nil
}

func (w *WindowsInput) PasteClipboard() error {
	return w.KeyPress("v", "ctrl")
}

func buttonFlags(button backend.MouseButton) (down, up uint32) {
	switch button {
	case backend.ButtonRight:
		return mousefRightDown, mousefRightUp
	case backend.ButtonMiddle:
		return mousefMiddleDown, mousefMiddleUp
	default:
		return mousefLeftDown, mousefLeftUp
	}
}

func keyToVK(key string) uint16 {
	switch strings.ToLower(key) {
	case "enter", "return":
		return 0x0D
	case "tab":
		return 0x09
	case "escape", "esc":
		return 0x1B
	case "backspace":
		return 0x08
	case "delete":
		return 0x2E
	case "up":
		return 0x26
	case "down":
		return 0x28
	case "left":
		return 0x25
	case "right":
		return 0x27
	case "home":
		return 0x24
	case "end":
		return 0x23
	case "pageup":
		return 0x21
	case "pagedown":
		return 0x22
	case "space":
		return 0x20
	case "insert":
		return 0x2D
	case "printscreen":
		return 0x2C
	case "ctrl":
		return 0x11
	case "alt":
		return 0x12
	case "shift":
		return 0x10
	case "meta":
		return 0x5B // Left Windows key
	case "f1":
		return 0x70
	case "f2":
		return 0x71
	case "f3":
		return 0x72
	case "f4":
		return 0x73
	case "f5":
		return 0x74
	case "f6":
		return 0x75
	case "f7":
		return 0x76
	case "f8":
		return 0x77
	case "f9":
		return 0x78
	case "f10":
		return 0x79
	case "f11":
		return 0x7A
	case "f12":
		return 0x7B
	case "a":
		return 0x41
	case "b":
		return 0x42
	case "c":
		return 0x43
	case "d":
		return 0x44
	case "e":
		return 0x45
	case "f":
		return 0x46
	case "g":
		return 0x47
	case "h":
		return 0x48
	case "i":
		return 0x49
	case "j":
		return 0x4A
	case "k":
		return 0x4B
	case "l":
		return 0x4C
	case "m":
		return 0x4D
	case "n":
		return 0x4E
	case "o":
		return 0x4F
	case "p":
		return 0x50
	case "q":
		return 0x51
	case "r":
		return 0x52
	case "s":
		return 0x53
	case "t":
		return 0x54
	case "u":
		return 0x55
	case "v":
		return 0x56
	case "w":
		return 0x57
	case "x":
		return 0x58
	case "y":
		return 0x59
	case "z":
		return 0x5A
	default:
		// Single character keys
		if len(key) == 1 {
			ch := key[0]
			if ch >= '0' && ch <= '9' {
				return uint16(ch)
			}
			if ch >= 'A' && ch <= 'Z' {
				return uint16(ch)
			}
		}
		return 0
	}
}
