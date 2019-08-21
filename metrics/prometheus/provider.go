package prometheus

import (
	"regexp"
	"strings"

	"github.com/tochka/core-kit/build"
	"github.com/tochka/core-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	reg          = regexp.MustCompile("[^\\w]+")
	sanitizeName = func(name string) string {
		return strings.ToLower(reg.ReplaceAllString(name, "_"))
	}

	namespace = sanitizeName(build.ServiceName)
)

type Provider struct{}

func (_ Provider) NewCounter(name string, labelValues ...string) metrics.Counter {
	return NewCounterFrom(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      sanitizeName(name),
		Help:      name,
	}, labelValues)
}

func (_ Provider) NewGauge(name string, labelValues ...string) metrics.Gauge {
	return NewGaugeFrom(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      sanitizeName(name),
		Help:      name,
	}, labelValues)
}

func (_ Provider) NewHistogram(name string, labelValues ...string) metrics.Histogram {
	return NewHistogramFrom(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      sanitizeName(name),
		Help:      name,
	}, labelValues)
}

func (_ Provider) Stop() {}
