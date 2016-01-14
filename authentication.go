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
	AuthToken     string `json:"auth-token"`
	Error         string `json:"error-simple"`
	ErrorReadable string `json:"error-humanreadable"`
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
	log.Debugf("%[1]s logged in as %[2]s successfully.", ip, af.Username)
	if !output(w, AuthResponse{AuthToken: authToken}, http.StatusOK) {
		log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, af.Username, err)
	}
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
	if invalidName(af.Username) {
		log.Debugf("%[1]s tried to register with an invalid name (%[2]s)", ip, af.Username)
		if !output(w, AuthResponse{Error: "invalidname", ErrorReadable: "The name you entered is invalid. Allowed names: [a-zA-Z0-9_-]{3,16}"}, http.StatusNotAcceptable) {
			log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, af.Username, err)
		}
		return
	}
	// Try to register
	authToken, err := data.Register(af.Username, []byte(af.Password))
	// Check if there was an error logging in.
	if err != nil {
		// Error detected.
		if err.Error() == "userexists" {
			// Username already in use.
			log.Debugf("%[1]s tried to register the name %[2]s, but it is already in use.", ip, af.Username)
			if !output(w, AuthResponse{Error: "userexists", ErrorReadable: "The given username is already in use."}, http.StatusNotAcceptable) {
				log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, af.Username, err)
			}
		} else {
			// Other error. Write internal server error status.
			log.Errorf("Register error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	log.Debugf("%[1]s registered as %[2]s successfully.", ip, af.Username)
	if !output(w, AuthResponse{AuthToken: authToken}, http.StatusOK) {
		log.Errorf("Failed to marshal output json to %[1]s@%[2]s: %[3]s", ip, af.Username, err)
	}
}

func invalidName(name string) bool {
	if len(name) < 3 || len(name) > 16 {
		return true
	}
	for char := range name {
		if (char >= 48 && char <= 57) || (char >= 65 && char <= 90) || (char >= 97 && char <= 122) || char == 95 || char == 45 {
			continue
		}
		return true
	}
	return false
}
