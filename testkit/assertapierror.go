package testkit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tochka/core-kit/apikit"
	"github.com/tochka/core-kit/errors"
)

func AssertAPIError(t *testing.T, apiErr apikit.Error, actual error) {
	ierr, ok := errors.As(actual, apiErr)
	if !ok {
		t.Fatalf("Expected API Error but actual %+v", actual)
	}

	assert.Equal(t, apiErr, ierr)
}
