package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"maunium.net/go/mauimageserver/data"
	"maunium.net/go/mauimageserver/random"
	log "maunium.net/go/maulogger"
	"net/http"
)

// InsertForm is the form for inserting images into the system. Requirement of AuthToken is configurable.
type InsertForm struct {
	Image       string `json:"image"`
	RequestName string `json:"request-name"`
	Username    string `json:"username"`
	AuthToken   string `json:"auth-token"`
}

// DeleteForm is the form for deleting images. AuthToken is required.
type DeleteForm struct {
	ImageName string `json:"image-name"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
}

// InsertResponse is the response for an insert call.
type InsertResponse struct {
	Status         string `json:"status-simple"`
	StatusReadable string `json:"status-humanreadable"`
}

func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// TODO: Implement getting images.
	//var path = r.URL.Path[1:]
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

	imageName := ifr.RequestName
	if len(imageName) == 0 {
		imageName = random.ImageName(5)
	}

	var replace = false
	owner := data.GetOwner(imageName)
	if len(owner) > 0 {
		if owner != ifr.Username {
			output(w, InsertResponse{Status: "already-exists", StatusReadable: "The requested image path is already in use by another user"}, http.StatusForbidden)
			log.Debugf("%[1]s@%[2]s attempted to override an image uploaded by %[3]s.", ifr.Username, ip, owner)
			return
		}
		replace = true
	}

	image, err := base64.StdEncoding.DecodeString(ifr.Image)
	if err != nil {
		log.Errorf("Error while decoding image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	err = ioutil.WriteFile(config.ImageLocation+"/"+imageName+".png", image, 0644)
	if err != nil {
		log.Errorf("Error while saving image from %[1]s@%[2]s: %[3]s", ifr.Username, ip, err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if !replace {
		err = data.Insert(imageName, ifr.Username, ip)
		if err != nil {
			log.Errorf("Error while inserting image from %[1]s@%[2]s into the database: %[3]s", ifr.Username, ip, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		output(w, InsertResponse{
			Status: "replaced",
			StatusReadable: "The image was successfully saved with the name " + imageName +
				", replacing your previous image with the same name",
		}, http.StatusAccepted)
	} else {
		output(w, InsertResponse{
			Status:         "created",
			StatusReadable: "The image was successfully saved with the name " + imageName,
		}, http.StatusCreated)
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// TODO: Implement deleting images. POST requests only.
}
