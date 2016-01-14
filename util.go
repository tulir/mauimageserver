package main

import (
	"encoding/json"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
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
	log.Infoln("Loading config...")
	var err error
	config, err = data.LoadConfig(*confPath)
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugln("Successfully loaded config.")
}

func loadDatabase() {
	log.Infoln("Loading database...")

	var err error
	err = data.LoadDatabase(config.SQL)
	if err != nil {
		log.Fatalf("Failed to load database: %[1]s", err)
		os.Exit(2)
	}

	log.Debugln("Successfully loaded database.")
}
