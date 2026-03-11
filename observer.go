package tsikuri

import (
	"context"
	"time"
)

// EventType represents the type of observer event.
type EventType int

const (
	// EventAppear fires when a pattern appears on screen.
	EventAppear EventType = iota
	// EventVanish fires when a pattern disappears from screen.
	EventVanish
	// EventChange fires when the region's visual content changes.
	EventChange
)

// ObserveEvent contains information about an observed event.
type ObserveEvent struct {
	Type    EventType
	Pattern *Pattern
	Region  *Region
	Match   *Match
}

// ObserveHandler is called when an observed event occurs.
type ObserveHandler func(event ObserveEvent)

type handler struct {
	eventType EventType
	pattern   *Pattern
	callback  ObserveHandler
}

// Observer watches a region for visual changes.
type Observer struct {
	region   *Region
	handlers []handler
	scanRate time.Duration
}

// OnAppear registers a handler that fires when the pattern appears.
// Returns the Observer for adding additional handlers.
func (r *Region) OnAppear(pat *Pattern, h ObserveHandler) *Observer {
	obs := &Observer{
		region:   r,
		scanRate: r.settings.ScanRate,
	}
	obs.handlers = append(obs.handlers, handler{
		eventType: EventAppear,
		pattern:   pat,
		callback:  h,
	})
	return obs
}

// OnVanish registers a handler that fires when the pattern disappears.
func (r *Region) OnVanish(pat *Pattern, h ObserveHandler) *Observer {
	obs := &Observer{
		region:   r,
		scanRate: r.settings.ScanRate,
	}
	obs.handlers = append(obs.handlers, handler{
		eventType: EventVanish,
		pattern:   pat,
		callback:  h,
	})
	return obs
}

// OnChange registers a handler that fires when the region content changes.
func (r *Region) OnChange(h ObserveHandler) *Observer {
	obs := &Observer{
		region:   r,
		scanRate: r.settings.ScanRate,
	}
	obs.handlers = append(obs.handlers, handler{
		eventType: EventChange,
		callback:  h,
	})
	return obs
}

// Also adds another handler to this observer.
func (o *Observer) Also(eventType EventType, pat *Pattern, h ObserveHandler) *Observer {
	o.handlers = append(o.handlers, handler{
		eventType: eventType,
		pattern:   pat,
		callback:  h,
	})
	return o
}

// Run starts the observer loop. Blocks until ctx is cancelled.
func (o *Observer) Run(ctx context.Context) error {
	ticker := time.NewTicker(o.scanRate)
	defer ticker.Stop()

	// Track previous state for each handler
	type state struct {
		wasPresent bool
		lastImg    []byte // simple hash for change detection
	}
	states := make([]state, len(o.handlers))

	// Initialize states
	for i, h := range o.handlers {
		if h.pattern != nil {
			m := o.region.Exists(h.pattern)
			states[i].wasPresent = m != nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for i, h := range o.handlers {
				switch h.eventType {
				case EventAppear:
					if h.pattern == nil {
						continue
					}
					m := o.region.Exists(h.pattern)
					isPresent := m != nil
					if isPresent && !states[i].wasPresent {
						h.callback(ObserveEvent{
							Type:    EventAppear,
							Pattern: h.pattern,
							Region:  o.region,
							Match:   m,
						})
					}
					states[i].wasPresent = isPresent

				case EventVanish:
					if h.pattern == nil {
						continue
					}
					m := o.region.Exists(h.pattern)
					isPresent := m != nil
					if !isPresent && states[i].wasPresent {
						h.callback(ObserveEvent{
							Type:    EventVanish,
							Pattern: h.pattern,
							Region:  o.region,
						})
					}
					states[i].wasPresent = isPresent

				case EventChange:
					// Change detection: capture and compare raw pixel data
					img, err := o.region.Capture()
					if err != nil {
						continue
					}
					currentHash := simpleHash(img.Pix)
					if states[i].lastImg != nil && !bytesEqual(currentHash, states[i].lastImg) {
						h.callback(ObserveEvent{
							Type:   EventChange,
							Region: o.region,
						})
					}
					states[i].lastImg = currentHash
				}
			}
		}
	}
}

// simpleHash returns a quick fingerprint of pixel data.
// Samples every 1024th byte to keep it fast.
func simpleHash(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	step := 1024
	if step > len(data) {
		step = 1
	}
	hash := make([]byte, 0, len(data)/step+1)
	for i := 0; i < len(data); i += step {
		hash = append(hash, data[i])
	}
	return hash
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
