package match

import (
	"image"
	"image/color"
	"testing"
)

func TestFindBest_ExactMatch(t *testing.T) {
	// Create a 100x100 haystack with a 10x10 red square at (30, 40)
	haystack := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			haystack.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}
	for y := 40; y < 50; y++ {
		for x := 30; x < 40; x++ {
			haystack.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Create a 10x10 red needle
	needle := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			needle.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	m := New()
	result, err := m.FindBest(haystack, needle, 0.9)
	if err != nil {
		t.Fatalf("FindBest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected a match, got nil")
	}

	if result.Score < 0.9 {
		t.Errorf("expected score >= 0.9, got %f", result.Score)
	}

	// Check position (should be near 30, 40)
	cx := (result.Rect.Min.X + result.Rect.Max.X) / 2
	cy := (result.Rect.Min.Y + result.Rect.Max.Y) / 2
	expectedCX := 35
	expectedCY := 45

	if abs(cx-expectedCX) > 2 || abs(cy-expectedCY) > 2 {
		t.Errorf("expected match center near (%d, %d), got (%d, %d)", expectedCX, expectedCY, cx, cy)
	}
}

func TestFindBest_NoMatch(t *testing.T) {
	// White haystack, black needle
	haystack := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			haystack.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	needle := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			needle.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	m := New()
	result, err := m.FindBest(haystack, needle, 0.9)
	if err != nil {
		t.Fatalf("FindBest failed: %v", err)
	}
	if result != nil {
		t.Errorf("expected no match, got score %f", result.Score)
	}
}

func TestFindBest_NeedleLargerThanHaystack(t *testing.T) {
	haystack := image.NewRGBA(image.Rect(0, 0, 5, 5))
	needle := image.NewRGBA(image.Rect(0, 0, 10, 10))

	m := New()
	result, err := m.FindBest(haystack, needle, 0.5)
	if err != nil {
		t.Fatalf("FindBest failed: %v", err)
	}
	if result != nil {
		t.Error("expected nil result when needle is larger than haystack")
	}
}

func TestFindAll_MultipleMatches(t *testing.T) {
	// Create haystack with a gradient pattern placed at two locations.
	// Using a gradient ensures NCC (not SSD fallback) which gives precise matches.
	haystack := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			haystack.Set(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	// Create a 10x10 gradient pattern
	makeGradient := func(img *image.RGBA, ox, oy int) {
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				v := uint8((x + y) * 25) // gradient
				img.Set(ox+x, oy+y, color.RGBA{R: v, G: 255 - v, B: v / 2, A: 255})
			}
		}
	}

	// Place gradient at (10, 10) and (70, 70)
	makeGradient(haystack, 10, 10)
	makeGradient(haystack, 70, 70)

	// Create needle with the same gradient
	needle := image.NewRGBA(image.Rect(0, 0, 10, 10))
	makeGradient(needle, 0, 0)

	m := New()
	results, err := m.FindAll(haystack, needle, 0.9)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestToGray(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255}) // white
	img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})       // black
	img.Set(0, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})     // red
	img.Set(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255})     // green

	gray := ToGray(img)

	if len(gray) != 2 || len(gray[0]) != 2 {
		t.Fatalf("expected 2x2 gray matrix, got %dx%d", len(gray), len(gray[0]))
	}

	// White should be ~1.0
	if gray[0][0] < 0.9 {
		t.Errorf("white pixel should be ~1.0, got %f", gray[0][0])
	}

	// Black should be ~0.0
	if gray[0][1] > 0.1 {
		t.Errorf("black pixel should be ~0.0, got %f", gray[0][1])
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
