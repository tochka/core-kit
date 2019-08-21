package errors

import (
	"errors"
	"reflect"
)

var ErrMissingValue = errors.New("(MISSING)")

func Wrap(err error, keyvals ...interface{}) error {
	if err == nil {
		return nil
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, ErrMissingValue)
	}
	return &ContextError{innerError: err, keyvals: keyvals}
}

// A Wrapper is an error implementation
// wrapping context around another error.
type Wrapper interface {
	// Unwrap returns the next error in the error chain.
	// If there is no next error, Unwrap returns nil.
	Unwrap() error
}

func Is(err, target error) bool {
	for {
		if err == target {
			return true
		}
		wrapper, ok := err.(Wrapper)
		if !ok {
			return false
		}
		err = wrapper.Unwrap()
		if err == nil {
			return false
		}
	}
}

func As(err error, expectedType interface{}) (e error, ok bool) {
	t := reflect.TypeOf(expectedType)
	for {

		if reflect.TypeOf(err) == t {
			return err, true
		}

		wrapper, ok := err.(Wrapper)
		if !ok {
			return nil, false
		}
		err = wrapper.Unwrap()
		if err == nil {
			return nil, false
		}
	}
}
