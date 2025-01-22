package random

import (
	"math/rand"
	"time"
)

func NewRandomString(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetLen := len(charset)

	res := make([]byte, length)

	for i := range length {
		randIndex := rnd.Intn(charsetLen)
		res[i] = charset[randIndex]
	}

	return string(res)
}
