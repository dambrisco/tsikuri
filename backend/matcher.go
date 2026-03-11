package backend

import "image"

// MatchResult represents a single template match result.
type MatchResult struct {
	// Score is the similarity score between 0.0 and 1.0.
	Score float64

	// Rect is the bounding rectangle of the match in the haystack image.
	Rect image.Rectangle
}

// MatchBackend performs image template matching.
type MatchBackend interface {
	// FindBest finds the single best match of needle in haystack.
	// Returns the best match if score >= minScore, or nil if no match qualifies.
	FindBest(haystack, needle image.Image, minScore float64) (*MatchResult, error)

	// FindAll finds all non-overlapping matches with score >= minScore.
	// Results are sorted by score descending.
	FindAll(haystack, needle image.Image, minScore float64) ([]MatchResult, error)
}
