package backend

import "image"

// OCRBackend extracts text from images.
type OCRBackend interface {
	// ReadText extracts all text from the given image.
	ReadText(img image.Image) (string, error)

	// Close releases OCR resources.
	Close() error
}
