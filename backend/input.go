package backend

// InputBackend simulates mouse and keyboard input.
type InputBackend interface {
	// MouseMove moves the mouse cursor to the given screen coordinates.
	MouseMove(x, y int) error

	// MouseClick performs a click at the given coordinates with the given button.
	MouseClick(x, y int, button MouseButton) error

	// MouseDoubleClick performs a double-click at the given coordinates.
	MouseDoubleClick(x, y int, button MouseButton) error

	// MouseDown presses the given mouse button without releasing.
	MouseDown(button MouseButton) error

	// MouseUp releases the given mouse button.
	MouseUp(button MouseButton) error

	// Scroll scrolls at the given coordinates in the given direction.
	Scroll(x, y int, direction ScrollDirection, amount int) error

	// DragDrop performs a drag from one point to another.
	DragDrop(fromX, fromY, toX, toY int) error

	// TypeText types the given text string character by character.
	TypeText(text string) error

	// KeyDown presses a key without releasing it.
	KeyDown(key string) error

	// KeyUp releases a previously pressed key.
	KeyUp(key string) error

	// KeyPress presses and releases a key, optionally with modifiers held.
	KeyPress(key string, modifiers ...string) error

	// SetClipboard sets the system clipboard text.
	SetClipboard(text string) error

	// PasteClipboard triggers a paste action (Ctrl+V / Cmd+V).
	PasteClipboard() error
}

// MouseButton represents a mouse button.
type MouseButton int

const (
	ButtonLeft MouseButton = iota
	ButtonRight
	ButtonMiddle
)

// ScrollDirection represents a scroll direction.
type ScrollDirection int

const (
	ScrollUp ScrollDirection = iota
	ScrollDown
	ScrollLeft
	ScrollRight
)
