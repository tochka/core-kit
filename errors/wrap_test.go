package errors

import (
	"errors"
	"testing"
)

func TestIs(t *testing.T) {
	err := errors.New("default error")
	target := errors.New("default target")

	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "error is nil",
			err:      nil,
			target:   target,
			expected: false,
		},
		{
			name:     "target is nil",
			err:      err,
			target:   nil,
			expected: false,
		},
		{
			name:     "error is wrapped but not equals target",
			err:      Wrap(err, "test", "test"),
			target:   target,
			expected: false,
		},
		{
			name:     "error is wrapped and nil",
			err:      &ContextError{innerError: nil, keyvals: []interface{}{}},
			target:   target,
			expected: false,
		},
		{
			name:     "error equals target",
			err:      target,
			target:   target,
			expected: true,
		},
		{
			name:     "error is wrapped but equals target",
			err:      Wrap(target, "test", "test"),
			target:   target,
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if Is(test.err, test.target) != test.expected {
				t.Fatalf("Is function return %v but expected %v", !test.expected, test.expected)
			}
		})
	}
}

func TestIsUnwrapError(t *testing.T) {
	expectedErr := errors.New("Expected error")
	actual := NotUnwrap(expectedErr)
	if Is(actual, expectedErr) {
		t.Fatal("Unwraperror doesn't return inner error")
	}
}
