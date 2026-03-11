package tsikuri

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/dambrisco/tsikuri/backend"
)

// Backends holds the pluggable backend implementations.
type Backends struct {
	Capture backend.CaptureBackend
	Match   backend.MatchBackend
	Input   backend.InputBackend
	OCR     backend.OCRBackend
}

// Settings controls Region behavior.
type Settings struct {
	// AutoWaitTimeout is the default timeout for Wait operations.
	AutoWaitTimeout time.Duration

	// ScanRate is how often to re-scan during Wait operations.
	ScanRate time.Duration

	// BasePath is the base directory for resolving relative image paths.
	BasePath string
}

// DefaultSettings returns sensible default settings.
func DefaultSettings() Settings {
	return Settings{
		AutoWaitTimeout: 3 * time.Second,
		ScanRate:        50 * time.Millisecond,
	}
}

// Region represents a rectangular area of the screen with
// find and interaction capabilities.
type Region struct {
	Bounds   image.Rectangle
	backends *Backends
	settings Settings
}

// NewRegion creates a Region with the given bounds and configuration.
func NewRegion(bounds image.Rectangle, backends *Backends, settings Settings) *Region {
	return &Region{
		Bounds:   bounds,
		backends: backends,
		settings: settings,
	}
}

// resolve implements the Target interface. A Region resolves to its center.
func (r *Region) resolve(_ context.Context, _ *Region) (Location, error) {
	center := r.center()
	return center, nil
}

func (r *Region) center() Location {
	return Location{
		X: (r.Bounds.Min.X + r.Bounds.Max.X) / 2,
		Y: (r.Bounds.Min.Y + r.Bounds.Max.Y) / 2,
	}
}

// --- Find operations ---

// Find locates the best match for the given pattern within this region.
func (r *Region) Find(ctx context.Context, pat *Pattern) (*Match, error) {
	if r.backends == nil || r.backends.Match == nil {
		return nil, ErrNoMatch
	}
	if r.backends.Capture == nil {
		return nil, ErrNoCapture
	}

	screenshot, err := r.Capture()
	if err != nil {
		return nil, fmt.Errorf("tsikuri: capture failed: %w", err)
	}

	p := r.resolvePattern(pat)
	needle, err := p.LoadImage()
	if err != nil {
		return nil, err
	}

	result, err := r.backends.Match.FindBest(screenshot, needle, p.similarity)
	if err != nil {
		return nil, fmt.Errorf("tsikuri: match failed: %w", err)
	}
	if result == nil {
		return nil, ErrNotFound
	}

	return r.matchFromResult(result, p), nil
}

// FindAll returns all matches for the pattern within this region.
func (r *Region) FindAll(ctx context.Context, pat *Pattern) ([]*Match, error) {
	if r.backends == nil || r.backends.Match == nil {
		return nil, ErrNoMatch
	}
	if r.backends.Capture == nil {
		return nil, ErrNoCapture
	}

	screenshot, err := r.Capture()
	if err != nil {
		return nil, fmt.Errorf("tsikuri: capture failed: %w", err)
	}

	p := r.resolvePattern(pat)
	needle, err := p.LoadImage()
	if err != nil {
		return nil, err
	}

	results, err := r.backends.Match.FindAll(screenshot, needle, p.similarity)
	if err != nil {
		return nil, fmt.Errorf("tsikuri: match failed: %w", err)
	}

	matches := make([]*Match, len(results))
	for i := range results {
		matches[i] = r.matchFromResult(&results[i], p)
	}
	return matches, nil
}

// Exists checks if the pattern exists right now, without waiting.
// Returns the Match if found, or nil if not found.
func (r *Region) Exists(pat *Pattern) *Match {
	m, err := r.Find(context.Background(), pat)
	if err != nil {
		return nil
	}
	return m
}

// Wait waits for the pattern to appear within the region.
// Uses the context deadline, or falls back to AutoWaitTimeout.
func (r *Region) Wait(ctx context.Context, pat *Pattern) (*Match, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	ticker := time.NewTicker(r.settings.ScanRate)
	defer ticker.Stop()

	// Try immediately first
	m, err := r.Find(ctx, pat)
	if err == nil {
		return m, nil
	}
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case <-ticker.C:
			m, err := r.Find(ctx, pat)
			if err == nil {
				return m, nil
			}
			if err != ErrNotFound {
				return nil, err
			}
		}
	}
}

// WaitVanish waits until the pattern is no longer visible.
func (r *Region) WaitVanish(ctx context.Context, pat *Pattern) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	ticker := time.NewTicker(r.settings.ScanRate)
	defer ticker.Stop()

	// Check immediately
	m := r.Exists(pat)
	if m == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ErrTimeout
		case <-ticker.C:
			m := r.Exists(pat)
			if m == nil {
				return nil
			}
		}
	}
}

