package backend

import "image"

// CaptureBackend captures screenshots from the display.
type CaptureBackend interface {
	// CaptureScreen captures the full screen of the given monitor index.
	// Pass -1 or 0 for the primary monitor.
	CaptureScreen(monitor int) (*image.RGBA, error)

	// CaptureRegion captures a rectangular region of the screen.
	CaptureRegion(x, y, w, h int) (*image.RGBA, error)

	// NumScreens returns the number of connected monitors.
	NumScreens() (int, error)

	// ScreenBounds returns the bounds of the given monitor.
	ScreenBounds(monitor int) (image.Rectangle, error)
}
