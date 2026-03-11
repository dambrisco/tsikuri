package tsikuri

import "context"

// Match represents the result of a find operation.
// It embeds Region (the matched area) and adds score and target info.
type Match struct {
	*Region

	// Score is the similarity score of this match (0.0–1.0).
	Score float64

	// Target is the computed click point (match center + pattern offset).
	Target Location
}

// resolve implements the Target interface. A Match resolves to its Target location.
func (m *Match) resolve(_ context.Context, _ *Region) (Location, error) {
	return m.Target, nil
}
