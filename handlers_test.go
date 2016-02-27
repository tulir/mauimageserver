package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"maunium.net/go/mauth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type insertTest struct {
	request  string
	status   int
	expected *InsertResponse
	database data.MISDatabase
	auth     mauth.System
	config   *data.Configuration
}

var image = "iVBORw0KGgoAAAANSUhEUgAAABUAAAARCAIAAAC95HDXAAAAFklEQVR42mP4ThlgGNU/qn9U/4jVDwBiDAmW9sWkNgAAAABJRU5ErkJggg=="

func TestInsert(t *testing.T) {
	log.InitWithWriter(nil)
	// The cases before json parsing has succeeded
	{
		config = &data.Configuration{}
		req, err := http.NewRequest("GET", "/insert", nil)
		if err != nil {
			t.Fatalf("Request error: %s", err)
		}
		req.RemoteAddr = "fakeIP"
		var recorder = httptest.NewRecorder()
		insert(recorder, req)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("Status code didn't match! Expected %d, but received %d", http.StatusMethodNotAllowed, recorder.Code)
		}
	}

	cases := []insertTest{
		{
			request:  "{}",
			status:   http.StatusBadRequest,
			expected: nil,
			config:   &data.Configuration{},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\"}",
			status:   http.StatusCreated,
			expected: &InsertResponse{Success: true, Status: "created"},
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\"}",
			status:   http.StatusUnauthorized,
			expected: &InsertResponse{Success: false, Status: "not-logged-in"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
			status:   http.StatusUnauthorized,
			expected: &InsertResponse{Success: false, Status: "invalid-authtoken"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: errors.New("fakeError")},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
			status:   http.StatusForbidden,
			expected: &InsertResponse{Success: false, Status: "already-exists"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{imageOwner: "fakeUser2"},
		},
		{
			request:  "{\"image\": \"" + image + "\"}",
			status:   http.StatusForbidden,
			expected: &InsertResponse{Success: false, Status: "already-exists"},
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{imageOwner: "fakeUser"},
		},
		{
			request:  "{\"image\": \"ZmFrZUltYWdlDQo=\"}",
			status:   http.StatusUnsupportedMediaType,
			expected: &InsertResponse{Success: false, Status: "invalid-mime"},
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"image-name\":\"as>?¿d/das\"}",
			status:   http.StatusInternalServerError,
			expected: nil,
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"totallyBase64\"}",
			status:   http.StatusUnsupportedMediaType,
			expected: &InsertResponse{Success: false, Status: "invalid-image-encoding"},
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
			status:   http.StatusCreated,
			expected: &InsertResponse{Success: true, Status: "created"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{},
		},
		{
			request:  "{\"image\": \"" + image + "\"}",
			status:   http.StatusInternalServerError,
			expected: &InsertResponse{Success: false, Status: "database-error"},
			config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{insertError: errors.New("fakeError")},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
			status:   http.StatusAccepted,
			expected: &InsertResponse{Success: true, Status: "replaced"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{imageOwner: "fakeUser"},
		},
		{
			request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
			status:   http.StatusInternalServerError,
			expected: &InsertResponse{Success: false, Status: "database-error"},
			config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
			auth:     fakeAuth{registerStr: "", loginStr: "", registerError: nil, loginError: nil, authTokenError: nil},
			database: fakeDatabase{updateError: errors.New("fakeError"), imageOwner: "fakeUser"},
		},
	}

	for _, c := range cases {
		database = c.database
		auth = c.auth
		config = c.config

		req, err := http.NewRequest("POST", "/insert", strings.NewReader(c.request))
		if err != nil {
			t.Fatalf("Request error: %s", err)
		}
		req.RemoteAddr = "fakeIP"

		var recorder = httptest.NewRecorder()

		insert(recorder, req)

		if c.expected == nil {
			if recorder.Code != c.status {
				t.Errorf("Status code didn't match! Expected %d, but received %d", c.status, recorder.Code)
			}
			continue
		}

		var received InsertResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &received)

		if err != nil {
			t.Errorf("Response JSON invalid: %s", err)
		} else if recorder.Code != c.status {
			t.Errorf("Status code didn't match! Expected %d, but received %d", c.status, recorder.Code)
		} else if received.Success != c.expected.Success {
			t.Errorf("Success value didn't match! Expected %b, but received %b", c.expected.Success, received.Success)
		} else if received.Status != c.expected.Status {
			t.Errorf("Status message didn't match! Expected %s, but received %s", c.expected.Status, received.Status)
		}
	}
}

type deleteTest struct {
	request  string
	status   int
	expected *InsertResponse
	database data.MISDatabase
	auth     mauth.System
	config   *data.Configuration
}

func TestDelete(t *testing.T) {
	log.InitWithWriter(nil)
	{
		config = &data.Configuration{}
		req, err := http.NewRequest("GET", "/delete", nil)
		if err != nil {
			t.Fatalf("Request error: %s", err)
		}
		req.RemoteAddr = "fakeIP"
		var recorder = httptest.NewRecorder()
		delete(recorder, req)
		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("Status code didn't match! Expected %d, but received %d", http.StatusMethodNotAllowed, recorder.Code)
		}
	}

	cases := []deleteTest{
	// TODO: Delete tests
	}

	for _, c := range cases {
		database = c.database
		auth = c.auth
		config = c.config

		req, err := http.NewRequest("POST", "/delete", strings.NewReader(c.request))
		if err != nil {
			t.Fatalf("Request error: %s", err)
		}
		req.RemoteAddr = "fakeIP"

		var recorder = httptest.NewRecorder()

		delete(recorder, req)

		if c.expected == nil {
			if recorder.Code != c.status {
				t.Errorf("Status code didn't match! Expected %d, but received %d", c.status, recorder.Code)
			}
			continue
		}

		var received InsertResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &received)

		if err != nil {
			t.Errorf("Response JSON invalid: %s", err)
		} else if recorder.Code != c.status {
			t.Errorf("Status code didn't match! Expected %d, but received %d", c.status, recorder.Code)
		} else if received.Success != c.expected.Success {
			t.Errorf("Success value didn't match! Expected %b, but received %b", c.expected.Success, received.Success)
		} else if received.Status != c.expected.Status {
			t.Errorf("Status message didn't match! Expected %s, but received %s", c.expected.Status, received.Status)
		}
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
