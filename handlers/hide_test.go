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

func TestHide(t *testing.T) {
	log.InitWithWriter(nil)
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/hide", assert: defaultAssert,
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusUnauthorized,
		expected: &GenericResponse{Success: false, Status: "invalid-authtoken"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{authTokenError: errors.New("fakeError")},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusForbidden,
		expected: &GenericResponse{Success: false, Status: "no-permissions"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser2"},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusNotFound,
		expected: &GenericResponse{Success: false, Status: "not-found"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser", hideError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": true}",
		status:   http.StatusAccepted,
		expected: &GenericResponse{Success: true, Status: "hidden"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}, {
		action: "POST", path: "/hide", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\",\"hidden\": false}",
		status:   http.StatusAccepted,
		expected: &GenericResponse{Success: true, Status: "unhidden"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}}

	for index, c := range cases {
		run(index+1, c, t)
	}
}
