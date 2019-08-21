package httpkit

import (
	"github.com/tochka/core-kit/codec"
	"github.com/tochka/core-kit/ping"
)

type Option func(o *Options)

type Options struct {
	dependencies                       []ping.Pinger
	port                               int
	useDefaultCodedForParseRequestBody bool
	defaultCodec                       codec.Codec
	serveMux                           ServeMux
	errorConvertor                     func(err error) *Response

	certFile     string
	keyFile      string
	httpsEnabled bool
}

func Dependency(p ping.Pinger) Option {
	return func(o *Options) {
		o.dependencies = append(o.dependencies, p)
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

func DefaultCodec(c codec.Codec) Option {
	return func(o *Options) {
		o.defaultCodec = c
	}
}

//
// WARNING This setting will make it impossible to switch to other content types.
//
// UseDefaultCodecForParseRequestBody setup the default codec for parsing request
// body when content type was not set up
//
func UseDefaultCodecForParseRequestBody() Option {
	return func(o *Options) {
		o.useDefaultCodedForParseRequestBody = true
	}
}
