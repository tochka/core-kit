package sender

import (
	"net/http"

	"github.com/tochka/core-kit/codec"
)

type Option func(o *Options)

type Options struct {
	httpClient   *http.Client
	healthPath   string
	errorHandler func(resp *Response) error
	defaultCodec codec.Codec
}

func HTTPClient(c *http.Client) Option {
	return func(o *Options) {
		o.httpClient = c
	}
}

func HealthPath(path string) Option {
	return func(o *Options) {
		o.healthPath = path
	}
}

func ErrorHandler(h func(resp *Response) error) Option {
	return func(o *Options) {
		o.errorHandler = h
	}
}

func DefaultCodec(c codec.Codec) Option {
	return func(o *Options) {
		o.defaultCodec = c
	}
}
