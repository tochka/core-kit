package httpkit

import (
	"github.com/tochka/core-kit/errors"
)

func httpRequestInfoErrHandler(meth, pat string, h APIHandler) APIHandler {
	return func(req *Request) (interface{}, error) {
		r, err := h(req)
		if err != nil {
			return nil, errors.Wrap(err, "method", meth, "path", pat, "user-agent", req.UserAgent())
		}
		return r, nil
	}
}
