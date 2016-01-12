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
	// TODO: Implement getting images. GET requests only.
}

func insert(w http.ResponseWriter, r *http.Request) {
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

	var image []byte
	base64.StdEncoding.Decode(image, []byte(ifr.Image))
	ioutil.WriteFile(config.ImageLocation+"/"+path+".png", image, 0644)

}

func delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement deleting images. POST requests only.
}
