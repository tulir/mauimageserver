package main

import (
	"encoding/json"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"maunium.net/go/mauth"
	"net/http"
	"os"
	"strings"
)

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func output(w http.ResponseWriter, response interface{}, status int) bool {
	// Marshal the response
	json, err := json.Marshal(response)
	// Check if there was an error marshaling the response.
	if err != nil {
		// Write internal server error status.
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	w.WriteHeader(status)
	w.Write(json)
	return true
}

func loadConfig() {
	log.Infof("Loading config...")
	var err error
	config, err = data.LoadConfig(*confPath)
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugln("Successfully loaded config.")
}

func loadDatabase() {
	log.Infof("Loading database...")

	database, err := data.LoadDatabase(config.SQL)
	if err != nil {
		log.Fatalf("Failed to load database: %[1]s", err)
		os.Exit(2)
	}

	auth, err = mauth.Create(database)
	if err != nil {
		log.Fatalf("Failed to load Mauth: %[1]s", err)
	}

	log.Debugln("Successfully loaded database.")
}

func loadTemplates() {
	log.Infof("Loading HTML templates...")
	err := data.LoadTemplates()
	if err != nil {
		log.Fatalf("Failed to load image page: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML templates")
}
