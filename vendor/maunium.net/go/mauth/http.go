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
	"encoding/json"
	"io"
	"net/http"
)

// AuthForm is the generic authentication form which can be used for both, logging in and registering.
type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse is the generic authentication response which can be used for both, logging in and registering.
type AuthResponse struct {
	AuthToken     string `json:"auth-token,omitempty"`
	Error         string `json:"error-simple,omitempty"`
	ErrorReadable string `json:"error-humanreadable,omitempty"`
}

func decoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}

func output(w http.ResponseWriter, response interface{}, status int) bool {
	json, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"error-simple\": \"marshalerror\",\"error-humanreadable\": \"The server failed to marshal the output JSON.\"}"))
		return false
	}
	w.WriteHeader(status)
	w.Write(json)
	return true
}
