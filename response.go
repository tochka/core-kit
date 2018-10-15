package corekit

import (
	"net/http"
)

// NoContent return new instance of HTTP response with status code NoContent
func NoContent() Response {
	return Response{
		StatusCode: http.StatusNoContent,
	}
}

// Created return new instance of HTTP response with status code Created
func Created() Response {
	return Response{
		StatusCode: http.StatusCreated,
	}
}

// OK return new instance of HTTP response with status code OK
func OK(payload interface{}) Response {
	return Response{
		StatusCode: http.StatusOK,
		Payload:    payload,
	}
}
