//go:build windows

package platform

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	procGetDesktopWindow = user32.NewProc("GetDesktopWindow")
	procGetDC            = user32.NewProc("GetDC")
	procReleaseDC        = user32.NewProc("ReleaseDC")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procGetDIBits              = gdi32.NewProc("GetDIBits")

	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW     = user32.NewProc("GetMonitorInfoW")
)

const (
	smCxScreen = 0
	smCyScreen = 1
	srccopy    = 0x00CC0020
	biRgb      = 0
)

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type bitmapInfo struct {
	bmiHeader bitmapInfoHeader
}

type rect struct {
	Left, Top, Right, Bottom int32
}

type monitorInfoEx struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
}

// WindowsCapture implements CaptureBackend using Win32 GDI.
type WindowsCapture struct {
	monitors []image.Rectangle
}

// NewCapture creates a new WindowsCapture backend.
func NewCapture() *WindowsCapture {
	return &WindowsCapture{}
}

func (c *WindowsCapture) CaptureScreen(monitor int) (*image.RGBA, error) {
	bounds, err := c.ScreenBounds(monitor)
	if err != nil {
		return nil, err
	}
	return c.CaptureRegion(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
}

func (c *WindowsCapture) CaptureRegion(x, y, w, h int) (*image.RGBA, error) {
	hwnd, _, _ := procGetDesktopWindow.Call()
	hdc, _, _ := procGetDC.Call(hwnd)
	if hdc == 0 {
		return nil, fmt.Errorf("tsikuri: GetDC failed")
	}
	defer procReleaseDC.Call(hwnd, hdc)

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return nil, fmt.Errorf("tsikuri: CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(memDC)

	bitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	if bitmap == 0 {
		return nil, fmt.Errorf("tsikuri: CreateCompatibleBitmap failed")
	}
	defer procDeleteObject.Call(bitmap)

	procSelectObject.Call(memDC, bitmap)

	ret, _, _ := procBitBlt.Call(memDC, 0, 0, uintptr(w), uintptr(h),
		hdc, uintptr(x), uintptr(y), srccopy)
	if ret == 0 {
		return nil, fmt.Errorf("tsikuri: BitBlt failed")
	}

	bi := bitmapInfo{
		bmiHeader: bitmapInfoHeader{
			biSize:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			biWidth:       int32(w),
			biHeight:      -int32(h), // top-down
			biPlanes:      1,
			biBitCount:    32,
			biCompression: biRgb,
		},
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	procGetDIBits.Call(memDC, bitmap, 0, uintptr(h),
		uintptr(unsafe.Pointer(&img.Pix[0])),
		uintptr(unsafe.Pointer(&bi)),
		0, // DIB_RGB_COLORS
	)

	// GDI returns BGRA, convert to RGBA
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+2] = img.Pix[i+2], img.Pix[i]
	}

	return img, nil
}

func (c *WindowsCapture) NumScreens() (int, error) {
	if err := c.enumMonitors(); err != nil {
		return 0, err
	}
	return len(c.monitors), nil
}

func (c *WindowsCapture) ScreenBounds(monitor int) (image.Rectangle, error) {
	if monitor <= 0 {
		// Primary monitor
		w, _, _ := procGetSystemMetrics.Call(uintptr(smCxScreen))
		h, _, _ := procGetSystemMetrics.Call(uintptr(smCyScreen))
		return image.Rect(0, 0, int(w), int(h)), nil
	}

	if err := c.enumMonitors(); err != nil {
		return image.Rectangle{}, err
	}
	if monitor > len(c.monitors) {
		return image.Rectangle{}, fmt.Errorf("tsikuri: monitor %d not found", monitor)
	}
	return c.monitors[monitor-1], nil
}

func (c *WindowsCapture) enumMonitors() error {
	if c.monitors != nil {
		return nil
	}

	c.monitors = nil
	cb := syscall.NewCallback(func(hMonitor uintptr, hdc uintptr, lprcMonitor *rect, lParam uintptr) uintptr {
		info := monitorInfoEx{cbSize: uint32(unsafe.Sizeof(monitorInfoEx{}))}
		ret, _, _ := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&info)))
		if ret != 0 {
			c.monitors = append(c.monitors, image.Rect(
				int(info.rcMonitor.Left),
				int(info.rcMonitor.Top),
				int(info.rcMonitor.Right),
				int(info.rcMonitor.Bottom),
			))
		}
		return 1 // continue enumeration
	})

	procEnumDisplayMonitors.Call(0, 0, cb, 0)

	if len(c.monitors) == 0 {
		// Fallback to primary screen
		w, _, _ := procGetSystemMetrics.Call(uintptr(smCxScreen))
		h, _, _ := procGetSystemMetrics.Call(uintptr(smCyScreen))
		c.monitors = append(c.monitors, image.Rect(0, 0, int(w), int(h)))
	}

	return nil
}
