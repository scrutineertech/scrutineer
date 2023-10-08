package util

import (
	"crypto/rand"
	"math/big"
)

// OrString returns the first non-empty string.
func OrString(o ...string) string {
	for _, s := range o {
		if s != "" {
			return s
		}
	}

	return ""
}

// GenToken generates a random token of length based on the base58 alphabet without lower case chars.
func GenToken(length int) string {
	if length < 1 {
		length = 6
	}

	// base58 alphabet to avoid character confusion without lower case chars
	alphabet := "123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	myReader := rand.Reader

	token := ""
	for i := 0; i < length; i++ {
		letterAt, _ := rand.Int(myReader, big.NewInt(int64(len(alphabet))))
		token += string(alphabet[letterAt.Uint64()])
	}

	return token
}
