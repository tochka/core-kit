package httpkit

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/tochka/core-kit/apikit"
	"github.com/tochka/core-kit/build"
	"github.com/tochka/core-kit/codec"
	"github.com/tochka/core-kit/codec/json"
	"github.com/tochka/core-kit/errors"
	"github.com/tochka/core-kit/ping"
)

// default probs
const (
	DefaultLivenessProb  = "/_service/liveness"
	DefaultReadinessProb = "/_service/readiness"
)

type APIHandler func(req *Request) (interface{}, error)

type Service interface {
	Get(path string, handler APIHandler)
	Post(path string, handler APIHandler)
	Put(path string, handler APIHandler)
	Del(path string, handler APIHandler)

	Run()
}

func NewService(opts ...Option) Service {
	options := &Options{
		dependencies:   make([]ping.Pinger, 0),
		serveMux:       newPatRouter(),
		errorConvertor: APIErrorConvertor,
		port:           8080,
		defaultCodec:   json.NewStrictCodec(),
	}

	for _, o := range opts {
		o(options)
	}
	options.serveMux = newMetricServeMux(options.serveMux)

	service := &service{
		options: *options,
	}
	return service
}

type service struct {
	options Options
}

func (s *service) Get(path string, handler APIHandler) {
	s.routeAdd(http.MethodGet, path, handler)
}

func (s *service) Post(path string, handler APIHandler) {
	s.routeAdd(http.MethodPost, path, handler)
}
func (s *service) Put(path string, handler APIHandler) {
	s.routeAdd(http.MethodPut, path, handler)
}
func (s *service) Del(path string, handler APIHandler) {
	s.routeAdd(http.MethodDelete, path, handler)
}

func (s *service) routeAdd(method, path string, handler APIHandler) {
	s.options.serveMux.Add(method, path, s.wrapAPIHandler(httpRequestInfoErrHandler(method, path, handler)))
}

func (s *service) Run() {
	log.Printf("[INFO] Service: %v version: %v. Start listening port %v\n",
		build.ServiceName, build.Version, s.options.port)

	m := http.NewServeMux()
	setupDefaultEndpoints(m, s)
	m.Handle("/", s.options.serveMux)

	server := http.Server{
		Addr:    fmt.Sprint(":", s.options.port),
		Handler: m,
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		log.Println("[INFO] Graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[ERROR] %+v\n", err)
		}

		log.Println("[INFO] Service stopped")
	}()

	var err error
	if s.options.httpsEnabled {
		err = server.ListenAndServeTLS(s.options.certFile, s.options.keyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		log.Printf("[ERROR] %+v\n", err)
	}
}

func setupDefaultEndpoints(m *http.ServeMux, s *service) {
	m.HandleFunc(DefaultLivenessProb, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	m.Handle("/_service/metrics", promhttp.Handler())

	currentReadness := false
	firstBackToNormal := true
	m.HandleFunc(DefaultReadinessProb, func(w http.ResponseWriter, r *http.Request) {
		newReadness := true
		for _, pinger := range s.options.dependencies {
			if err := pinger.Ping(context.Background()); err != nil {
				newReadness = false
				log.Printf("[ERROR] Dependency error: %+v\n", err)
			}
		}
		// go back to normal
		if !currentReadness && newReadness {
			msg := "[INFO] Service back to normal"

			if firstBackToNormal {
				firstBackToNormal = false
				msg = "[INFO] Service ready to handle requests"
			}
			log.Println(msg)
		}
		currentReadness = newReadness

		if currentReadness {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (s *service) wrapAPIHandler(h APIHandler) http.Handler {
	return s.encoding(func(req *Request) *Response {
		result, err := h(req)
		if err != nil {
			resp := s.options.errorConvertor(err)
			if resp.StatusCode/100 == 5 {
				log.Printf("[ERROR] %+v\n", err)
			}
			return resp
		}
		var response *Response
		switch v := result.(type) {
		case *Response:
			response = v
		case Response:
			response = &v
		default:
			response = &Response{
				StatusCode: http.StatusOK,
				Object:     v,
			}
		}
		return response
	})
}

func (s *service) encoding(f func(req *Request) *Response) http.Handler {
	const contentTypeHeader = "Content-Type"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cd codec.Codec
		if r.Header.Get(contentTypeHeader) != "" {
			// skip this error because we will check on non nil code when programmer try to unmarshal
			mimeType, _, err := mime.ParseMediaType(r.Header.Get(contentTypeHeader))
			if err == nil {
				switch mimeType {
				case json.Name:
					cd = json.NewStrictCodec()
				default:
					cd = codec.Get(mimeType)
				}
			}
		}
		if s.options.useDefaultCodedForParseRequestBody && cd == nil {
			cd = s.options.defaultCodec
		}

		response := f(&Request{Request: r, param: s.options.serveMux.GetPathParam, codec: cd})

		if response.StatusCode == 0 {
			panic("Response status code must be setup")
		}

		for k, v := range response.Header {
			if len(v) != 0 {
				w.Header().Set(k, v[0])
			}
		}

		if response.Object == nil {
			w.WriteHeader(response.StatusCode)
			return
		}

		if response.Header.Get(contentTypeHeader) != "" {
			cd = codec.Get(response.Header.Get(contentTypeHeader))
			if cd == nil {
				panic("Unsupported codec type " + response.Header.Get(contentTypeHeader))
			}
		} else {
			respTypes := strings.Split(r.Header.Get("Accept"), ",")
			for i := range respTypes {
				mimeType, _, err := mime.ParseMediaType(respTypes[i])
				if err != nil {
					continue
				}

				if respCd := codec.Get(mimeType); respCd != nil {
					cd = respCd
					break
				}
			}
		}
		if cd == nil {
			cd = s.options.defaultCodec
		}

		w.Header().Set(contentTypeHeader, cd.Name())
		w.WriteHeader(response.StatusCode)

		data, err := cd.Marshal(response.Object)
		if err != nil {
			log.Printf("[ERROR] cannot unmarshal response data: %+v\n", err)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Printf("[ERROR] cannot write response data: %+v\n", err)
		}
	})
}

func APIErrorConvertor(err error) *Response {
	var apiErr apikit.Error
	iErr, ok := errors.As(err, apiErr)
	if !ok {
		iErr = apikit.ErrInternalServer
	}

	apiErr = iErr.(apikit.Error)
	return &Response{
		StatusCode: HTTPStatusFromAPICode(apiErr.Code),
		Object:     apiErr,
	}
}
