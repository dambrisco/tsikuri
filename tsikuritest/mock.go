package tsikuritest

import (
	"image"
	"sync"
)

// MockCapture records capture calls and returns pre-configured images.
type MockCapture struct {
	mu          sync.Mutex
	screenImage *image.RGBA
	bounds      image.Rectangle
}

// NewMockCapture creates a MockCapture that returns the given image.
func NewMockCapture(img *image.RGBA) *MockCapture {
	return &MockCapture{
		screenImage: img,
		bounds:      img.Bounds(),
	}
}

func (m *MockCapture) CaptureScreen(monitor int) (*image.RGBA, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.screenImage, nil
}

func (m *MockCapture) CaptureRegion(x, y, w, h int) (*image.RGBA, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub := image.NewRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			sx, sy := x+dx, y+dy
			if sx >= 0 && sy >= 0 && sx < m.bounds.Dx() && sy < m.bounds.Dy() {
				sub.Set(dx, dy, m.screenImage.At(sx, sy))
			}
		}
	}
	return sub, nil
}

func (m *MockCapture) NumScreens() (int, error) { return 1, nil }

func (m *MockCapture) ScreenBounds(monitor int) (image.Rectangle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bounds, nil
}

// SetImage updates the screen image returned by capture calls.
func (m *MockCapture) SetImage(img *image.RGBA) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.screenImage = img
	m.bounds = img.Bounds()
}

// MockInput records all input operations for later assertions.
type MockInput struct {
	mu     sync.Mutex
	Clicks []MockClick
	Keys   []MockKey
	Typed  []string
}

// MockClick records a click event.
type MockClick struct {
	X, Y   int
	Button int
	Double bool
}

// MockKey records a key event.
type MockKey struct {
	Key       string
	Modifiers []string
	Down      bool
	Up        bool
}

// NewMockInput creates a new MockInput.
func NewMockInput() *MockInput {
	return &MockInput{}
}

func (m *MockInput) MouseMove(x, y int) error { return nil }

func (m *MockInput) MouseClick(x, y int, button int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Clicks = append(m.Clicks, MockClick{X: x, Y: y, Button: button})
	return nil
}

func (m *MockInput) MouseDoubleClick(x, y int, button int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Clicks = append(m.Clicks, MockClick{X: x, Y: y, Button: button, Double: true})
	return nil
}

func (m *MockInput) MouseDown(button int) error                       { return nil }
func (m *MockInput) MouseUp(button int) error                         { return nil }
func (m *MockInput) Scroll(x, y int, direction int, amount int) error { return nil }
func (m *MockInput) DragDrop(fromX, fromY, toX, toY int) error       { return nil }

func (m *MockInput) TypeText(text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Typed = append(m.Typed, text)
	return nil
}

func (m *MockInput) KeyDown(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = append(m.Keys, MockKey{Key: key, Down: true})
	return nil
}

func (m *MockInput) KeyUp(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = append(m.Keys, MockKey{Key: key, Up: true})
	return nil
}

func (m *MockInput) KeyPress(key string, modifiers ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Keys = append(m.Keys, MockKey{Key: key, Modifiers: modifiers})
	return nil
}

func (m *MockInput) SetClipboard(text string) error { return nil }
func (m *MockInput) PasteClipboard() error          { return nil }

// AssertClicked checks that a click occurred at the given coordinates.
func (m *MockInput) AssertClicked(x, y int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.Clicks {
		if c.X == x && c.Y == y {
			return true
		}
	}
	return false
}

// MockOCR returns pre-configured text.
type MockOCR struct {
	Text string
	Err  error
}

func (m *MockOCR) ReadText(img image.Image) (string, error) {
	return m.Text, m.Err
}

func (m *MockOCR) Close() error { return nil }
