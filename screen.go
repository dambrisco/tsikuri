package tsikuri

import "image"

// Screen represents a full monitor screen.
// It embeds Region covering the entire display.
type Screen struct {
	*Region
	Index int
}

// NewScreen creates a Screen for the given monitor index with the given options.
// Pass -1 or 0 for the primary monitor.
func NewScreen(opts ...Option) (*Screen, error) {
	cfg := &config{
		settings: DefaultSettings(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.backends == nil {
		cfg.backends = DefaultBackends()
	}

	if cfg.settings.BasePath == "" {
		cfg.settings.BasePath = cfg.basePath
	}

	monitor := cfg.monitor
	var bounds image.Rectangle

	if cfg.backends.Capture != nil {
		b, err := cfg.backends.Capture.ScreenBounds(monitor)
		if err != nil {
			return nil, err
		}
		bounds = b
	} else {
		// Fallback: assume a reasonable default
		bounds = image.Rect(0, 0, 1920, 1080)
	}

	region := NewRegion(bounds, cfg.backends, cfg.settings)

	return &Screen{
		Region: region,
		Index:  monitor,
	}, nil
}
