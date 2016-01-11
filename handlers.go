package main

import (
	"net/http"
)

// InsertForm is the form for inserting images into the system. Requirement of AuthToken is configurable.
type InsertForm struct {
	Image       string `json:"image"`
	RequestPath string `json:"request-path"`
	AuthToken   string `json:"auth-token"`
}

// DeleteForm is the form for deleting images. AuthToken is required.
type DeleteForm struct {
	ImagePath string `json:"image-path"`
	AuthToken string `json:"auth-token"`
}

func get(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement getting images. GET requests only.
}

func insert(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement inserting images. POST requests only.
}

func delete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement deleting images. POST requests only.
}
