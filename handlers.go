package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"maunium.net/go/mauimageserver/data"
	"maunium.net/go/mauimageserver/random"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// InsertForm is the form for inserting images into the system. Requirement of AuthToken is configurable.
type InsertForm struct {
	Image     string `json:"image"`
	ImageName string `json:"image-name"`
	Client    string `json:"client-name"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
}

// DeleteForm is the form for deleting images. AuthToken is required.
type DeleteForm struct {
	ImageName string `json:"image-name"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
}

// InsertResponse is the response for an insert call.
type InsertResponse struct {
	Success        bool   `json:"success"`
	Status         string `json:"status-simple"`
	StatusReadable string `json:"status-humanreadable"`
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path[1:]

	if path == "" || path == "index.html" || path == "index.php" || path == "index" || path == "index.htm" {
		// TODO: Index page?
	} else if path == "favicon.ico" {
		w.WriteHeader(http.StatusFound)
		w.Write(favicon)
		return
	}

	adder, _, client, timestamp, index, err := data.Query(path)
	if err == nil {
		date := time.Unix(timestamp, 0).Format(config.DateFormat)
		r.URL.Path = r.URL.Path + ".png"
		image.Execute(w, imageTemplate{
			ImageName: path,
			ImageAddr: r.URL.String(),
			Uploader:  adder,
			Client:    client,
			Date:      date,
			Index:     strconv.Itoa(index),
		})
		return
	}

	data, err := ioutil.ReadFile(config.ImageLocation + r.URL.Path)
	if err != nil {
		log.Errorf("%[1]s tried to get the non-existent image %[2]s", getIP(r), path)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusFound)
	w.Write(data)
}

func insert(w http.ResponseWriter, r *http.Request) {
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

	imageName := ifr.ImageName
	if len(imageName) == 0 {
		imageName = random.ImageName(5)
	}

	if len(ifr.Username) == 0 || len(ifr.AuthToken) == 0 {
		if config.RequireAuth {
			if !output(w, InsertResponse{
				Success:        false,
				Status:         "not-logged-in",
				StatusReadable: "This MIS server requires authentication. Please log in or register.",
			}, http.StatusUnauthorized) {
				log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, ifr.Username, err)
			}
		} else {
			ifr.Username = "anonymous"
		}
	} else {
		err = data.CheckAuthToken(ifr.Username, []byte(ifr.AuthToken))
		// Check if the auth token was correct
		if err != nil {
			log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, ifr.Username)
			if !output(w, InsertResponse{
				Success:        false,
				Status:         "invalid-authtoken",
				StatusReadable: "Your authentication token was incorrect. Please try logging in again.",
			}, http.StatusUnauthorized) {
				log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, ifr.Username, err)
			}
			return
		}
	}

	var replace = false
	owner := data.GetOwner(imageName)
	if len(owner) > 0 {
		if owner != ifr.Username || ifr.Username == "anonymous" {
			output(w, InsertResponse{Success: false, Status: "already-exists", StatusReadable: "The requested image name is already in use by another user"}, http.StatusForbidden)
			log.Debugf("%[1]s@%[2]s attempted to override an image uploaded by %[3]s.", ifr.Username, ip, owner)
			return
		}
		replace = true
	}

	image, err := base64.StdEncoding.DecodeString(ifr.Image)
	if err != nil {
		log.Errorf("Error while decoding image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = ioutil.WriteFile(config.ImageLocation+"/"+imageName+".png", image, 0644)
	if err != nil {
		log.Errorf("Error while saving image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(ifr.Client) == 0 {
		ifr.Client = "Unknown Client"
	}

	if !replace {
		err = data.Insert(imageName, ifr.Username, ip, ifr.Client)
		if err != nil {
			log.Errorf("Error while inserting image from %[1]s@%[2]s into the database: %[3]s", ifr.Username, ip, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Debugf("%[1]s@%[2]s successfully uploaded an image with the name %[3]s (new).", ifr.Username, ip, imageName)
		if !output(w, InsertResponse{
			Success:        true,
			Status:         "created",
			StatusReadable: "The image was successfully saved with the name " + imageName,
		}, http.StatusCreated) {
			log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, ifr.Username, err)
		}
	} else {
		log.Debugf("%[1]s@%[2]s successfully uploaded an image with the name %[3]s (replaced).", ifr.Username, ip, imageName)
		if !output(w, InsertResponse{
			Success: true,
			Status:  "replaced",
			StatusReadable: "The image was successfully saved with the name " + imageName +
				", replacing your previous image with the same name",
		}, http.StatusAccepted) {
			log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, ifr.Username, err)
		}
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
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

	err = data.CheckAuthToken(dfr.Username, []byte(dfr.AuthToken))
	// Check if the auth token was correct
	if err != nil {
		log.Debugf("%[1]s tried to authenticate as %[2]s with the wrong token.", ip, dfr.Username)
		if !output(w, InsertResponse{
			Success:        false,
			Status:         "invalid-authtoken",
			StatusReadable: "The authentication token was incorrect. Please try logging in again.",
		}, http.StatusUnauthorized) {
			log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, dfr.Username, err)
		}
		return
	}

	owner := data.GetOwner(dfr.ImageName)
	if len(owner) > 0 {
		if owner != dfr.Username {
			log.Debugf("%[1]s@%[2]s attempted to delete an image uploaded by %[3]s.", dfr.Username, ip, owner)
			if !output(w, InsertResponse{Success: false, Status: "no-permissions", StatusReadable: "The image you requested to be deleted was not uploaded by you."}, http.StatusForbidden) {
				log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, dfr.Username, err)
			}
			return
		}
	} else {
		log.Debugf("%[1]s@%[2]s attempted to delete an image that doesn't exist.", dfr.Username, ip, owner)
		if !output(w, InsertResponse{Success: false, Status: "does-not-exist", StatusReadable: "The image you requested to be deleted does not exist."}, http.StatusNotFound) {
			log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, dfr.Username, err)
		}
		return
	}

	err = data.Remove(dfr.ImageName)
	if err != nil {
		log.Warnf("Error deleting %[4]s from the database (requested by %[1]s@%[2]s): %[3]s", dfr.Username, ip, err, dfr.ImageName)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = os.Remove(config.ImageLocation + "/" + dfr.ImageName + ".png")
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
	if !output(w, InsertResponse{
		Success:        true,
		Status:         "deleted",
		StatusReadable: "The image " + dfr.ImageName + " was successfully deleted.",
	}, http.StatusAccepted) {
		log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, dfr.Username, err)
	}
}
