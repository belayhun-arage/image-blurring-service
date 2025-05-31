package helper

import (
	"math/rand"
	"time"
)

// NUMBERS const numbers
const NUMBERS = "1234567890"

// CHARACTERS const field
const CHARACTERS = "abcdefghijelmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_1234567890"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func GenerateRandomString(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
