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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"net/http"
	"strings"
)

// InsertForm is the form for inserting images into the system. Requirement of AuthToken is configurable.
type InsertForm struct {
	Image       string `json:"image"`
	ImageName   string `json:"image-name"`
	ImageFormat string `json:"image-format"`
	Client      string `json:"client-name"`
	Username    string `json:"username"`
	AuthToken   string `json:"auth-token"`
	Hidden      bool   `json:"hidden"`
}

// Insert handles insert requests
func Insert(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var ifr InsertForm
	// Decode the payload.
	err := decoder.Decode(&ifr)
	// Check if there was an error decoding.
	if err != nil || len(ifr.Image) == 0 {
		log.Debugf("%[1]s sent an invalid insert request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fill out all non-necessary unfilled values.
	if len(ifr.ImageName) == 0 {
		ifr.ImageName = ImageName(5)
	}
	if len(ifr.ImageFormat) == 0 {
		ifr.ImageFormat = "png"
	}
	if len(ifr.Client) == 0 {
		ifr.Client = "Unknown Client"
	}

	if len(ifr.Username) == 0 || len(ifr.AuthToken) == 0 {
		// Username or authentication token not supplied.
		if config.RequireAuth {
			// The user is not logged in, but the config is set to require authentication, send error.
			log.Debugf("%[1]s tried to upload an image without authentication, even though authentication is required.", ip)
			output(w, GenericResponse{
				Success:        false,
				Status:         "not-logged-in",
				StatusReadable: "This MIS server requires authentication. Please log in or register.",
			}, http.StatusUnauthorized)
			return
		}
		// The user is not logged in, but login is not required, set username to "anonymous"
		ifr.Username = "anonymous"
	} else {
		// Username and authentication token supplied, check them.
		err = auth.CheckAuthToken(ifr.Username, []byte(ifr.AuthToken))
		if err != nil {
			log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, ifr.Username)
			output(w, GenericResponse{
				Success:        false,
				Status:         "invalid-authtoken",
				StatusReadable: "Your authentication token was incorrect. Please try logging in again.",
			}, http.StatusUnauthorized)
			return
		}
	}

	// If the image already exists, make sure that the uploader is the owner of the image.
	var replace = false
	owner := database.GetOwner(ifr.ImageName)
	if len(owner) > 0 {
		if owner != ifr.Username || ifr.Username == "anonymous" {
			output(w, GenericResponse{
				Success:        false,
				Status:         "already-exists",
				StatusReadable: "The requested image name is already in use by another user",
			}, http.StatusForbidden)
			log.Debugf("%[1]s@%[2]s attempted to override an image uploaded by %[3]s.", ifr.Username, ip, owner)
			return
		}
		replace = true
	}

	// Decode the base64 image from the JSON request.
	image, err := base64.StdEncoding.DecodeString(ifr.Image)
	if err != nil {
		output(w, GenericResponse{Success: false, Status: "invalid-image-encoding",
			StatusReadable: "The given image is not properly encoded in base64."}, http.StatusUnsupportedMediaType)
		log.Errorf("Error while decoding image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		return
	}

	mimeType := http.DetectContentType(image)

	if !strings.HasPrefix(mimeType, "image/") {
		log.Debugf("%[1]s@%[2]s attempted to upload an image with an incorrect MIME type.", ifr.Username, ip, owner)
		output(w, GenericResponse{
			Success:        false,
			Status:         "invalid-mime",
			StatusReadable: "The uploaded data is of an incorrect MIME type.",
		}, http.StatusUnsupportedMediaType)
		return
	}
	mimeType = mimeType[len("image/"):]

	// Write the image to disk.
	err = ioutil.WriteFile(config.ImageLocation+"/"+ifr.ImageName+"."+ifr.ImageFormat, image, 0644)
	if err != nil {
		log.Errorf("Error while saving image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !replace {
		// The image name has not been used. Insert it into the database.
		err = database.Insert(ifr.ImageName, ifr.ImageFormat, mimeType, ifr.Username, ip, ifr.Client, ifr.Hidden)
		if err != nil {
			log.Errorf("Error while inserting image from %[1]s@%[2]s into the database: %[3]s", ifr.Username, ip, err)
			output(w, GenericResponse{
				Success:        false,
				Status:         "database-error",
				StatusReadable: "An internal server error occurred while attempting to save image information to the database.",
			}, http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf("%[1]s@%[2]s successfully uploaded an image with the name %[3]s (new).", ifr.Username, ip, ifr.ImageName)
		output(w, GenericResponse{
			Success:        true,
			Status:         "created",
			StatusReadable: "The image was successfully saved with the name " + ifr.ImageName,
		}, http.StatusCreated)
	} else {
		// The image name was in use. Update the data in the database.
		err = database.Update(ifr.ImageName, ifr.ImageFormat, mimeType, ip, ifr.Client, ifr.Hidden)
		if err != nil {
			log.Errorf("Error while updating data of image from %[1]s@%[2]s into the database: %[3]s", ifr.Username, ip, err)
			output(w, GenericResponse{
				Success:        false,
				Status:         "database-error",
				StatusReadable: "An internal server error occurred while attempting to save image information to the database.",
			}, http.StatusInternalServerError)
			return
		}
		log.Debugf("%[1]s@%[2]s successfully uploaded an image with the name %[3]s (replaced).", ifr.Username, ip, ifr.ImageName)
		output(w, GenericResponse{
			Success: true,
			Status:  "replaced",
			StatusReadable: "The image was successfully saved with the name " + ifr.ImageName +
				", replacing your previous image with the same name",
		}, http.StatusAccepted)
	}
}
