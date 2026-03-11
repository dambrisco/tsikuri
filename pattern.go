package tsikuri

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

const defaultSimilarity = 0.7

// Pattern describes how to find an image on screen.
type Pattern struct {
	imagePath  string
	similarity float64
	offset     image.Point
	basePath   string
}

// Image creates a Pattern from an image file path.
// The path is resolved relative to the configured base path.
func Image(path string) *Pattern {
	return &Pattern{
		imagePath:  path,
		similarity: defaultSimilarity,
	}
}

// WithSimilarity returns a new Pattern with the given minimum match score (0.0–1.0).
func (p *Pattern) WithSimilarity(s float64) *Pattern {
	cp := *p
	cp.similarity = s
	return &cp
}

// WithOffset returns a new Pattern with a click offset from the match center.
func (p *Pattern) WithOffset(dx, dy int) *Pattern {
	cp := *p
	cp.offset = image.Pt(dx, dy)
	return &cp
}

// WithBasePath returns a new Pattern that resolves its image path relative to dir.
func (p *Pattern) WithBasePath(dir string) *Pattern {
	cp := *p
	cp.basePath = dir
	return &cp
}

// Similarity returns the minimum match score for this pattern.
func (p *Pattern) Similarity() float64 {
	return p.similarity
}

// LoadImage loads and decodes the pattern's image file.
func (p *Pattern) LoadImage() (image.Image, error) {
	path := p.resolvedPath()
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidImage, err)
	}
	return img, nil
}

func (p *Pattern) resolvedPath() string {
	if filepath.IsAbs(p.imagePath) {
		return p.imagePath
	}
	if p.basePath != "" {
		return filepath.Join(p.basePath, p.imagePath)
	}
	return p.imagePath
}

// resolve implements the Target interface. Finding the pattern on screen
// and returning the match center (plus offset).
func (p *Pattern) resolve(ctx context.Context, within *Region) (Location, error) {
	m, err := within.Find(ctx, p)
	if err != nil {
		return Location{}, err
	}
	return m.Target, nil
}
