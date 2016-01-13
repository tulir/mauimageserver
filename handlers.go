package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"maunium.net/go/mauimageserver/random"
	log "maunium.net/go/maulogger"
	"net/http"
)

// InsertForm is the form for inserting images into the system. Requirement of AuthToken is configurable.
type InsertForm struct {
	Image       string `json:"image"`
	RequestPath string `json:"request-path"`
	Username    string `json:"username"`
	AuthToken   string `json:"auth-token"`
}

// DeleteForm is the form for deleting images. AuthToken is required.
type DeleteForm struct {
	ImagePath string `json:"image-path"`
	Username  string `json:"username"`
	AuthToken string `json:"auth-token"`
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
		log.Debugf("%[1]s sent an invalid insert request.", getIP(r))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var path = ifr.RequestPath
	if len(path) == 0 {
		path = random.ImageName(5)
	}

	// TODO: Check if config.ImageLocation/path exists.

	image, err := base64.StdEncoding.DecodeString(ifr.Image)
	if err != nil {
		// TODO: Handle error
	}
	ioutil.WriteFile(config.ImageLocation+"/"+path+".png", image, 0644)

	// TODO: Insert info into database.
}

func delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// TODO: Implement deleting images. POST requests only.
}
