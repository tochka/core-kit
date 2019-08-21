package httpkit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/tochka/core-kit/metrics"
	"github.com/tochka/core-kit/metrics/provider"
)

func newMetricServeMux(mux ServeMux) ServeMux {
	return &metricServeMux{
		mux:     mux,
		latency: provider.DefaultProvider.NewHistogram("http_latency", "method", "path", "status"),
	}
}

type metricServeMux struct {
	mux     ServeMux
	latency metrics.Histogram
}

type statusCodeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *statusCodeResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (m *metricServeMux) Add(meth string, pat string, h http.Handler) {
	wrap := func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		statusCodeRW := &statusCodeResponseWriter{w, http.StatusOK}
		h.ServeHTTP(statusCodeRW, r)

		m.latency.
			With("method", meth, "path", pat, "status", strconv.Itoa(statusCodeRW.statusCode)).
			Observe(time.Since(t).Seconds())
	}
	m.mux.Add(meth, pat, http.HandlerFunc(wrap))
}

func (m *metricServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *metricServeMux) GetPathParam(req *http.Request, key string) string {
	return m.mux.GetPathParam(req, key)
}
