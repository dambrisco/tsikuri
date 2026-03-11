package tsikuri

import "time"

// Option configures a Screen.
type Option func(*config)

type config struct {
	backends *Backends
	settings Settings
	basePath string
	monitor  int
}

// WithBackends sets custom backends.
func WithBackends(b *Backends) Option {
	return func(c *config) {
		c.backends = b
	}
}

// WithTimeout sets the default auto-wait timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.settings.AutoWaitTimeout = d
	}
}

// WithScanRate sets how often to re-scan during wait operations.
func WithScanRate(d time.Duration) Option {
	return func(c *config) {
		c.settings.ScanRate = d
	}
}

// WithBasePath sets the base directory for resolving relative image paths.
func WithBasePath(dir string) Option {
	return func(c *config) {
		c.basePath = dir
	}
}

// WithMonitor selects which monitor to use (0 = primary).
func WithMonitor(index int) Option {
	return func(c *config) {
		c.monitor = index
	}
}
