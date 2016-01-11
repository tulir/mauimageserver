package main

import (
	"net/http"
)

// AuthForm is the generic authentication form which can be used for both, logging in and registering.
type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement login
}

func register(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement register
}
