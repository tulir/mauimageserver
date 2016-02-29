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

// Package handlers contains the MIS-specific HTTP request handlers
package handlers

import (
	"encoding/json"
	"fmt"
	log "maunium.net/go/maulogger"
	"net/http"
	"time"
)

// SearchForm is the form for searching for images.
type SearchForm struct {
	Format    string `json:"image-format"`
	Adder     string `json:"adder"`
	Client    string `json:"client-name"`
	MinTime   int64  `json:"uploaded-after"`
	MaxTime   int64  `json:"uploaded-before"`
	AuthToken string `json:"auth-token"`
}

// String turns a SearchForm into a string
func (sf SearchForm) String() string {
	return fmt.Sprintf("<%[1]s|%[2]s|%[3]s|%[4]d|%[5]d>", sf.Format, sf.Adder, sf.Client, sf.MinTime, sf.MaxTime)
}

// Search handles search requests
func Search(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !config.AllowSearch {
		log.Warnf("%[1]s attempted to execute a search, even though it's not allowed", ip)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var sf SearchForm
	// Decode the payload.
	err := decoder.Decode(&sf)
	// Check if there was an error decoding.
	if err != nil || (len(sf.Format) == 0 && len(sf.Adder) == 0 && len(sf.Client) == 0 && sf.MinTime <= 0 && sf.MaxTime <= 0) {
		log.Debugf("%[1]s sent an invalid search request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if sf.MinTime <= 0 && sf.MaxTime > 0 {
		sf.MinTime = 1
	} else if sf.MinTime > 0 && sf.MaxTime <= 0 {
		sf.MaxTime = time.Now().Unix()
	}

	var authenticated = false
	if len(sf.AuthToken) != 0 {
		err = auth.CheckAuthToken(sf.Adder, []byte(sf.AuthToken))
		if err != nil {
			log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, sf.Adder)
			output(w, GenericResponse{
				Success:        false,
				Status:         "invalid-authtoken",
				StatusReadable: "The authentication token was incorrect. Please try logging in again.",
			}, http.StatusUnauthorized)
			return
		}
		authenticated = true
	}

	results, err := database.Search(sf.Format, sf.Adder, sf.Client, sf.MinTime, sf.MaxTime, authenticated)
	if err != nil {
		log.Errorf("Failed to execute search %[2]s by %[1]s: %[3]s", ip, sf.String(), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("%[1]s executed a search: %s", ip, sf.String())
	output(w, results, http.StatusOK)
}