// FindBest finds the best match among multiple patterns.
// Returns the match, the index of the matched pattern, and any error.
func (r *Region) FindBest(ctx context.Context, pats ...*Pattern) (*Match, int, error) {
	var bestMatch *Match
	bestIdx := -1

	for i, pat := range pats {
		m, err := r.Find(ctx, pat)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, -1, err
		}
		if bestMatch == nil || m.Score > bestMatch.Score {
			bestMatch = m
			bestIdx = i
		}
	}

	if bestMatch == nil {
		return nil, -1, ErrNotFound
	}
	return bestMatch, bestIdx, nil
}

// FindAny finds matches for any of the given patterns.
// Returns all found matches.
func (r *Region) FindAny(ctx context.Context, pats ...*Pattern) ([]*Match, error) {
	var found []*Match
	for _, pat := range pats {
		m, err := r.Find(ctx, pat)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		found = append(found, m)
	}
	return found, nil
}

// --- Mouse operations ---

// Click clicks the center of this region, or finds and clicks the given target.
func (r *Region) Click(ctx context.Context, targets ...Target) error {
	loc, err := r.resolveTarget(ctx, targets)
	if err != nil {
		return err
	}
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	return r.backends.Input.MouseClick(loc.X, loc.Y, backend.ButtonLeft)
}

// DoubleClick double-clicks the target.
func (r *Region) DoubleClick(ctx context.Context, targets ...Target) error {
	loc, err := r.resolveTarget(ctx, targets)
	if err != nil {
		return err
	}
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	return r.backends.Input.MouseDoubleClick(loc.X, loc.Y, backend.ButtonLeft)
}

// RightClick right-clicks the target.
func (r *Region) RightClick(ctx context.Context, targets ...Target) error {
	loc, err := r.resolveTarget(ctx, targets)
	if err != nil {
		return err
	}
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	return r.backends.Input.MouseClick(loc.X, loc.Y, backend.ButtonRight)
}

// Hover moves the mouse to the target without clicking.
func (r *Region) Hover(ctx context.Context, targets ...Target) error {
	loc, err := r.resolveTarget(ctx, targets)
	if err != nil {
		return err
	}
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	return r.backends.Input.MouseMove(loc.X, loc.Y)
}

// DragDrop drags from one target to another.
func (r *Region) DragDrop(ctx context.Context, from, to Target) error {
	fromLoc, err := from.resolve(ctx, r)
	if err != nil {
		return fmt.Errorf("tsikuri: resolve drag source: %w", err)
	}
	toLoc, err := to.resolve(ctx, r)
	if err != nil {
		return fmt.Errorf("tsikuri: resolve drag target: %w", err)
	}
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	return r.backends.Input.DragDrop(fromLoc.X, fromLoc.Y, toLoc.X, toLoc.Y)
}

// Wheel scrolls within the region.
func (r *Region) Wheel(direction ScrollDirection, steps int) error {
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	center := r.center()
	return r.backends.Input.Scroll(center.X, center.Y, backend.ScrollDirection(direction), steps)
}

// --- Keyboard operations ---

// Type types the given text string, optionally with modifier keys held.
func (r *Region) Type(text string, modifiers ...Key) error {
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}

	// Hold modifiers
	for _, mod := range modifiers {
		if err := r.backends.Input.KeyDown(string(mod)); err != nil {
			return err
		}
	}

	err := r.backends.Input.TypeText(text)

	// Release modifiers in reverse order
	for i := len(modifiers) - 1; i >= 0; i-- {
		if rerr := r.backends.Input.KeyUp(string(modifiers[i])); rerr != nil && err == nil {
			err = rerr
		}
	}

	return err
}

// Paste pastes text via the clipboard.
func (r *Region) Paste(text string) error {
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	if err := r.backends.Input.SetClipboard(text); err != nil {
		return err
	}
	return r.backends.Input.PasteClipboard()
}

// KeyDown presses keys without releasing them.
func (r *Region) KeyDown(keys ...Key) error {
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	for _, k := range keys {
		if err := r.backends.Input.KeyDown(string(k)); err != nil {
			return err
		}
	}
	return nil
}

// KeyUp releases previously pressed keys.
func (r *Region) KeyUp(keys ...Key) error {
	if r.backends == nil || r.backends.Input == nil {
		return ErrNoInput
	}
	for _, k := range keys {
		if err := r.backends.Input.KeyUp(string(k)); err != nil {
			return err
		}
	}
	return nil
}

