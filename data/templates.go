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

// Package data contains all data storage things (config, database, etc...)
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
func LoadTemplates(path string) error {
	var err error
	image, err = template.ParseFiles(path)
	if err != nil {
		return err
	}
	return nil
}
