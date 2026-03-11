package tsikuri

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPatternDefaults(t *testing.T) {
	p := Image("button.png")
	if p.similarity != defaultSimilarity {
		t.Errorf("default similarity: got %f, want %f", p.similarity, defaultSimilarity)
	}
	if p.offset != (image.Point{}) {
		t.Errorf("default offset: got %v, want (0,0)", p.offset)
	}
}

func TestPatternWithSimilarity(t *testing.T) {
	p := Image("button.png").WithSimilarity(0.95)
	if p.similarity != 0.95 {
		t.Errorf("similarity: got %f, want 0.95", p.similarity)
	}
}

func TestPatternWithOffset(t *testing.T) {
	p := Image("button.png").WithOffset(10, -5)
	if p.offset.X != 10 || p.offset.Y != -5 {
		t.Errorf("offset: got (%d, %d), want (10, -5)", p.offset.X, p.offset.Y)
	}
}

func TestPatternImmutable(t *testing.T) {
	p1 := Image("button.png")
	p2 := p1.WithSimilarity(0.95)

	if p1.similarity == p2.similarity {
		t.Error("WithSimilarity should return a new Pattern, not mutate the original")
	}
}

func TestPatternLoadImage(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	p := Image("test.png").WithBasePath(tmpDir)
	loaded, err := p.LoadImage()
	if err != nil {
		t.Fatalf("LoadImage failed: %v", err)
	}

	bounds := loaded.Bounds()
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("loaded image size: got %dx%d, want 5x5", bounds.Dx(), bounds.Dy())
	}
}

func TestPatternLoadImageNotFound(t *testing.T) {
	p := Image("nonexistent.png")
	_, err := p.LoadImage()
	if err == nil {
		t.Error("expected error for nonexistent image")
	}
}
