//go:build windows

package tsikuri

import (
	"github.com/dambrisco/tsikuri/match"
	"github.com/dambrisco/tsikuri/platform"
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
