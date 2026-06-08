package utility

import (
	"crypto/sha256"
)

func HashSHA256(inp []byte) []byte {
	defer WipeBytes(inp)

	hash := sha256.New()
	hash.Write(inp)
	return hash.Sum(nil)
}

func HashSHA256PreserveInput(inp []byte) []byte {
	hash := sha256.New()
	hash.Write(inp)
	return hash.Sum(nil)
}
