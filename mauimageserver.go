package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"strconv"
)

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "./config.json", "The path of the mauImageServer configuration file.")

var config *data.Configuration

func main() {
	flag.Parse()

	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }
	// Initialize the logger
	log.Init()

	log.Infof("Initializing mauImageServer")
	loadConfig()
	loadDatabase()

	log.Infof("Registering handlers")
	http.HandleFunc("/auth/login", login)
	http.HandleFunc("/auth/register", register)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/", get)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
