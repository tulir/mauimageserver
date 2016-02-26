// mauImageServer - A self-hosted server to store and easily share images.
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"maunium.net/go/mauth"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const version = "2.0.0-B3"

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "./config.json", "The path of the mauImageServer configuration file.")
var disableSafeShutdown = flag.Bool("no-safe-shutdown", false, "Disable Interrupt/SIGTERM catching and handling.")

var config *data.Configuration
var auth mauth.System

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

	log.Infof("Registering handlers")
	http.HandleFunc("/auth/login", login)
	http.HandleFunc("/auth/register", register)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/hide", hide)
	http.HandleFunc("/search", search)
	http.HandleFunc("/", get)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
