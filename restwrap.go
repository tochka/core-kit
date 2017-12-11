package corekit

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/tochka/core-kit/apierror"
)

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
