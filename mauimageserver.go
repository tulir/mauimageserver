package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const version = "2.0.0-B1"

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "./config.json", "The path of the mauImageServer configuration file.")
var disableSafeShutdown = flag.Bool("no-safe-shutdown", false, "Disable Interrupt/SIGTERM catching and handling.")

var config *data.Configuration
var favicon []byte

func init() {
	flag.Parse()

	if !*disableSafeShutdown {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			os.Stdout.Write([]byte("\n"))
			log.Infof("Shutting down mauImageServer...")
			data.UnloadDatabase()
			os.Exit(0)
		}()
	}
}

func main() {
	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }
	// Initialize the logger
	log.Init()

	log.Infof("Initializing mauImageServer " + version)
	loadConfig()
	loadDatabase()
	loadTemplates()

	log.Debugln("Loading favicon...")
	favicon, _ = ioutil.ReadFile(config.Favicon)

	log.Infof("Registering handlers")
	http.HandleFunc("/auth/login", login)
	http.HandleFunc("/auth/register", register)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/", get)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
