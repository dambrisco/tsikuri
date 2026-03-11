//go:build windows

package tsikuri

import (
	"github.com/dambrisco/tsikuri/internal/match"
	"github.com/dambrisco/tsikuri/internal/platform"
)

// DefaultBackends returns the default backends for Windows.
func DefaultBackends() *Backends {
	return &Backends{
		Capture: platform.NewCapture(),
		Match:   match.New(),
		Input:   platform.NewInput(),
		OCR:     platform.NewOCR(),
	}
}
