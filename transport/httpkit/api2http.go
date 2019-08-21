package httpkit

import (
	"log"
	"net/http"

	"github.com/tochka/core-kit/apikit"
)

func HTTPStatusFromAPICode(code apikit.Code) int {
	switch code {
	case apikit.Canceled:
		return http.StatusRequestTimeout
	case apikit.Unknown:
		return http.StatusInternalServerError
	case apikit.InvalidArgument:
		return http.StatusBadRequest
	case apikit.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case apikit.NotFound:
		return http.StatusNotFound
	case apikit.AlreadyExists:
		return http.StatusConflict
	case apikit.PermissionDenied:
		return http.StatusForbidden
	case apikit.Unauthenticated:
		return http.StatusUnauthorized
	case apikit.ResourceExhausted:
		return http.StatusTooManyRequests
	case apikit.FailedPrecondition:
		return http.StatusPreconditionFailed
	case apikit.Aborted:
		return http.StatusConflict
	case apikit.OutOfRange:
		return http.StatusBadRequest
	case apikit.Unimplemented:
		return http.StatusNotImplemented
	case apikit.Internal:
		return http.StatusInternalServerError
	case apikit.Unavailable:
		return http.StatusServiceUnavailable
	case apikit.DataLoss:
		return http.StatusInternalServerError
	}

	log.Printf("[ERROR] Unknown API error code: %v\n", code)
	return http.StatusInternalServerError
}

func APICodeFromHTTPStatus(status int) apikit.Code {
	switch status {
	case http.StatusRequestTimeout:
		return apikit.Canceled
	case http.StatusInternalServerError:
		return apikit.Internal
	case http.StatusBadRequest:
		return apikit.InvalidArgument
	case http.StatusGatewayTimeout:
		return apikit.DeadlineExceeded
	case http.StatusNotFound:
		return apikit.NotFound
	case http.StatusConflict:
		return apikit.AlreadyExists
	case http.StatusForbidden:
		return apikit.PermissionDenied
	case http.StatusUnauthorized:
		return apikit.Unauthenticated
	case http.StatusTooManyRequests:
		return apikit.ResourceExhausted
	case http.StatusPreconditionFailed:
		return apikit.FailedPrecondition
	case http.StatusNotImplemented:
		return apikit.Unimplemented
	case http.StatusServiceUnavailable:
		return apikit.Unavailable
	}

	log.Printf("[ERROR] Unknown HTTP status: %v\n", status)
	return apikit.Internal
}
