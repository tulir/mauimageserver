// mauth - Maunium Authentication System for Golang.
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

// Package mauth is the main package for the Maunium Authentication System.
package mauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func (sys isystem) CheckAuthToken(username string, authtoken []byte) error {
	result, err := sys.db.Query("SELECT authtoken FROM users WHERE username=?;", username)
	if err == nil {
		defer result.Close()
		for result.Next() {
			if result.Err() != nil {
				break
			}
			// Read the data in the current row.
			var hash []byte
			result.Scan(&hash)

			if len(hash) != 0 {
				err = bcrypt.CompareHashAndPassword(hash, authtoken)
				if err != nil {
					return fmt.Errorf("invalid-authtoken")
				}
				return nil
			}
		}
	}
	return fmt.Errorf("invalid-authtoken")
}

func generateAuthToken() (string, []byte) {
	var authToken string
	b := make([]byte, 32)
	// Fill the byte array with cryptographically random bytes.
	n, err := rand.Read(b)
	if n == len(b) && err == nil {
		authToken = base64.RawStdEncoding.EncodeToString(b)
		if authToken == "" {
			return "", nil
		}
	}

	authHash, err := bcrypt.GenerateFromPassword([]byte(authToken), bcrypt.DefaultCost-3)
	if err != nil {
		return "", nil
	}
	return authToken, authHash
}
