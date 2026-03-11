package tsikuri

import "testing"

func TestLocationOffset(t *testing.T) {
	l := NewLocation(10, 20)

	got := l.Offset(5, -3)
	if got.X != 15 || got.Y != 17 {
		t.Errorf("Offset(5, -3): got (%d, %d), want (15, 17)", got.X, got.Y)
	}
}

func TestLocationAbove(t *testing.T) {
	l := NewLocation(10, 20)

	got := l.Above(5)
	if got.X != 10 || got.Y != 15 {
		t.Errorf("Above(5): got (%d, %d), want (10, 15)", got.X, got.Y)
	}
}

func TestLocationBelow(t *testing.T) {
	l := NewLocation(10, 20)

	got := l.Below(5)
	if got.X != 10 || got.Y != 25 {
		t.Errorf("Below(5): got (%d, %d), want (10, 25)", got.X, got.Y)
	}
}

func TestLocationLeft(t *testing.T) {
	l := NewLocation(10, 20)

	got := l.Left(3)
	if got.X != 7 || got.Y != 20 {
		t.Errorf("Left(3): got (%d, %d), want (7, 20)", got.X, got.Y)
	}
}

func TestLocationRight(t *testing.T) {
	l := NewLocation(10, 20)

	got := l.Right(3)
	if got.X != 13 || got.Y != 20 {
		t.Errorf("Right(3): got (%d, %d), want (13, 20)", got.X, got.Y)
	}
}
