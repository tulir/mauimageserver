package main

import (
	log "maunium.net/go/maulogger"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	errdata, err := auth.LoginHTTP(w, r)
	if err != nil {
		switch err.Error() {
		case "illegalmethod":
			log.Debugf("%[1]s tried to send a login request using HTTP %[2]s", ip, errdata)
		case "invalidrequest":
			log.Debugf("%[1]s sent an invalid login request.", ip)
		case "incorrectpassword":
			log.Debugf("%[1]s tried to log in as %[2]s with the incorrect password.", ip, errdata)
		default:
			log.Errorf("Login error: %s", err)
		}
	} else {
		log.Debugf("%[1]s logged in as %[2]s successfully.", ip, errdata)
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	var ip = getIP(r)
	errdata, err := auth.RegisterHTTP(w, r)
	if err != nil {
		switch err.Error() {
		case "illegalmethod":
			log.Debugf("%[1]s tried to send a register request using HTTP %[2]s", ip, errdata)
		case "invalidrequest":
			log.Debugf("%[1]s sent an invalid register request.", ip)
		case "userexists":
			log.Debugf("%[1]s tried to register the name %[2]s, but it is already in use.", ip, errdata)
		case "invalidname":
			log.Debugf("%[1]s tried to register a name with illegal characters.", ip)
		default:
			log.Errorf("Register error: %s", err)
		}
	} else {
		log.Debugf("%[1]s registered as %[2]s successfully.", ip, errdata)
	}
}
