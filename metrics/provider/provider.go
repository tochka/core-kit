package provider

import (
	"github.com/tochka/core-kit/metrics"
	"github.com/tochka/core-kit/metrics/prometheus"
)

var DefaultProvider Provider = prometheus.Provider{}

// Provider abstracts over constructors and lifecycle management functions for
// each supported metrics backend. It should only be used by those who need to
// swap out implementations dynamically.
//
// This is primarily useful for intermediating frameworks, and is likely
// unnecessary for most Go kit services. See the package-level doc comment for
// more typical usage instructions.
type Provider interface {
	NewCounter(name string, labelValues ...string) metrics.Counter
	NewGauge(name string, labelValues ...string) metrics.Gauge
	NewHistogram(name string, labelValues ...string) metrics.Histogram
	Stop()
}
