package random

import (
	crand "crypto/rand"
	"encoding/base64"
	mrand "math/rand"
	"time"
)

func authToken() string {
	b := make([]byte, 32)
	n, err := crand.Read(b)
	if n == len(b) && err == nil {
		return base64.RawStdEncoding.EncodeToString(b)
	}
	return ""
}

const imageNameAC = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = mrand.NewSource(time.Now().UnixNano())

func imageName() string {
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
