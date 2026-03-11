package tsikuri

// Key represents a keyboard key or modifier.
type Key string

// Standard keys.
const (
	KeyEnter       Key = "enter"
	KeyReturn      Key = "return"
	KeyTab         Key = "tab"
	KeyEscape      Key = "escape"
	KeyBackspace   Key = "backspace"
	KeyDelete      Key = "delete"
	KeyUp          Key = "up"
	KeyDown        Key = "down"
	KeyLeft        Key = "left"
	KeyRight       Key = "right"
	KeyHome        Key = "home"
	KeyEnd         Key = "end"
	KeyPageUp      Key = "pageup"
	KeyPageDown    Key = "pagedown"
	KeySpace       Key = "space"
	KeyInsert      Key = "insert"
	KeyPrintScreen Key = "printscreen"
	KeyF1          Key = "f1"
	KeyF2          Key = "f2"
	KeyF3          Key = "f3"
	KeyF4          Key = "f4"
	KeyF5          Key = "f5"
	KeyF6          Key = "f6"
	KeyF7          Key = "f7"
	KeyF8          Key = "f8"
	KeyF9          Key = "f9"
	KeyF10         Key = "f10"
	KeyF11         Key = "f11"
	KeyF12         Key = "f12"
	KeyF13         Key = "f13"
	KeyF14         Key = "f14"
	KeyF15         Key = "f15"
)

// Modifier keys.
const (
	ModCtrl  Key = "ctrl"
	ModAlt   Key = "alt"
	ModShift Key = "shift"
	ModMeta  Key = "meta"
)

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
