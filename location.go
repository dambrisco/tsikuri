package tsikuri

import "context"

// Location represents a point on the screen.
type Location struct {
	X, Y int
}

// NewLocation creates a Location at the given coordinates.
func NewLocation(x, y int) Location {
	return Location{X: x, Y: y}
}

// Offset returns a new Location shifted by dx and dy.
func (l Location) Offset(dx, dy int) Location {
	return Location{X: l.X + dx, Y: l.Y + dy}
}

// Above returns a new Location shifted up by dy pixels.
func (l Location) Above(dy int) Location {
	return Location{X: l.X, Y: l.Y - dy}
}

// Below returns a new Location shifted down by dy pixels.
func (l Location) Below(dy int) Location {
	return Location{X: l.X, Y: l.Y + dy}
}

// Left returns a new Location shifted left by dx pixels.
func (l Location) Left(dx int) Location {
	return Location{X: l.X - dx, Y: l.Y}
}

// Right returns a new Location shifted right by dx pixels.
func (l Location) Right(dx int) Location {
	return Location{X: l.X + dx, Y: l.Y}
}

// resolve implements the Target interface. A Location resolves to itself.
func (l Location) resolve(_ context.Context, _ *Region) (Location, error) {
	return l, nil
}
