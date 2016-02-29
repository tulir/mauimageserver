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
	"errors"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"testing"
)

var image = "iVBORw0KGgoAAAANSUhEUgAAABUAAAARCAIAAAC95HDXAAAAFklEQVR42mP4ThlgGNU/qn9U/4jVDwBiDAmW9sWkNgAAAABJRU5ErkJggg=="

func TestInsert(t *testing.T) {
	log.InitWithWriter(nil)
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/insert", assert: defaultAssert,
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusCreated,
		expected: &GenericResponse{Success: true, Status: "created"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusUnauthorized,
		expected: &GenericResponse{Success: false, Status: "not-logged-in"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusUnauthorized,
		expected: &GenericResponse{Success: false, Status: "invalid-authtoken"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{authTokenError: errors.New("fakeError")},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusForbidden,
		expected: &GenericResponse{Success: false, Status: "already-exists"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser2"},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusForbidden,
		expected: &GenericResponse{Success: false, Status: "already-exists"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"ZmFrZUltYWdlDQo=\"}",
		status:   http.StatusUnsupportedMediaType,
		expected: &GenericResponse{Success: false, Status: "invalid-mime"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"image-name\":\"as>?¿d/das\"}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"totallyBase64\"}",
		status:   http.StatusUnsupportedMediaType,
		expected: &GenericResponse{Success: false, Status: "invalid-image-encoding"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusCreated,
		expected: &GenericResponse{Success: true, Status: "created"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusInternalServerError,
		expected: &GenericResponse{Success: false, Status: "database-error"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{insertError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusAccepted,
		expected: &GenericResponse{Success: true, Status: "replaced"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}, {
		action: "POST", path: "/insert", assert: defaultAssert,
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusInternalServerError,
		expected: &GenericResponse{Success: false, Status: "database-error"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{updateError: errors.New("fakeError"), imageOwner: "fakeUser"},
	}}

	for index, c := range cases {
		run(index+1, c, t)
	}
}
