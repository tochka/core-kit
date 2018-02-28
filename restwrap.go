package corekit

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tochka/core-kit/apierror"
)

var (
	httpMetric  prometheus.Summary
	errorMetric prometheus.Counter
)

var regReplacer = regexp.MustCompile("[^a-zA-Z0-9_:]+")

// initMetrics is executed in NewService
func initMetrics(serviceName string) {
	serviceName = strings.ToLower(serviceName)
	serviceName = strings.Replace(serviceName, " ", "_", -1)
	serviceName = regReplacer.ReplaceAllString(serviceName, "")

	httpMetric = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "http_durations_seconds",
		Namespace:  "service_" + serviceName,
		Help:       "RPC latency distributions.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}})
	errorMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "errors_count",
		Namespace: "service_" + serviceName,
		Help:      "Count errors",
	})
	prometheus.MustRegister(httpMetric)
	prometheus.MustRegister(errorMetric)
}

// API Handler
type APIHandler func(req *http.Request) (interface{}, error)

func wrapAPIHandler(log func(format string, args ...interface{})) func(handler APIHandler) http.Handler {
	return func(handler APIHandler) http.Handler {
		wrap := func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			w.Header().Set("Content-Type", "application/json")

			result, err := handler(r)
			if err != nil {
				var apiErr apierror.APIError

				innerErr := errors.Cause(err)
				if apiErr, ok = innerErr.(apierror.APIError); !ok {
					log("[ERROR] API wrapper: %+v", err)
					errorMetric.Inc()

					apiErr = apierror.InternalServerErr
				}
				w.WriteHeader(apiErr.StatusCode)

				if apiErr.StatusCode != http.StatusBadRequest {
					return
				}

				b, _ := json.Marshal(apiErr)
				w.Write(b)
				return
			}

			if result == nil {
				w.WriteHeader(http.StatusCreated)
				return
			}

			w.WriteHeader(http.StatusOK)
			var body []byte
			if body, ok = result.([]byte); !ok {
				body, _ = json.Marshal(result)
			}
			w.Write(body)
		}

		return http.HandlerFunc(wrap)
	}
}

func rps(handler http.Handler) http.Handler {
	wrap := func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		handler.ServeHTTP(w, r)
		httpMetric.Observe(time.Since(begin).Seconds())
	}
	return http.HandlerFunc(wrap)
}
