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

type test struct {
	action   string
	path     string
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
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/insert",
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusCreated,
		expected: &InsertResponse{Success: true, Status: "created"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusUnauthorized,
		expected: &InsertResponse{Success: false, Status: "not-logged-in"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusUnauthorized,
		expected: &InsertResponse{Success: false, Status: "invalid-authtoken"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{authTokenError: errors.New("fakeError")},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusForbidden,
		expected: &InsertResponse{Success: false, Status: "already-exists"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser2"},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusForbidden,
		expected: &InsertResponse{Success: false, Status: "already-exists"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"ZmFrZUltYWdlDQo=\"}",
		status:   http.StatusUnsupportedMediaType,
		expected: &InsertResponse{Success: false, Status: "invalid-mime"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"image-name\":\"as>?¿d/das\"}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"totallyBase64\"}",
		status:   http.StatusUnsupportedMediaType,
		expected: &InsertResponse{Success: false, Status: "invalid-image-encoding"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusCreated,
		expected: &InsertResponse{Success: true, Status: "created"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\"}",
		status:   http.StatusInternalServerError,
		expected: &InsertResponse{Success: false, Status: "database-error"},
		config:   &data.Configuration{RequireAuth: false, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{insertError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusAccepted,
		expected: &InsertResponse{Success: true, Status: "replaced"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{imageOwner: "fakeUser"},
	}, {
		action: "POST", path: "/insert",
		request:  "{\"image\": \"" + image + "\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusInternalServerError,
		expected: &InsertResponse{Success: false, Status: "database-error"},
		config:   &data.Configuration{RequireAuth: true, ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{updateError: errors.New("fakeError"), imageOwner: "fakeUser"},
	}}

	for _, c := range cases {
		runTest(c, t)
	}
}

func TestDelete(t *testing.T) {
	log.InitWithWriter(nil)
	log.PrintLevel = 9002
	cases := []test{{
		action: "GET", path: "/delete",
		request:  "",
		status:   http.StatusMethodNotAllowed,
		expected: nil,
		config:   &data.Configuration{},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusBadRequest,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusUnauthorized,
		expected: &InsertResponse{Success: false, Status: "invalid-authtoken"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{authTokenError: errors.New("fakeError")},
		database: fakeDatabase{},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusForbidden,
		expected: &InsertResponse{Success: false, Status: "no-permissions"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryImage: data.ImageEntry{ImageName: "image", Format: "png", Adder: "fakeUser2"}},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusNotFound,
		expected: &InsertResponse{Success: false, Status: "not-found"},
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryError: errors.New("asd")},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusInternalServerError,
		expected: nil,
		config:   &data.Configuration{ImageLocation: "/tmp"},
		auth:     fakeAuth{},
		database: fakeDatabase{queryImage: data.ImageEntry{ImageName: "image", Format: "png", Adder: "fakeUser"}, removeError: errors.New("fakeError")},
	}, {
		action: "POST", path: "/delete",
		request:  "{\"image-name\":\"fakeImage\",\"username\": \"fakeUser\",\"auth-token\": \"fakeAuthToken\"}",
		status:   http.StatusAccepted,
		expected: &InsertResponse{Success: true, Status: "deleted"},
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

	for _, c := range cases {
		runTest(c, t)
	}
}

func runTest(c test, t *testing.T) {
	database = c.database
	auth = c.auth
	config = c.config

	req, err := http.NewRequest(c.action, c.path, strings.NewReader(c.request))
	if err != nil {
		t.Fatalf("Request error: %s", err)
	}
	req.RemoteAddr = "fakeIP"

	var recorder = httptest.NewRecorder()

	if c.path == "/insert" {
		insert(recorder, req)
	} else if c.path == "/delete" {
		delete(recorder, req)
	}

	if c.expected == nil {
		if recorder.Code != c.status {
			t.Errorf("Status code didn't match! Expected %d, but received %d", c.status, recorder.Code)
		}
		return
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
