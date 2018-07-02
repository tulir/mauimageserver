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
	"database/sql"
	"net/http"
)

// System is an instance of Maunium Authentication.
type System interface {
	// Register creates an account and generates an authentication token for it.
	Register(username string, password []byte) (string, error)
	// RegisterHTTP handles a HTTP register request.
	RegisterHTTP(w http.ResponseWriter, r *http.Request) (string, error)
	// RegisterHTTPD calls RegisterHTTP, but doesn't return anything. Best suited for use with http.HandleFunc
	RegisterHTTPD(w http.ResponseWriter, r *http.Request)

	// Login generates an authentication token for the user.
	Login(username string, password []byte) (string, error)
	// LoginHTTP handles a HTTP login request.
	LoginHTTP(w http.ResponseWriter, r *http.Request) (string, error)
	// LoginHTTPD calls LoginHTTP, but doesn't return anything. Best suited for use with http.HandleFunc
	LoginHTTPD(w http.ResponseWriter, r *http.Request)

	// CheckAuthToken checks if the given auth token is valid for the given user.
	CheckAuthToken(username string, authtoken []byte) error
}

type isystem struct {
	db *sql.DB
}

// Create a System.
func Create(database *sql.DB) (System, error) {
	_, err := database.Exec("CREATE TABLE IF NOT EXISTS users (username VARCHAR(16) PRIMARY KEY, password BINARY(60) NOT NULL, authtoken BINARY(60));")
	if err != nil {
		return isystem{}, err
	}
	return isystem{database}, nil
}

func validName(name string) bool {
	if len(name) < 3 || len(name) > 16 {
		return false
	}
	for _, char := range name {
		if (char >= 48 && char <= 57) || (char >= 65 && char <= 90) || (char >= 97 && char <= 122) || char == 95 || char == 45 {
			continue
		}
		return false
	}
	return true
}
