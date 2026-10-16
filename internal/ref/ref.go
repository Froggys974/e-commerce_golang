package ref

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const (
	Product = "PDT"
	Cart    = "BSK"
	Order   = "CMD"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const length = 6

func New(prefix string) string {
	return prefix + "-" + Code(length)
}

func Code(size int) string {
	randomBytes := make([]byte, size)
	// rand read jamais en erreur depuis go 1 24
	rand.Read(randomBytes)

	code := make([]byte, size)
	for i, randomByte := range randomBytes {
		alphabetI := randomByte % byte(len(alphabet))
		code[i] = alphabet[alphabetI]
	}
	return string(code)
}

func Hash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
