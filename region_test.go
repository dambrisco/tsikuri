package tsikuri

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/dambrisco/tsikuri/internal/match"
	"github.com/dambrisco/tsikuri/tsikuritest"
)

func TestRegionCenter(t *testing.T) {
	r := NewRegion(image.Rect(10, 20, 110, 120), nil, DefaultSettings())
	c := r.center()
	if c.X != 60 || c.Y != 70 {
		t.Errorf("center: got (%d, %d), want (60, 70)", c.X, c.Y)
	}
}

func TestRegionManipulation(t *testing.T) {
	r := NewRegion(image.Rect(100, 100, 200, 200), nil, DefaultSettings())

	t.Run("Above", func(t *testing.T) {
		above := r.Above(50)
		if above.Bounds != image.Rect(100, 50, 200, 100) {
			t.Errorf("Above(50): got %v, want Rect(100,50,200,100)", above.Bounds)
		}
	})

	t.Run("Below", func(t *testing.T) {
		below := r.Below(50)
		if below.Bounds != image.Rect(100, 200, 200, 250) {
			t.Errorf("Below(50): got %v, want Rect(100,200,200,250)", below.Bounds)
		}
	})

	t.Run("Left", func(t *testing.T) {
		left := r.Left(30)
		if left.Bounds != image.Rect(100, 100, 130, 200) {
			t.Errorf("Left(30): got %v, want Rect(100,100,130,200)", left.Bounds)
		}
	})

	t.Run("Right", func(t *testing.T) {
		right := r.Right(30)
		if right.Bounds != image.Rect(170, 100, 200, 200) {
			t.Errorf("Right(30): got %v, want Rect(170,100,200,200)", right.Bounds)
		}
	})

	t.Run("Grow", func(t *testing.T) {
		grown := r.Grow(10)
		if grown.Bounds != image.Rect(90, 90, 210, 210) {
			t.Errorf("Grow(10): got %v, want Rect(90,90,210,210)", grown.Bounds)
		}
	})

	t.Run("Offset", func(t *testing.T) {
		offset := r.RegionOffset(5, -5)
		if offset.Bounds != image.Rect(105, 95, 205, 195) {
			t.Errorf("RegionOffset(5,-5): got %v, want Rect(105,95,205,195)", offset.Bounds)
		}
	})
}

func TestRegionFind(t *testing.T) {
	// Create a 100x100 screen image with a red square at (30, 40)
	screenImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			screenImg.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	for y := 40; y < 50; y++ {
		for x := 30; x < 40; x++ {
			screenImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Write needle as a PNG file
	tmpDir := t.TempDir()
	needlePath := filepath.Join(tmpDir, "needle.png")
	needleImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			needleImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	f, err := os.Create(needlePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, needleImg); err != nil {
		t.Fatal(err)
	}
	f.Close()

	backends := &Backends{
		Capture: tsikuritest.NewMockCapture(screenImg),
		Match:   match.New(),
		Input:   tsikuritest.NewMockInput(),
	}

	r := NewRegion(image.Rect(0, 0, 100, 100), backends, DefaultSettings())
	ctx := context.Background()

	m, err := r.Find(ctx, Image("needle.png").WithBasePath(tmpDir).WithSimilarity(0.8))
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if m.Score < 0.8 {
		t.Errorf("expected score >= 0.8, got %f", m.Score)
	}

	// Match center should be near (35, 45)
	if abs(m.Target.X-35) > 2 || abs(m.Target.Y-45) > 2 {
		t.Errorf("expected target near (35, 45), got (%d, %d)", m.Target.X, m.Target.Y)
	}
}

func TestRegionClick(t *testing.T) {
	screenImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	mockInput := tsikuritest.NewMockInput()

	backends := &Backends{
		Capture: tsikuritest.NewMockCapture(screenImg),
		Input:   mockInput,
	}

	r := NewRegion(image.Rect(0, 0, 100, 100), backends, DefaultSettings())
	ctx := context.Background()

	// Click at region center
	if err := r.Click(ctx); err != nil {
		t.Fatalf("Click failed: %v", err)
	}

	if !mockInput.AssertClicked(50, 50) {
		t.Error("expected click at center (50, 50)")
	}
}

func TestRegionClickLocation(t *testing.T) {
	screenImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	mockInput := tsikuritest.NewMockInput()

	backends := &Backends{
		Capture: tsikuritest.NewMockCapture(screenImg),
		Input:   mockInput,
	}

	r := NewRegion(image.Rect(0, 0, 100, 100), backends, DefaultSettings())
	ctx := context.Background()

	if err := r.Click(ctx, NewLocation(25, 75)); err != nil {
		t.Fatalf("Click failed: %v", err)
	}

	if !mockInput.AssertClicked(25, 75) {
		t.Error("expected click at (25, 75)")
	}
}

func TestRegionNoBackends(t *testing.T) {
	r := NewRegion(image.Rect(0, 0, 100, 100), nil, DefaultSettings())
	ctx := context.Background()

	_, err := r.Find(ctx, Image("test.png"))
	if err != ErrNoMatch {
		t.Errorf("expected ErrNoMatch, got %v", err)
	}

	err = r.Click(ctx)
	if err != ErrNoInput {
		t.Errorf("expected ErrNoInput, got %v", err)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
