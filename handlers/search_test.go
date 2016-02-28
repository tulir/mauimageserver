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
	"encoding/json"
	"errors"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch(t *testing.T) {
	log.InitWithWriter(nil)
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/search", assert: defaultAssert,
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/search", assert: defaultAssert,
		request:  "{}",
		status:   http.StatusForbidden,
		expected: nil,
		config:   &data.Configuration{AllowSearch: false},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/search", assert: defaultAssert,
		request:  "{}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{AllowSearch: true},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/search", assert: defaultAssert,
		request:  "{\"uploaded-after\": 12345}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{AllowSearch: true},
		auth:     fakeAuth{},
		database: fakeDatabase{searchError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/search",
		request:  "{\"uploaded-before\": 12345}",
		status:   http.StatusOK,
		expected: nil,
		config:   &data.Configuration{AllowSearch: true},
		auth:     fakeAuth{},
		database: fakeDatabase{searchImages: []data.ImageEntry{{ImageName: "asd"}, {ImageName: "dsa"}}},
		assert: func(index int, c test, t *testing.T, recorder *httptest.ResponseRecorder) {
			var expected = []data.ImageEntry{{ImageName: "asd"}, {ImageName: "dsa"}}

			var received []data.ImageEntry
			err := json.Unmarshal(recorder.Body.Bytes(), &received)
			if err != nil {
				t.Errorf("[%s #%d] Response JSON invalid: %s", c.path, index, err)
			} else if recorder.Code != c.status {
				t.Errorf("[%s #%d] Status code didn't match! Expected %d, but received %d", c.path, index, c.status, recorder.Code)
			} else if len(received) != len(expected) {
				t.Errorf("[%s #%d] Number of results didn't match! Expected %d, but received %d", c.path, index, len(expected), len(received))
			} else {
				for i := 0; i < len(received); i++ {
					if received[i].ImageName != expected[i].ImageName {
						t.Errorf("[%s #%d] Image name of result #%d didn't match! Expected %s, but received %s", c.path, index, i, expected[i].ImageName, received[i].ImageName)
					}
				}
			}
		},
	}}

	for index, c := range cases {
		run(index+1, c, t)
	}
}
