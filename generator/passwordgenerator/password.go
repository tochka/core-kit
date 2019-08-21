package passwordgenerator

import (
	"crypto/rand"
)

var (
	defaultSource = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
)

//
// New return new instance for password generator
//
func New(n uint8) *Generator {
	return &Generator{int(n), defaultSource}
}

//
// PasswordGenerator provide generator for password
//
type Generator struct {
	n      int
	source []byte
}

//
// Generate return random password string
//
func (g *Generator) Generate() (string, error) {
	return randomChars(g.n, g.source)
}

func randomChars(length int, chars []byte) (string, error) {
	result := make([]byte, length)
	randomData := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := rand.Read(randomData); err != nil {
			return "", err
		}
		for _, c := range randomData {
			if maxrb != 0x00 && c >= maxrb {
				continue
			}
			result[i] = chars[c%clen]
			i++
			if i == length {
				return string(result), nil
			}
		}
	}
}
