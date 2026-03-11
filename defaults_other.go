//go:build !windows && !linux

package tsikuri

import "github.com/dambrisco/tsikuri/internal/match"

// DefaultBackends returns minimal backends for unsupported platforms.
// Only the pure-Go matcher is available; capture and input backends
// must be provided by the user.
func DefaultBackends() *Backends {
	return &Backends{
		Match: match.New(),
	}
}
