package randomgenerator

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/tochka/core-kit/errors"
)

// NewRandom return new instance N bytes random generator
func NewRandom(n uint8) Random {
	return Random{n}
}

// Random is N bytes random generator
type Random struct {
	n uint8
}

// Generate return UUID
func (r Random) Generate() (id string, err error) {
	b := make([]byte, r.n)
	_, err = rand.Read(b)
	if err != nil {
		return "", errors.Wrap(err,
			"component", "generator", "type", "random", "method", "generate", "size", strconv.Itoa(int(r.n)))
	}
	return hex.EncodeToString(b), nil
}
