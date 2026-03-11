//go:build linux

package platform

import (
	"fmt"
	"image"
	"os/exec"
	"strconv"
	"strings"
)

// LinuxCapture implements CaptureBackend using X11 tools.
// It shells out to xdpyinfo and import (ImageMagick) or xwd for capture,
// keeping the implementation pure Go with no CGO.
type LinuxCapture struct {
	screenWidth  int
	screenHeight int
	initialized  bool
}

// NewCapture creates a new LinuxCapture backend.
func NewCapture() *LinuxCapture {
	return &LinuxCapture{}
}

func (c *LinuxCapture) init() error {
	if c.initialized {
		return nil
	}

	// Get screen dimensions from xdpyinfo
	out, err := exec.Command("xdpyinfo").Output()
	if err != nil {
		// Fallback defaults
		c.screenWidth = 1920
		c.screenHeight = 1080
		c.initialized = true
		return nil
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "dimensions:") {
			// Format: "dimensions:    1920x1080 pixels (508x285 millimeters)"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dims := strings.Split(parts[1], "x")
				if len(dims) == 2 {
					c.screenWidth, _ = strconv.Atoi(dims[0])
					c.screenHeight, _ = strconv.Atoi(dims[1])
				}
			}
			break
		}
	}

	if c.screenWidth == 0 || c.screenHeight == 0 {
		c.screenWidth = 1920
		c.screenHeight = 1080
	}

	c.initialized = true
	return nil
}

func (c *LinuxCapture) CaptureScreen(monitor int) (*image.RGBA, error) {
	bounds, err := c.ScreenBounds(monitor)
	if err != nil {
		return nil, err
	}
	return c.CaptureRegion(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
}

func (c *LinuxCapture) CaptureRegion(x, y, w, h int) (*image.RGBA, error) {
	// Use xwd + convert pipeline, or import directly
	// Try import (ImageMagick) first
	geometry := fmt.Sprintf("%dx%d+%d+%d", w, h, x, y)
	cmd := exec.Command("import", "-window", "root", "-crop", geometry, "ppm:-")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tsikuri: screen capture failed (is ImageMagick installed?): %w", err)
	}

	return parsePPM(out, w, h)
}

func (c *LinuxCapture) NumScreens() (int, error) {
	// For X11, we report 1 screen unless xrandr says otherwise
	out, err := exec.Command("xrandr", "--listmonitors").Output()
	if err != nil {
		return 1, nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 1 {
		// First line is "Monitors: N"
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			n, err := strconv.Atoi(parts[1])
			if err == nil {
				return n, nil
			}
		}
	}
	return 1, nil
}

func (c *LinuxCapture) ScreenBounds(monitor int) (image.Rectangle, error) {
	if err := c.init(); err != nil {
		return image.Rectangle{}, err
	}
	// For the primary monitor, use stored dimensions
	return image.Rect(0, 0, c.screenWidth, c.screenHeight), nil
}

// parsePPM parses a binary PPM (P6) image from raw bytes.
func parsePPM(data []byte, expectedW, expectedH int) (*image.RGBA, error) {
	// Simple PPM P6 parser
	idx := 0
	// Read magic number
	if len(data) < 2 || string(data[0:2]) != "P6" {
		return nil, fmt.Errorf("tsikuri: not a PPM P6 image")
	}
	idx = 2

	// Skip whitespace and comments
	skipWS := func() {
		for idx < len(data) {
			if data[idx] == '#' {
				for idx < len(data) && data[idx] != '\n' {
					idx++
				}
			} else if data[idx] == ' ' || data[idx] == '\n' || data[idx] == '\r' || data[idx] == '\t' {
				idx++
			} else {
				break
			}
		}
	}

	readInt := func() int {
		skipWS()
		n := 0
		for idx < len(data) && data[idx] >= '0' && data[idx] <= '9' {
			n = n*10 + int(data[idx]-'0')
			idx++
		}
		return n
	}

	w := readInt()
	h := readInt()
	_ = readInt() // maxval

	// Skip single whitespace after maxval
	if idx < len(data) {
		idx++
	}

	if w == 0 || h == 0 {
		w = expectedW
		h = expectedH
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if idx+2 >= len(data) {
				break
			}
			r, g, b := data[idx], data[idx+1], data[idx+2]
			idx += 3
			off := (y*w + x) * 4
			if off+3 < len(img.Pix) {
				img.Pix[off] = r
				img.Pix[off+1] = g
				img.Pix[off+2] = b
				img.Pix[off+3] = 255
			}
		}
	}

	return img, nil
}
