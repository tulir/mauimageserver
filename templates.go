package main

import (
	"html/template"
	log "maunium.net/go/maulogger"
	"os"
)

var image *template.Template

type imageTemplate struct {
	ImageName string
	ImageAddr string
	Uploader  string
	Date      string
	Client    string
	Index     string
}

func loadTemplates() {
	log.Infof("Loading HTML templates...")
	var err error
	/*index, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatalf("Failed to load index page: %s", err)
		os.Exit(3)
	}*/
	image, err = template.ParseFiles("image.html")
	if err != nil {
		log.Fatalf("Failed to load image page: %s", err)
		os.Exit(3)
	}
	log.Debugln("Successfully loaded HTML templates")
}
