package sender

import (
	"fmt"
	"net/http"

	"github.com/tochka/core-kit/errors"

	"github.com/tochka/core-kit/codec"
)

type Request struct {
	Method   string
	Endpoint RequestEndpoint
	Header   http.Header
	Payload  interface{}
}

type RequestEndpoint struct {
	Format string
	Params []interface{}
}

func (e RequestEndpoint) String() string {
	return fmt.Sprintf(e.Format, e.Params...)
}

func Endpoint(format string, params ...interface{}) RequestEndpoint {
	return RequestEndpoint{
		Format: format,
		Params: params,
	}
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	cd         codec.Codec
	errCtx     *errors.ErrorContext
}

func (r *Response) Unmarshal(v interface{}) error {
	err := r.cd.Unmarshal(r.Body, v)
	if err != nil && r.errCtx != nil {
		return errors.Wrap(err, r.errCtx.Data...)
	}
	return nil
}
