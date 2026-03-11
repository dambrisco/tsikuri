package tsikuri

import "context"

// Target is anything that can be clicked or interacted with.
// It can resolve itself to a screen Location, potentially performing
// a find operation in the process.
//
// Built-in implementations: Location, *Pattern, *Match, *Region.
type Target interface {
	resolve(ctx context.Context, within *Region) (Location, error)
}
