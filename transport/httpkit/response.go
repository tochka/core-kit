package httpkit

import "net/http"

type Response struct {
	Object     interface{}
	StatusCode int
	Header     http.Header
}

func OK(obj interface{}) *Response {
	return &Response{
		StatusCode: http.StatusOK,
		Object:     obj,
		Header:     make(http.Header),
	}
}

func Created(obj interface{}) *Response {
	return &Response{
		StatusCode: http.StatusCreated,
		Object:     obj,
		Header:     make(http.Header),
	}
}

func NoContent() *Response {
	return &Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
	}
}
