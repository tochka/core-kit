package errors

import "fmt"

func NotUnwrap(err error) error {
	return NotUnwrapError{err}
}

type NotUnwrapError struct {
	err error
}

func (err NotUnwrapError) Error() string {
	return fmt.Sprintf("NotUnwrapError(error: %v)", err.err)
}
