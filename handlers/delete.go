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
	"os"
	"strings"
)

// DeleteForm is the form for deleting images. AuthToken is required.
type DeleteForm struct {
	ImageName string `json:"image-name"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
}

// Delete handles delete requests
func Delete(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var dfr DeleteForm
	// Decode the payload.
	err := decoder.Decode(&dfr)
	// Check if there was an error decoding.
	if err != nil || len(dfr.ImageName) == 0 || len(dfr.Username) == 0 || len(dfr.AuthToken) == 0 {
		log.Debugf("%[1]s sent an invalid delete request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = auth.CheckAuthToken(dfr.Username, []byte(dfr.AuthToken))
	// Check if the auth token was correct
	if err != nil {
		log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, dfr.Username)
		output(w, InsertResponse{
			Success:        false,
			Status:         "invalid-authtoken",
			StatusReadable: "The authentication token was incorrect. Please try logging in again.",
		}, http.StatusUnauthorized)
		return
	}

	data, err := database.Query(dfr.ImageName)
	if err != nil {
		log.Debugf("%[1]s@%[2]s attempted to delete an image that doesn't exist.", dfr.Username, ip, data.Adder)
		output(w, InsertResponse{Success: false, Status: "not-found", StatusReadable: "The image you requested to be deleted does not exist."}, http.StatusNotFound)
		return
	}
	if data.Adder != dfr.Username {
		log.Debugf("%[1]s@%[2]s attempted to delete an image uploaded by %[3]s.", dfr.Username, ip, data.Adder)
		output(w, InsertResponse{Success: false, Status: "no-permissions",
			StatusReadable: "The image you requested to be deleted was not uploaded by you."}, http.StatusForbidden)
		return
	}

	err = database.Remove(dfr.ImageName)
	if err != nil {
		log.Warnf("Error deleting %[4]s from the database (requested by %[1]s@%[2]s): %[3]s", dfr.Username, ip, err, dfr.ImageName)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = os.Remove(config.ImageLocation + "/" + dfr.ImageName + "." + data.Format)
	if err != nil {
		// If the file just didn't exist, warn about the error. If the error was something else, cancel.
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			log.Warnf("Error deleting %[3]s from the filesystem (requested by %[1]s@%[2]s): File not found", dfr.Username, ip, dfr.ImageName)
		} else {
			log.Errorf("Error deleting %[4]s from the filesystem (requested by %[1]s@%[2]s): %[3]s", dfr.Username, ip, err, dfr.ImageName)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	log.Debugf("%[1]s@%[2]s successfully deleted the image with the name %[3]s.", dfr.Username, ip, dfr.ImageName)
	output(w, InsertResponse{
		Success:        true,
		Status:         "deleted",
		StatusReadable: "The image " + dfr.ImageName + " was successfully deleted.",
	}, http.StatusAccepted)
}
