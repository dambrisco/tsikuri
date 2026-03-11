package tsikuri

import "errors"

var (
	// ErrNotFound is returned when a pattern cannot be found on screen.
	ErrNotFound = errors.New("tsikuri: pattern not found")

	// ErrTimeout is returned when a wait operation exceeds its deadline.
	ErrTimeout = errors.New("tsikuri: wait timed out")

	// ErrNoOCR is returned when OCR is requested but no OCR backend is configured.
	ErrNoOCR = errors.New("tsikuri: OCR backend not configured")

	// ErrInvalidImage is returned when an image file cannot be loaded.
	ErrInvalidImage = errors.New("tsikuri: could not load image file")

	// ErrOutOfBounds is returned when a region extends outside the screen.
	ErrOutOfBounds = errors.New("tsikuri: region out of screen bounds")

	// ErrNoCapture is returned when no capture backend is configured.
	ErrNoCapture = errors.New("tsikuri: capture backend not configured")

	// ErrNoInput is returned when no input backend is configured.
	ErrNoInput = errors.New("tsikuri: input backend not configured")

	// ErrNoMatch is returned when no match backend is configured.
	ErrNoMatch = errors.New("tsikuri: match backend not configured")
)
