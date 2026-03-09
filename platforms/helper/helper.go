package helper

import (
	"math/rand"
	"sync"
)

const NUMBERS = "1234567890"
const CHARACTERS = "abcdefghijelmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_1234567890"

var mu sync.Mutex

func GenerateRandomString(length int, charset string) string {
	b := make([]byte, length)
	mu.Lock()
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	mu.Unlock()
	return string(b)
}
