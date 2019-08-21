package errors

import (
	"bytes"

	"github.com/go-logfmt/logfmt"
)

func Defer(err *error, ctx *ErrorContext) {
	if err != nil && *err != nil {
		*err = Wrap(*err, ctx.Data...)
	}
}

func Context(data ...interface{}) *ErrorContext {
	return &ErrorContext{Data: data}
}

type ErrorContext struct {
	Data []interface{}
}

func (ctx *ErrorContext) Add(data ...interface{}) {
	ctx.Data = append(ctx.Data, data...)
}

type ContextError struct {
	innerError error
	keyvals    []interface{}
}

func (err *ContextError) Error() string {
	b := bytes.NewBuffer(nil)
	enc := logfmt.NewEncoder(b)

	enc.EncodeKeyvals(err.keyvals...)
	enc.EncodeKeyval("error", err.innerError)

	return b.String()
}

func (err *ContextError) Unwrap() error {
	return err.innerError
}
