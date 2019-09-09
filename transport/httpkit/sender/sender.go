package sender

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tochka/core-kit/build"

	"github.com/tochka/core-kit/codec/json"

	"github.com/tochka/core-kit/apikit"
	"github.com/tochka/core-kit/codec"
	"github.com/tochka/core-kit/errors"
	"github.com/tochka/core-kit/metrics"
	"github.com/tochka/core-kit/metrics/provider"
	"github.com/tochka/core-kit/transport/httpkit"
)

var (
	initMetricsOnce = sync.Once{}
	metricsLatency  metrics.Histogram
)

// initMetrics it's lazy initialize metrics it's used for overide DefaultProvider
func initMetrics() {
	initMetricsOnce.Do(func() {
		metricsLatency = provider.DefaultProvider.NewHistogram("http_sender", "address", "method", "path", "status")
	})
}

func NewSender(address string, opts ...Option) *Sender {
	options := &Options{
		httpClient:   &http.Client{},
		healthPath:   httpkit.DefaultLivenessProb,
		errorHandler: DefaultErrorHandler,
		defaultCodec: json.NewCodec(),
	}

	for _, o := range opts {
		o(options)
	}
	sender := &Sender{
		options: options,
		address: address,
	}

	initMetrics()

	return sender
}

type Sender struct {
	address string
	options *Options
}

func (s *Sender) Send(ctx context.Context, req *Request, respObj interface{}) error {
	resp, err := s.SendWithResponse(ctx, req)
	if err != nil {
		return err
	}
	if respObj != nil {
		if err := resp.Unmarshal(respObj); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sender) SendWithResponse(ctx context.Context, req *Request) (result *Response, err error) {
	const contentTypeHeader = "Content-Type"
	startTime := time.Now()

	errCtx := errors.Context("component", "http_sender", "method", "send", "address", s.address,
		"http_method", req.Method, "http_path", req.Endpoint.Format)
	defer errors.Defer(&err, errCtx)

	if req.Header == nil {
		req.Header = http.Header{}
	}

	cd := s.options.defaultCodec

	var reqBody []byte
	if req.Payload != nil {
		if reqCd := codec.Get(req.Header.Get(contentTypeHeader)); reqCd != nil {
			cd = reqCd
		}
		errCtx.Add("request_codec", cd.Name())
		req.Header.Set(contentTypeHeader, cd.Name())

		var err error
		reqBody, err = cd.Marshal(req.Payload)
		if err != nil {
			return nil, err
		}
	}
	r, err := http.NewRequest(req.Method, s.address+req.Endpoint.String(), bytes.NewReader(reqBody))
	if err != nil {
		errCtx.Add("action", "http_new_request")
		return nil, err
	}

	for k, v := range req.Header {
		if len(v) != 0 {
			r.Header.Set(k, v[0])
		}
	}
	if r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", cd.Name())
	}
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", build.ServiceName+"/"+build.Version)
	}

	resp, err := s.options.httpClient.Do(r)
	if err != nil {
		errCtx.Add("action", "http_send_request")
		return nil, err
	}
	defer resp.Body.Close()

	defer metricsLatency.
		With("address", s.address, "method", req.Method, "path", req.Endpoint.Format, "status", strconv.Itoa(resp.StatusCode)).
		Observe(time.Since(startTime).Seconds())

	errCtx.Add("response_status", resp.StatusCode)

	if respCd := codec.Get(resp.Header.Get(contentTypeHeader)); respCd != nil {
		cd = respCd
	}
	errCtx.Add("response_codec", cd.Name())

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errCtx.Add("action", "read_response_body")
		return nil, err
	}
	errCtx.Add("response_body", base64.StdEncoding.EncodeToString(respBody))

	result = &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       respBody,
		cd:         cd,
	}
	if result.StatusCode/100 == 2 {
		result.errCtx = errCtx
		return result, nil
	}
	if err = s.options.errorHandler(result); err == nil {
		panicError := errors.Wrap(fmt.Errorf("HTTP sender error handler should return non nil error"), errCtx.Data...)
		panic(panicError)
	}
	return nil, err
}

func (s *Sender) Ping(ctx context.Context) (err error) {
	errCtx := errors.Context("component", "http_sender", "method", "ping", "address", s.address)
	defer errors.Defer(&err, errCtx)

	resp, err := s.options.httpClient.Get(s.address + s.options.healthPath)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		errCtx.Add("action", "send_request")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		errCtx.Add("response_status", resp.StatusCode)
		return fmt.Errorf("service doesn't ready")
	}
	return nil
}

func DefaultErrorHandler(resp *Response) error {
	apiErr := apikit.Error{
		Code: httpkit.APICodeFromHTTPStatus(resp.StatusCode),
	}
	if len(resp.Body) != 0 {
		if err := resp.Unmarshal(&apiErr); err != nil {
			return err
		}
	}
	return apiErr
}
