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
// If the login/register was successful, the error field should not exist. In other cases, the auth token field
// should not exist.
type AuthResponse struct {
	AuthToken string `json:"auth-token"`
}

func login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var af AuthForm
	err := decoder.Decode(&af)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	authToken, err := data.Login(af.Username, []byte(af.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	json, err := json.Marshal(AuthResponse{AuthToken: authToken})
	if err != nil {
		log.Errorf("Failed to marshal output json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func register(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement register
}
