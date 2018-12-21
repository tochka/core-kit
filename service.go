package corekit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bmizerany/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Service interface {
	Get(path string, handler APIHandler)
	Post(path string, handler APIHandler)
	Put(path string, handler APIHandler)
	Del(path string, handler APIHandler)
	Stream(path string, handler StreamAPIHandler)

	Run()
}

type ServeMux interface {
	Add(meth string, pat string, h http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Option func(o *Options)

type Options struct {
	name             string
	version          string
	dependencies     map[string]func() error
	dependenciesInfo map[string]func() interface{}
	params           map[string]string
	port             int
	certFile         string
	keyFile          string
	serveMux         ServeMux
	httpsEnabled     bool
	logger           func(format string, args ...interface{})
}

func Name(n string) Option {
	return func(o *Options) {
		o.name = n
	}
}

func Version(v string) Option {
	return func(o *Options) {
		o.version = v
	}
}

func DependencyInfo(name string, f func() interface{}) Option {
	return func(o *Options) {
		o.dependenciesInfo[name] = f
	}
}

func Dependency(name string, f func() error) Option {
	return func(o *Options) {
		o.dependencies[name] = f
	}
}

func Param(name, val string) Option {
	return func(o *Options) {
		o.params[name] = val
	}
}

func Port(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

func Https(certFile, keyFile string) Option {
	return func(o *Options) {
		o.certFile = certFile
		o.keyFile = keyFile
		o.httpsEnabled = true
	}
}

func UseServeMux(mux ServeMux) Option {
	return func(o *Options) {
		o.serveMux = mux
	}
}

func Logger(l func(format string, args ...interface{})) Option {
	return func(o *Options) {
		o.logger = l
	}
}

func NewService(opts ...Option) Service {

	defaultLogger := log.New(os.Stdout, "", log.LUTC|log.LstdFlags|log.Lmicroseconds)

	options := &Options{
		dependenciesInfo: make(map[string]func() interface{}),
		dependencies:     make(map[string]func() error),
		params:           map[string]string{},
		serveMux:         &adoptPatRouter{pat.New()},
		logger:           defaultLogger.Printf,
	}

	for _, o := range opts {
		o(options)
	}

	service := &service{
		options:          *options,
		wrapAPIHandler:   wrapAPIHandler(options.logger),
		streamAPIHandler: streamWrapAPIHandler(options.logger),
	}

	service.options.serveMux.Add(http.MethodGet, "/health/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.options.logger("[INFO] /health/status was deprecated. Please use /_service/status.\n")
		w.WriteHeader(http.StatusOK)
	}))

	// It's duplicate for deprecated way convention
	service.options.serveMux.Add(http.MethodGet, "/_service/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	// It's duplicate for new way convention
	service.options.serveMux.Add(http.MethodGet, "/_service/liveness", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	infoH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		dp := map[string]interface{}{}
		for name, d := range options.dependenciesInfo {
			dp[name] = d()
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":         options.name,
			"version":      options.version,
			"params":       options.params,
			"dependencies": dp,
		})
	})
	service.options.serveMux.Add(http.MethodGet, "/service/info", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.options.logger("[INFO] /service/info was deprecated. Please use /_service/info.\n")
		infoH.ServeHTTP(w, r)
	}))
	// It's duplicate for new way convention
	service.options.serveMux.Add(http.MethodGet, "/_service/info", infoH)

	readnessH := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusOK
		for name, check := range options.dependencies {
			if err := check(); err != nil {
				statusCode = http.StatusInternalServerError
				service.options.logger("[ERROR] %v dependency: %+v\n", name, err)
			}
		}
		w.WriteHeader(statusCode)
	})
	service.options.serveMux.Add(http.MethodGet, "/_service/health", readnessH)
	// It's duplicate for new way convention
	service.options.serveMux.Add(http.MethodGet, "/_service/readiness", readnessH)

	initMetrics(options.name)
	metricsH := promhttp.Handler()
	service.options.serveMux.Add(http.MethodGet, "/service/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.options.logger("[INFO] /service/metrics was deprecated. Please use /_service/metrics.\n")
		metricsH.ServeHTTP(w, r)
	}))
	// It's duplicate for new way convention
	service.options.serveMux.Add(http.MethodGet, "/_service/metrics", metricsH)

	return service
}

type service struct {
	options          Options
	wrapAPIHandler   func(handler APIHandler) http.Handler
	streamAPIHandler func(handler StreamAPIHandler) http.Handler
}

func (s *service) Get(path string, handler APIHandler) {
	s.options.serveMux.Add(http.MethodGet, path, s.wrapAPIHandler(handler))
}

func (s *service) Post(path string, handler APIHandler) {
	s.options.serveMux.Add(http.MethodPost, path, s.wrapAPIHandler(handler))
}
func (s *service) Put(path string, handler APIHandler) {
	s.options.serveMux.Add(http.MethodPut, path, s.wrapAPIHandler(handler))
}
func (s *service) Del(path string, handler APIHandler) {
	s.options.serveMux.Add(http.MethodDelete, path, s.wrapAPIHandler(handler))
}

func (s *service) Stream(path string, handler StreamAPIHandler) {
	s.options.serveMux.Add(http.MethodGet, path, s.streamAPIHandler(handler))
}

func (s *service) Run() {
	s.options.logger("[INFO] Start listening address :%v\n", s.options.port)

	server := http.Server{
		Addr:    fmt.Sprint(":", s.options.port),
		Handler: rps(s.options.serveMux),
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		s.options.logger("[INFO] Graceful shutdown...\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			s.options.logger("[ERROR] %+v\n", err)
		}

		s.options.logger("[INFO] Service stoped\n")
	}()

	var err error
	if s.options.httpsEnabled {
		err = server.ListenAndServeTLS(s.options.certFile, s.options.keyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		s.options.logger("[ERROR] %+v\n", err)
	}
}
