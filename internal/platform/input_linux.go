//go:build linux

package platform

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dambrisco/tsikuri/backend"
)

// LinuxInput implements InputBackend using xdotool.
type LinuxInput struct{}

// NewInput creates a new LinuxInput backend.
func NewInput() *LinuxInput {
	return &LinuxInput{}
}

func (l *LinuxInput) MouseMove(x, y int) error {
	return exec.Command("xdotool", "mousemove", strconv.Itoa(x), strconv.Itoa(y)).Run()
}

func (l *LinuxInput) MouseClick(x, y int, button backend.MouseButton) error {
	if err := l.MouseMove(x, y); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	btn := linuxButton(button)
	return exec.Command("xdotool", "click", btn).Run()
}

func (l *LinuxInput) MouseDoubleClick(x, y int, button backend.MouseButton) error {
	if err := l.MouseMove(x, y); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	btn := linuxButton(button)
	return exec.Command("xdotool", "click", "--repeat", "2", "--delay", "50", btn).Run()
}

func (l *LinuxInput) MouseDown(button backend.MouseButton) error {
	btn := linuxButton(button)
	return exec.Command("xdotool", "mousedown", btn).Run()
}

func (l *LinuxInput) MouseUp(button backend.MouseButton) error {
	btn := linuxButton(button)
	return exec.Command("xdotool", "mouseup", btn).Run()
}

func (l *LinuxInput) Scroll(x, y int, direction backend.ScrollDirection, amount int) error {
	if err := l.MouseMove(x, y); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)

	var btn string
	switch direction {
	case backend.ScrollUp:
		btn = "4"
	case backend.ScrollDown:
		btn = "5"
	case backend.ScrollLeft:
		btn = "6"
	case backend.ScrollRight:
		btn = "7"
	}

	return exec.Command("xdotool", "click", "--repeat", strconv.Itoa(amount), btn).Run()
}

func (l *LinuxInput) DragDrop(fromX, fromY, toX, toY int) error {
	if err := l.MouseMove(fromX, fromY); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := l.MouseDown(backend.ButtonLeft); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := l.MouseMove(toX, toY); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return l.MouseUp(backend.ButtonLeft)
}

func (l *LinuxInput) TypeText(text string) error {
	return exec.Command("xdotool", "type", "--clearmodifiers", text).Run()
}

func (l *LinuxInput) KeyDown(key string) error {
	xkey := keyToXdotool(key)
	return exec.Command("xdotool", "keydown", xkey).Run()
}

func (l *LinuxInput) KeyUp(key string) error {
	xkey := keyToXdotool(key)
	return exec.Command("xdotool", "keyup", xkey).Run()
}

func (l *LinuxInput) KeyPress(key string, modifiers ...string) error {
	combo := ""
	for _, mod := range modifiers {
		combo += keyToXdotool(mod) + "+"
	}
	combo += keyToXdotool(key)
	return exec.Command("xdotool", "key", combo).Run()
}

func (l *LinuxInput) SetClipboard(text string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (l *LinuxInput) PasteClipboard() error {
	return l.KeyPress("v", "ctrl")
}

func linuxButton(button backend.MouseButton) string {
	switch button {
	case backend.ButtonRight:
		return "3"
	case backend.ButtonMiddle:
		return "2"
	default:
		return "1"
	}
}

func keyToXdotool(key string) string {
	switch key {
	case "enter", "return":
		return "Return"
	case "tab":
		return "Tab"
	case "escape":
		return "Escape"
	case "backspace":
		return "BackSpace"
	case "delete":
		return "Delete"
	case "up":
		return "Up"
	case "down":
		return "Down"
	case "left":
		return "Left"
	case "right":
		return "Right"
	case "home":
		return "Home"
	case "end":
		return "End"
	case "pageup":
		return "Prior"
	case "pagedown":
		return "Next"
	case "space":
		return "space"
	case "insert":
		return "Insert"
	case "printscreen":
		return "Print"
	case "ctrl":
		return "ctrl"
	case "alt":
		return "alt"
	case "shift":
		return "shift"
	case "meta":
		return "super"
	default:
		return key
	}
}
