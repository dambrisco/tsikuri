package tsikuri

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dambrisco/tsikuri/internal/match"
	"github.com/dambrisco/tsikuri/tsikuritest"
)

func TestObserverOnAppear(t *testing.T) {
	// Start with a blank screen
	blankScreen := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			blankScreen.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	mockCapture := tsikuritest.NewMockCapture(blankScreen)
	backends := &Backends{
		Capture: mockCapture,
		Match:   match.New(),
	}

	settings := DefaultSettings()
	settings.ScanRate = 10 * time.Millisecond
	r := NewRegion(image.Rect(0, 0, 100, 100), backends, settings)

	// Create needle with a gradient
	tmpDir := t.TempDir()
	needlePath := filepath.Join(tmpDir, "target.png")
	needleImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			v := uint8((x + y) * 25)
			needleImg.Set(x, y, color.RGBA{R: v, G: 255 - v, B: v / 2, A: 255})
		}
	}
	f, _ := os.Create(needlePath)
	png.Encode(f, needleImg)
	f.Close()

	var appeared int32
	obs := r.OnAppear(Image("target.png").WithBasePath(tmpDir).WithSimilarity(0.9), func(e ObserveEvent) {
		atomic.AddInt32(&appeared, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go obs.Run(ctx)

	// After a short delay, make the target appear
	time.Sleep(50 * time.Millisecond)
	screenWithTarget := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			screenWithTarget.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			v := uint8((x + y) * 25)
			screenWithTarget.Set(50+x, 50+y, color.RGBA{R: v, G: 255 - v, B: v / 2, A: 255})
		}
	}
	mockCapture.SetImage(screenWithTarget)

	// Wait for observer to detect it
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&appeared) == 0 {
		t.Error("expected OnAppear handler to fire")
	}
}
