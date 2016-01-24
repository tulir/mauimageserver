package data

import (
	"html/template"
	"net/http"
)

var image *template.Template

// ImagePage contains the data needed for an image viewing page.
type ImagePage struct {
	ImageName string
	ImageAddr string
	Uploader  string
	Date      string
	Client    string
	Index     string
}

// Send sends this ImagePage to the given response writer.
func (ip ImagePage) Send(w http.ResponseWriter) {
	image.Execute(w, ip)
}

// LoadTemplates loads all required templates.
func LoadTemplates() error {
	var err error
	/*index, err = template.ParseFiles("index.html")
		if err != nil {
	        return err
		}*/
	image, err = template.ParseFiles("image.html")
	if err != nil {
		return err
	}
	return nil
}
