package main

import (
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
	Error     string `json:"error"`
}

func login(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement login
}

func register(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement register
}
