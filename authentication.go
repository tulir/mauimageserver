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
package main

import (
	log "maunium.net/go/maulogger"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	errdata, err := auth.LoginHTTP(w, r)
	if err != nil {
		switch err.Error() {
		case "illegalmethod":
			log.Debugf("%[1]s tried to send a login request using HTTP %[2]s", ip, errdata)
		case "invalidrequest":
			log.Debugf("%[1]s sent an invalid login request.", ip)
		case "incorrectpassword":
			log.Debugf("%[1]s tried to log in as %[2]s with the incorrect password.", ip, errdata)
		default:
			log.Errorf("Login error: %s", err)
		}
	} else {
		log.Debugf("%[1]s logged in as %[2]s successfully.", ip, errdata)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	errdata, err := auth.RegisterHTTP(w, r)
	if err != nil {
		switch err.Error() {
		case "illegalmethod":
			log.Debugf("%[1]s tried to send a register request using HTTP %[2]s", ip, errdata)
		case "invalidrequest":
			log.Debugf("%[1]s sent an invalid register request.", ip)
		case "userexists":
			log.Debugf("%[1]s tried to register the name %[2]s, but it is already in use.", ip, errdata)
		case "invalidname":
			log.Debugf("%[1]s tried to register a name with illegal characters.", ip)
		default:
			log.Errorf("Register error: %s", err)
		}
	} else {
		log.Debugf("%[1]s registered as %[2]s successfully.", ip, errdata)
	}
}
