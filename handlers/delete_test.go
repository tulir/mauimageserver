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

func TestDelete(t *testing.T) {
	log.InitWithWriter(nil)
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/delete", assert: defaultAssert,
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusUnauthorized,
		expected: &GenericResponse{Success: false, Status: "invalid-authtoken"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{authTokenError: errors.New("fakeError")},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusForbidden,
		expected: &GenericResponse{Success: false, Status: "no-permissions"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryImage: data.ImageEntry{ImageName: "image", Format: "png", Adder: "fakeUser2"}},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusNotFound,
		expected: &GenericResponse{Success: false, Status: "not-found"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryError: errors.New("asd")},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryImage: data.ImageEntry{ImageName: "image", Format: "png", Adder: "fakeUser"}, removeError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/delete", assert: defaultAssert,
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusAccepted,
		expected: &GenericResponse{Success: true, Status: "deleted"},
		config:   &data.Configuration{ImageLocation: "/"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryImage: data.ImageEntry{ImageName: "image", Format: "png", Adder: "fakeUser"}},
	}}

	/*{ TODO: Create a test case that makes os.Remove throw an error other than no such file or directory.
		request:  "{\"image-name\":\"fake/Image\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	},*/

	for index, c := range cases {
		run(index+1, c, t)
	}
}
