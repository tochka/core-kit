package apikit

import "fmt"

//
// Error is API error that service or dependency should return
//
type Error struct {
	Code    Code   `json:"-"`
	SubCode int    `json:"code"`
	Message string `json:"message"`
}

//
// Error is implementation of Error interface
//
func (e Error) Error() string {
	return fmt.Sprintf("API error (code: %v sub_code: %v message: %s)", e.Code, e.SubCode, e.Message)
}

// Common API errors
var (
	ErrInternalServer = Error{
		Code:    Internal,
		SubCode: 100000,
		Message: "internal server error",
	}
	ErrServiceUnavailable = Error{
		Code:    Unavailable,
		SubCode: 100001,
		Message: "service unavailable",
	}
	ErrEntityNotFound = Error{
		Code:    NotFound,
		SubCode: 100002,
		Message: "entity not found",
	}

	ErrDecodeRequest = Error{
		Code:    InvalidArgument,
		SubCode: 100100,
		Message: "invalid request format",
	}
	ErrUnsupportedCodec = Error{
		Code:    InvalidArgument,
		SubCode: 100101,
		Message: "unsupported encoded format",
	}

	ErrPermissionDenied = Error{
		Code:    PermissionDenied,
		SubCode: 100200,
		Message: "sender doesn't have permission to execute the operation",
	}

	ErrRequestUnauthenticated = Error{
		Code:    Unauthenticated,
		SubCode: 100300,
		Message: "authorization token is invalid",
	}
)
