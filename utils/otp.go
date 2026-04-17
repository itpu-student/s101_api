package utils

import (
	"crypto/rand"
	"math/big"
)

// NewOTP6 returns a 6-digit numeric code using crypto/rand.
func NewOTP6() string {
	out := make([]byte, 6)
	max := big.NewInt(10)
	for i := range out {
		n, _ := rand.Int(rand.Reader, max)
		out[i] = byte('0' + n.Int64())
	}
	return string(out)
}
