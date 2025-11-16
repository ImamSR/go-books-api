package util

import (
	"crypto/rand"
	"fmt"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

// RandomID menghasilkan 21-char URL-safe (mirip nanoid) dari crypto/rand.
func RandomID() string {
	const n = 21
	b := make([]byte, n)
	if _, err := rand.Read(b); err == nil {
		for i := 0; i < n; i++ {
			b[i] = alphabet[int(b[i])%len(alphabet)]
		}
		return string(b)
	}
	// Fallback yang tetap unik (sangat kecil kemungkinan tabrakan)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
