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
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func (sys isystem) Register(username string, password []byte) (string, error) {
	if !validName(username) {
		return "", fmt.Errorf("invalidname")
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashgen")
	}

	authToken, authHash := generateAuthToken()
	if len(authToken) == 0 {
		return "", fmt.Errorf("authtoken-generror")
	}

	result, err := sys.db.Query("SELECT EXISTS(SELECT 1 FROM users WHERE username=?)", username)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			// Read the data in the current row.
			var res int
			result.Scan(&res)
			if res == 1 {
				return "", fmt.Errorf("userexists")
			}
		}
	}

	_, err = sys.db.Query("INSERT INTO users VALUES(?, ?, ?)", username, hash, authHash)
	if err != nil {
		return "", fmt.Errorf("inserterror")
	}

	return authToken, nil
}

func (sys isystem) RegisterHTTPD(w http.ResponseWriter, r *http.Request) {
	sys.LoginHTTP(w, r)
}

func (sys isystem) RegisterHTTP(w http.ResponseWriter, r *http.Request) (string, error) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return r.Method, fmt.Errorf("illegalmethod")
	}
	decoder := decoder(r.Body)
	var af AuthForm
	err := decoder.Decode(&af)
	if err != nil || len(af.Password) == 0 || len(af.Username) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err != nil {
			return err.Error(), fmt.Errorf("invalidrequest")
		}
		return "", fmt.Errorf("invalidrequest")
	}
	authToken, err := sys.Register(af.Username, []byte(af.Password))
	if err != nil {
		if err.Error() == "userexists" {
			output(w, AuthResponse{Error: "userexists", ErrorReadable: "The given username is already in use."}, http.StatusNotAcceptable)
			return af.Username, fmt.Errorf("userexists")
		} else if err.Error() == "invalidname" {
			output(w, AuthResponse{Error: "invalidname", ErrorReadable: "The name you entered is invalid. Allowed names: [a-zA-Z0-9_-]{3,16}"}, http.StatusNotAcceptable)
			return af.Username, fmt.Errorf("invalidname")
		}
		w.WriteHeader(http.StatusInternalServerError)
		return "", err
	}
	output(w, AuthResponse{AuthToken: authToken}, http.StatusOK)
	return af.Username, nil
}