// --- OCR ---

// Text extracts text from this region using OCR.
func (r *Region) Text() (string, error) {
	if r.backends == nil || r.backends.OCR == nil {
		return "", ErrNoOCR
	}
	img, err := r.Capture()
	if err != nil {
		return "", fmt.Errorf("tsikuri: capture for OCR failed: %w", err)
	}
	return r.backends.OCR.ReadText(img)
}

// --- Region manipulation ---

// RegionOffset returns a new Region shifted by dx, dy.
func (r *Region) RegionOffset(dx, dy int) *Region {
	return &Region{
		Bounds:   r.Bounds.Add(image.Pt(dx, dy)),
		backends: r.backends,
		settings: r.settings,
	}
}

// Grow returns a new Region expanded by pixels in each direction.
func (r *Region) Grow(pixels int) *Region {
	return &Region{
		Bounds: image.Rect(
			r.Bounds.Min.X-pixels,
			r.Bounds.Min.Y-pixels,
			r.Bounds.Max.X+pixels,
			r.Bounds.Max.Y+pixels,
		),
		backends: r.backends,
		settings: r.settings,
	}
}

// Above returns a new Region above this one with the given height.
func (r *Region) Above(height int) *Region {
	return &Region{
		Bounds: image.Rect(
			r.Bounds.Min.X,
			r.Bounds.Min.Y-height,
			r.Bounds.Max.X,
			r.Bounds.Min.Y,
		),
		backends: r.backends,
		settings: r.settings,
	}
}

// Below returns a new Region below this one with the given height.
func (r *Region) Below(height int) *Region {
	return &Region{
		Bounds: image.Rect(
			r.Bounds.Min.X,
			r.Bounds.Max.Y,
			r.Bounds.Max.X,
			r.Bounds.Max.Y+height,
		),
		backends: r.backends,
		settings: r.settings,
	}
}

// Left returns a new Region to the left of this one with the given width.
// If used on a Screen, returns the leftmost portion of this region.
func (r *Region) Left(width int) *Region {
	return &Region{
		Bounds: image.Rect(
			r.Bounds.Min.X,
			r.Bounds.Min.Y,
			r.Bounds.Min.X+width,
			r.Bounds.Max.Y,
		),
		backends: r.backends,
		settings: r.settings,
	}
}

// Right returns a new Region to the right of this one with the given width.
// If used on a Screen, returns the rightmost portion of this region.
func (r *Region) Right(width int) *Region {
	return &Region{
		Bounds: image.Rect(
			r.Bounds.Max.X-width,
			r.Bounds.Min.Y,
			r.Bounds.Max.X,
			r.Bounds.Max.Y,
		),
		backends: r.backends,
		settings: r.settings,
	}
}

// --- Highlight ---

// Highlight draws a rectangle around this region for visual debugging.
func (r *Region) Highlight(d time.Duration) error {
	// Highlight is a best-effort operation — requires platform support.
	// For now this is a no-op; platform implementations can be added.
	_ = d
	return nil
}

// --- Screenshot ---

// Capture takes a screenshot of this region.
func (r *Region) Capture() (*image.RGBA, error) {
	if r.backends == nil || r.backends.Capture == nil {
		return nil, ErrNoCapture
	}
	return r.backends.Capture.CaptureRegion(
		r.Bounds.Min.X,
		r.Bounds.Min.Y,
		r.Bounds.Dx(),
		r.Bounds.Dy(),
	)
}

// --- Internal helpers ---

func (r *Region) resolveTarget(ctx context.Context, targets []Target) (Location, error) {
	if len(targets) == 0 {
		return r.center(), nil
	}
	return targets[0].resolve(ctx, r)
}

func (r *Region) resolvePattern(pat *Pattern) *Pattern {
	if pat.basePath == "" && r.settings.BasePath != "" {
		return pat.WithBasePath(r.settings.BasePath)
	}
	return pat
}

func (r *Region) matchFromResult(result *backend.MatchResult, pat *Pattern) *Match {
	matchRegion := &Region{
		Bounds:   result.Rect,
		backends: r.backends,
		settings: r.settings,
	}
	center := Location{
		X: (result.Rect.Min.X + result.Rect.Max.X) / 2,
		Y: (result.Rect.Min.Y + result.Rect.Max.Y) / 2,
	}
	target := Location{
		X: center.X + pat.offset.X,
		Y: center.Y + pat.offset.Y,
	}
	return &Match{
		Region: matchRegion,
		Score:  result.Score,
		Target: target,
	}
}

func (r *Region) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, r.settings.AutoWaitTimeout)
}
