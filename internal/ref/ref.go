package ref

import (
	"crypto/rand"
)

const (
	Product = "PDT"
	Cart    = "BSK"
	Order   = "CMD"
)

// 32 chars, skip I : 1 | O : 0 to avoid missinterpretation
const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// Test with 5 to see test fail | docker compose exec go go test -v ./internal/ref/
const length = 6

func New(prefix string) string {
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)

	code := make([]byte, length)
	for i, randomByte := range randomBytes {
		alphabetI := randomByte % byte(len(alphabet))
		code[i] = alphabet[alphabetI]
	}
	return prefix + "-" + string(code)
}
