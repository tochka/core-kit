package apierror

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (err APIError) Error() string {
	return fmt.Sprint("code: ", err.Code, " Message: ", err.Message)
}

var (
	// STATUS CODE: 500
	InternalServerErr = APIError{
		StatusCode: http.StatusInternalServerError,
	}

	// STATUS CODE: 401
	UnauthorizedRequestErr = APIError{
		StatusCode: http.StatusUnauthorized,
	}

	// STATUS CODE: 403
	ForbiddenErr = APIError{
		StatusCode: http.StatusForbidden,
	}

	// STATUS CODE: 404
	EntityNotFoundErr = APIError{
		StatusCode: http.StatusNotFound,
	}

	// STATUS CODE: 400
	JSONInvalidErr = APIError{
		Code:       30000,
		StatusCode: http.StatusBadRequest,
		Message:    "JSON specified as a request is invalid",
	}
	SnapshotIncorrectErr = APIError{
		Code:       30001,
		StatusCode: http.StatusBadRequest,
		Message:    "Request snapshot invalid",
	}
)
