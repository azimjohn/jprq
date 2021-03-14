package jprq

import (
	cryptorand "crypto/rand"
	"fmt"
)

func generateToken() string {
	b := make([]byte, 32)
	cryptorand.Read(b)
	return fmt.Sprintf("%x", b)
}
