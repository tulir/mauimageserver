package main

import (
	"encoding/json"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
)

// AuthForm is the generic authentication form which can be used for both, logging in and registering.
type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse is the generic authentication response which can be used for both, logging in and registering.
type AuthResponse struct {
	AuthToken string `json:"auth-token"`
}

func login(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var af AuthForm
	// Decode the payload.
	err := decoder.Decode(&af)
	// Check if there was an error decoding.
	if err != nil || len(af.Password) == 0 || len(af.Username) == 0 {
		log.Debugf("%[1]s sent an invalid login request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Try to login
	authToken, err := data.Login(af.Username, []byte(af.Password))
	// Check if there was an error logging in.
	if err != nil {
		// Error detected.
		if err.Error() == "incorrectpassword" {
			log.Debugf("%[1]s tried to log in as %[2]s with the incorrect password.", ip, af.Username)
			// Incorrect password. Write unauthorized status.
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Errorf("Login error: %s", err)
			// Other error. Write internal server error status.
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Marshal the response
	json, err := json.Marshal(AuthResponse{AuthToken: authToken})
	// Check if there was an error marshaling the response.
	if err != nil {
		// Error detected. Log it.
		log.Errorf("Failed to marshal output json to %[1]s (%[2]s): %[3]s", ip, af.Username, err)
		// Write internal serevr error status.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debugf("%[1]s logged in as %[2]s successfully.", ip, af.Username)
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func register(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Create a json decoder for the payload.
	decoder := json.NewDecoder(r.Body)
	var af AuthForm
	// Decode the payload.
	err := decoder.Decode(&af)
	// Check if there was an error decoding.
	if err != nil || len(af.Password) == 0 || len(af.Username) == 0 {
		log.Debugf("%[1]s sent an invalid register request.", ip)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Try to register
	authToken, err := data.Register(af.Username, []byte(af.Password))
	// Check if there was an error logging in.
	if err != nil {
		// Error detected.
		if err.Error() == "userexists" {
			log.Debugf("%[1]s tried to register the name %[2]s, but it is already in use.", ip, af.Username)
			// Username in use. Write not acceptable status.
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			log.Errorf("Register error: %s", err)
			// Other error. Write internal server error status.
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Marshal the response
	json, err := json.Marshal(AuthResponse{AuthToken: authToken})
	// Check if there was an error marshaling the response.
	if err != nil {
		// Error detected. Log it.
		log.Errorf("Failed to marshal output json to %[1]s (%[2]s): %[3]s", ip, af.Username, err)
		// Write internal serevr error status.
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Debugf("%[1]s registered and logged in as %[2]s successfully.", ip, af.Username)
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
