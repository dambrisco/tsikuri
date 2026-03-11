//go:build !windows

package platform

import (
	"errors"
	"image"
)

// StubOCR is a no-op OCR backend for platforms without native OCR.
type StubOCR struct{}

// NewOCR creates a stub OCR backend that returns an error for all operations.
func NewOCR() *StubOCR {
	return &StubOCR{}
}

func (o *StubOCR) ReadText(img image.Image) (string, error) {
	return "", errors.New("tsikuri: OCR backend not configured (not available on this platform)")
}

func (o *StubOCR) Close() error {
	return nil
}
