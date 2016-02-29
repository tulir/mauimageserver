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
	"database/sql"
	"encoding/json"
	"maunium.net/go/mauimageserver/data"
	"maunium.net/go/mauth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type test struct {
	action   string
	path     string
	request  string
	status   int
	expected *GenericResponse
	database data.MISDatabase
	auth     mauth.System
	config   *data.Configuration
	assert   func(index int, c test, t *testing.T, recorder *httptest.ResponseRecorder)
}

func run(index int, c test, t *testing.T) {
	Init(c.config, c.database, c.auth)

	req, err := http.NewRequest(c.action, c.path, strings.NewReader(c.request))
	if err != nil {
		t.Fatalf("Request error: %s", err)
	}
	req.RemoteAddr = "fakeIP"

	var recorder = httptest.NewRecorder()

	if c.path == "/insert" {
		Insert(recorder, req)
	} else if c.path == "/delete" {
		Delete(recorder, req)
	} else if c.path == "/hide" {
		Hide(recorder, req)
	} else if c.path == "/search" {
		Search(recorder, req)
	}

	c.assert(index, c, t, recorder)
}

func defaultAssert(index int, c test, t *testing.T, recorder *httptest.ResponseRecorder) {
	if c.expected == nil {
		if recorder.Code != c.status {
			t.Errorf("[%s #%d] Status code didn't match! Expected %d, but received %d", c.path, index, c.status, recorder.Code)
		}
		return
	}

	var received GenericResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &received)

	if err != nil {
		t.Errorf("[%s #%d] Response JSON invalid: %s", c.path, index, err)
	} else if recorder.Code != c.status {
		t.Errorf("[%s #%d] Status code didn't match! Expected %d, but received %d", c.path, index, c.status, recorder.Code)
	} else if received.Success != c.expected.Success {
		t.Errorf("[%s #%d] Success value didn't match! Expected %b, but received %b", c.path, index, c.expected.Success, received.Success)
	} else if received.Status != c.expected.Status {
		t.Errorf("[%s #%d] Status message didn't match! Expected %s, but received %s", c.path, index, c.expected.Status, received.Status)
	}
}

type fakeAuth struct {
	registerStr   string
	registerError error

	loginStr   string
	loginError error

	authTokenError error
}

func (fake fakeAuth) RegisterHTTPD(w http.ResponseWriter, r *http.Request) {}
func (fake fakeAuth) LoginHTTPD(w http.ResponseWriter, r *http.Request)    {}

func (fake fakeAuth) Register(username string, password []byte) (string, error) {
	return fake.registerStr, fake.registerError
}
func (fake fakeAuth) RegisterHTTP(w http.ResponseWriter, r *http.Request) (string, error) {
	return fake.registerStr, fake.registerError
}
func (fake fakeAuth) Login(username string, password []byte) (string, error) {
	return fake.loginStr, fake.loginError
}
func (fake fakeAuth) LoginHTTP(w http.ResponseWriter, r *http.Request) (string, error) {
	return fake.loginStr, fake.loginError
}
func (fake fakeAuth) CheckAuthToken(username string, authtoken []byte) error {
	return fake.authTokenError
}

type fakeDatabase struct {
	queryImage data.ImageEntry
	queryError error
	imageOwner string

	searchImages []data.ImageEntry
	searchError  error

	removeError error
	hideError   error
	insertError error
	updateError error
}

func (fake fakeDatabase) Load() error            { return nil }
func (fake fakeDatabase) Unload() error          { return nil }
func (fake fakeDatabase) GetInternalDB() *sql.DB { return nil }

func (fake fakeDatabase) Insert(imageName, imageFormat, mimeType, adder, adderip, client string, hidden bool) error {
	return fake.insertError
}
func (fake fakeDatabase) Update(imageName, imageFormat, mimeType, adderip, client string, hidden bool) error {
	return fake.updateError
}
func (fake fakeDatabase) Remove(imageName string) error {
	return fake.removeError
}
func (fake fakeDatabase) SetHidden(imageName string, hidden bool) error {
	return fake.hideError
}
func (fake fakeDatabase) Query(imageName string) (data.ImageEntry, error) {
	return fake.queryImage, fake.queryError
}
func (fake fakeDatabase) GetOwner(imageName string) string {
	return fake.imageOwner
}
func (fake fakeDatabase) Search(format, adder, client string, timeMin, timeMax int64) ([]data.ImageEntry, error) {
	return fake.searchImages, fake.searchError
}
