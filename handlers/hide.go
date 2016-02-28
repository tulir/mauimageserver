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
	log "maunium.net/go/maulogger"
	"net/http"
)

// HideForm is the form for hiding/unhiding images. AuthToken is required.
type HideForm struct {
	ImageName string `json:"image-name"`
	Hidden    bool   `json:"hidden"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
}

// Hide handles hide/unhide requests
func Hide(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var hfr HideForm
	// Decode the payload.
	err := decoder.Decode(&hfr)
	// Check if there was an error decoding.
	if err != nil || len(hfr.ImageName) == 0 || len(hfr.Username) == 0 || len(hfr.AuthToken) == 0 {
		log.Debugf("%[1]s sent an invalid hide request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = auth.CheckAuthToken(hfr.Username, []byte(hfr.AuthToken))
	// Check if the auth token was correct
	if err != nil {
		log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, hfr.Username)
		output(w, InsertResponse{
			Success:        false,
			Status:         "invalid-authtoken",
			StatusReadable: "The authentication token was incorrect. Please try logging in again.",
		}, http.StatusUnauthorized)
		return
	}

	owner := database.GetOwner(hfr.ImageName)
	if len(owner) > 0 {
		if owner != hfr.Username {
			log.Debugf("%[1]s@%[2]s attempted to hide an image uploaded by %[3]s.", hfr.Username, ip, owner)
			output(w, InsertResponse{
				Success:        false,
				Status:         "no-permissions",
				StatusReadable: "The image you requested to be deleted was not uploaded by you.",
			}, http.StatusForbidden)
			return
		}
	} else {
		log.Debugf("%[1]s@%[2]s attempted to hide an image that doesn't exist.", hfr.Username, ip, owner)
		output(w, InsertResponse{Success: false, Status: "not-found", StatusReadable: "The image you requested to be deleted does not exist."}, http.StatusNotFound)
		return
	}

	err = database.SetHidden(hfr.ImageName, hfr.Hidden)
	if err != nil {
		log.Warnf("Error changing hide status of %[4]s (requested by %[1]s@%[2]s): %[3]s", hfr.Username, ip, err, hfr.ImageName)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var hid string
	if hfr.Hidden {
		hid = "hidden"
	} else {
		hid = "unhidden"
	}

	log.Debugf("%[1]s@%[2]s successfully changed hidden status to %[4]t of the image with the name %[3]s.", hfr.Username, ip, hfr.ImageName, hfr.Hidden)
	output(w, InsertResponse{
		Success:        true,
		Status:         hid,
		StatusReadable: "The image " + hfr.ImageName + " was successfully " + hid + ".",
	}, http.StatusAccepted)
}
