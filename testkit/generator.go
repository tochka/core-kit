package testkit

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateString(len int) string {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:len]
}
