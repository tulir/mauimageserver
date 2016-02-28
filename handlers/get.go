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

// Package handlers contains the MIS-specific HTTP request handlers
package handlers

import (
	"io/ioutil"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Get handles get requests
func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path[1:]

	img, err := database.Query(path)
	if err == nil {
		date := time.Unix(img.Timestamp, 0).Format(config.DateFormat)
		r.URL.Path = r.URL.Path + "." + img.Format
		data.ImagePage{
			ImageName: img.ImageName,
			ImageAddr: r.URL.String(),
			Uploader:  img.Adder,
			Client:    img.Client,
			Date:      date,
			Index:     strconv.Itoa(img.ID),
		}.Send(w)
		return
	}

	imgData, err := ioutil.ReadFile(config.ImageLocation + r.URL.Path)
	if err != nil {
		log.Errorf("Failed to read image at %[2]s requested by %[1]s: %[3]s", getIP(r), path, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)

	split := strings.Split(path, ".")
	if len(split) > 0 {
		img, err = database.Query(split[0])
		if err == nil && len(img.Format) > 0 {
			w.Header().Set("Content-type", "image/"+img.Format)
		} else if len(split) > 1 {
			w.Header().Set("Content-type", "image/"+split[len(split)-1])
		}
	}

	w.Write(imgData)
}
