package httpkit

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/tochka/core-kit/apikit"

	"github.com/tochka/core-kit/codec"
)

type Request struct {
	*http.Request
	codec codec.Codec
	param func(req *http.Request, key string) string
}

func (r *Request) Unmarshal(obj interface{}) error {
	if r.codec == nil {
		return apikit.ErrUnsupportedCodec
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apikit.ErrDecodeRequest
	}
	if err = r.codec.Unmarshal(body, obj); err != nil {
		return apikit.ErrDecodeRequest
	}
	return nil
}

func (r *Request) Param(key string) string {
	return r.param(r.Request, key)
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}
