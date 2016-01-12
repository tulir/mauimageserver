package random

import (
	crand "crypto/rand"
	"encoding/base64"
	mrand "math/rand"
	"time"
)

// AuthToken generates 32 cryptographically random bytes, encodes them with base64 and returns the base64 string.
func AuthToken() string {
	// Create a byte array.
	b := make([]byte, 32)
	// Fill it with cryptographically random bytes.
	n, err := crand.Read(b)
	// Check if there was an error.
	if n == len(b) && err == nil {
		// Encode the bytes with base64 and return.
		return base64.RawStdEncoding.EncodeToString(b)
	}
	// There was an error, return an empty string.
	return ""
}

const imageNameAC = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = mrand.NewSource(time.Now().UnixNano())

// ImageName generates a string matching [a-zA-Z0-9]{5}
func ImageName() string {
	b := make([]byte, 5)
	for i, cache, remain := 4, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(imageNameAC) {
			b[i] = imageNameAC[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
