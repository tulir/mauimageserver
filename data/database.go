package data

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"maunium.net/go/mauimageserver/random"
	"strings"
)

var database *sql.DB

// LoadDatabase loads the database based on the given configuration.
func LoadDatabase(conf SQLConfig) error {
	var err error
	sqlType := strings.ToLower(conf.Type)
	if sqlType == "mysql" {
		database, err = sql.Open(sqlType, fmt.Sprintf("%[1]s@%[2]s/%[3]s", conf.Authentication.ToString(), conf.Connection.ToString(), conf.Database))
	} else {
		return fmt.Errorf("%[1]s is not yet supported", conf.Type)
	}

	if err != nil {
		return err
	} else if database == nil {
		return fmt.Errorf("Failed to open SQL connection!")
	}
	_, err = database.Query("CREATE TABLE IF NOT EXISTS users (username VARCHAR(16) PRIMARY KEY, password BINARY(60) NOT NULL, authtoken VARCHAR(64));")
	if err != nil {
		return err
	}
	_, err = database.Query("CREATE TABLE IF NOT EXISTS images (name VARCHAR(32) PRIMARY KEY, path VARCHAR(255) NOT NULL, adder VARCHAR(16), adderip VARCHAR(64));")
	if err != nil {
		return err
	}
	return nil
}

// InsertResult tells the result of inserting an image.
type InsertResult int

// Errored ...
const Errored InsertResult = -2

// AlreadyExists ...
const AlreadyExists InsertResult = -1

// Inserted ...
const Inserted InsertResult = 1

// Replaced ...
const Replaced InsertResult = 2

// Insert inserts the given name->path mapping and marks it owned by the given username.
func Insert(path, name, adder, adderip string) (InsertResult, string) {
	var oldPath = ""
	result, err := database.Query("SELECT adder, path FROM images WHERE name=?", name)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var adder2, path2 string
			result.Scan(&adder2, &path2)
			if adder2 != adder || adder == "anonymous" {
				return AlreadyExists, adder2
			}
			oldPath = path2
		}
	}
	if len(oldPath) != 0 {
		_, err = database.Query("UPDATE images SET path=?,adderip=? WHERE name=?;", path, adderip, name)
		if err != nil {
			return Errored, err.Error()
		}
		return Replaced, oldPath
	}
	_, err = database.Query("INSERT INTO images VALUES(?, ?, ?, ?);", name, path, adder, adderip)
	if err != nil {
		return Errored, err.Error()
	}
	return Inserted, ""
}

// Login generates an authentication token for the user.
func Login(username string, password []byte) string {
	var correctPassword = false
	result, err := database.Query("SELECT password FROM users WHERE name=?", username)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var hash []byte
			result.Scan(&hash)
			if len(hash) != 0 {
				err = bcrypt.CompareHashAndPassword(hash, password)
				correctPassword = err == nil
			}
		}
	}
	if !correctPassword {
		return "pwd"
	}
	authToken := random.AuthToken()
	if authToken == "" {
		return "authgen"
	}
	database.Query("UPDATE users SET authtoken=? WHERE name=?;", authToken, username)
	return authToken
}

// Register creates an account and generates an authentication token for it.
func Register(username string, password []byte) string {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return "hash"
	}

	authToken := random.AuthToken()
	if authToken == "" {
		return "authgen"
	}

	result, err := database.Query("SELECT EXISTS(SELECT 1 FROM users WHERE username=?)", username)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var res int
			result.Scan(&res)
			if res == 1 {
				return "userexists"
			}
		}
	}
	_, err = database.Query("INSERT INTO users VALUES(?, ?, ?)", username, hash, authToken)
	if err != nil {
		return "inserterror"
	}
	return authToken
}

/*// Insert inserts the given URL, short url and redirect type into the database.
// If the URL has already been shortened with the same redirect type, the already existing short URL will be returned.
// In any other case, the requested short URL will be returned.
// Warning: This will NOT check if the short URL is in use.
func Insert(url, ishort, redirect string) string {
	redirect = strings.ToLower(redirect)
	if redirect != "http" && redirect != "html" && redirect != "js" {
		redirect = "http"
	}
	result, err := database.Query("SELECT short FROM links WHERE url=? AND redirect=?;", url, redirect)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var short string
			result.Scan(&short)
			if len(short) != 0 {
				return short
			}
		}
	}
	InsertDirect(ishort, url, redirect)
	return ishort
}

// InsertDirect inserts the given values into the database, no questions asked (except by the database itself)
func InsertDirect(short, url, redirect string) error {
	_, err := database.Query("INSERT INTO links VALUES(?, ?, ?);", url, short, redirect)
	if err != nil {
		return err
	}
	return nil
}

// Query queries for the given short URL and returns the long URL and redirect type.
func Query(short string) (string, string, error) {
	result, err := database.Query("SELECT url, redirect FROM links WHERE short=?;", short)
	if err != nil {
		return "", "", err
	}
	defer result.Close()
	for result.Next() {
		if result.Err() != nil {
			return "", "", result.Err()
		}
		var long, redirect string
		result.Scan(&long, &redirect)
		if len(long) == 0 {
			continue
		} else if len(redirect) == 0 {
			redirect = "http"
		}
		return long, redirect, nil
	}
	result.Close()
	return "", "", fmt.Errorf("ID not found")
}*/
