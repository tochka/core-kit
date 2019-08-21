package uuid

import (
	"github.com/tochka/core-kit/errors"
	"gopkg.in/gofrs/uuid.v3"
)

// NewV4 return new instance of UUID V4
func NewV4() V4 {
	return V4{}
}

// V4 is UUID generator V4
type V4 struct {
}

// Generate return UUID
func (_ V4) Generate() (id string, err error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(err, "component", "generator", "type", "uuid_v4", "method", "generate")
	}
	return uid.String(), nil
}
