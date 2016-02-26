// mauImageServer - A self-hosted server to store and easily share images.
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package random is the package used for generating random image names.
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

// ImageName generates a string matching [a-zA-Z0-9]{length}
func ImageName(length int) string {
	b := make([]byte, length)
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
