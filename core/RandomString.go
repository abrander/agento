package core

import (
	"crypto/rand"
	"io"
)

// RandomString will generate a pseudo-random string consisting of length
// alphanumeric runes. Do not use this for anything security related.
func RandomString(length int) string {
	if length == 0 {
		return ""
	}

	var vocabulary = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	vlen := byte(len(vocabulary))
	maxrb := byte(256 - (256 % len(vocabulary)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, r); err != nil {
			panic("error reading from random source: " + err.Error())
		}
		for _, c := range r {
			if c >= maxrb {
				// Skip to avoid modulo bias.
				continue
			}
			b[i] = vocabulary[c%vlen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
