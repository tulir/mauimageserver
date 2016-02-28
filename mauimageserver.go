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
	"maunium.net/go/mauimageserver/handlers"
	log "maunium.net/go/maulogger"
	"maunium.net/go/mauth"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const version = "2.0.0"

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "/etc/mis2/config.json", "The path of the mauImageServer configuration file.")
var disableSafeShutdown = flag.Bool("no-safe-shutdown", false, "Disable Interrupt/SIGTERM catching and handling.")

var config *data.Configuration
var database data.MISDatabase
var auth mauth.System

func main() {
	flag.Parse()

	if !*disableSafeShutdown {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			os.Stdout.Write([]byte("\n"))
			log.Infof("Shutting down mauImageServer...")
			database.Unload()
			os.Exit(0)
		}()
	}

	// Configure the logger
	if *debug {
		log.PrintLevel = 0
	}
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }
	// Initialize the logger
	log.Init()

	log.Infof("Initializing mauImageServer " + version)
	loadConfig()
	loadDatabase()
	loadTemplates()

	handlers.Init(config, database, auth)

	log.Infof("Registering handlers")
	http.HandleFunc("/auth/login", handlers.Login)
	http.HandleFunc("/auth/register", handlers.Register)
	http.HandleFunc("/insert", handlers.Insert)
	http.HandleFunc("/delete", handlers.Delete)
	http.HandleFunc("/hide", handlers.Hide)
	http.HandleFunc("/search", handlers.Search)
	http.HandleFunc("/", handlers.Get)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}
