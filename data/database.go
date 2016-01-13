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
func Login(username string, password []byte) (string, error) {
	var correctPassword = false
	// Get the password of the given user.
	result, err := database.Query("SELECT password FROM users WHERE username=?;", username)
	// Check if there was an error.
	if err == nil {
		// Loop through the result rows.
		for result.Next() {
			// Check if the current result has an error.
			if result.Err() != nil {
				break
			}
			// Define the byte array for the password hash in the database.
			var hash []byte
			// Scan the hash from the database result into the previously defined byte array.
			result.Scan(&hash)
			// Make sure the scan was successful.
			if len(hash) != 0 {
				// Compare the hash and the given password.
				err = bcrypt.CompareHashAndPassword(hash, password)
				// Set the correctPassword field to the correct value.
				correctPassword = err == nil
			}
		}
	}
	// Check if the password was correct.
	if !correctPassword {
		// Return error if the password was wrong.
		return "", fmt.Errorf("incorrectpassword")
	}
	// Generate an authentication token.
	authToken := random.AuthToken()
	// Make sure it was generated.
	if authToken == "" {
		// Generation failed, return error
		return "", fmt.Errorf("authtoken-generror")
	}
	// Update database.
	database.Query("UPDATE users SET authtoken=? WHERE name=?;", authToken, username)
	// Return auth token.
	return authToken, nil
}

// Register creates an account and generates an authentication token for it.
func Register(username string, password []byte) (string, error) {
	// Generate the bcrypt hash from the given password.
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	// Make sure nothing went wrong.
	if err != nil {
		// Something went wrong, return error.
		return "", fmt.Errorf("hashgen")
	}

	// Generate an authentication token.
	authToken := random.AuthToken()
	// Make sure it was generated.
	if authToken == "" {
		// Generation failed, return error.
		return "", fmt.Errorf("authtoken-generror")
	}

	// Check if the username already exists in the database.
	result, err := database.Query("SELECT EXISTS(SELECT 1 FROM users WHERE username=?)", username)
	if err == nil {
		for result.Next() {
			if result.Err() != nil {
				break
			}
			var res int
			result.Scan(&res)
			if res == 1 {
				// User exists, return error.
				return "", fmt.Errorf("userexists")
			}
		}
	}

	// Insert user into database.
	_, err = database.Query("INSERT INTO users VALUES(?, ?, ?)", username, hash, authToken)
	// Make sure nothing went wrong.
	if err != nil {
		// Something went wrong, return error.
		return "", fmt.Errorf("inserterror")
	}

	// Return the auth token.
	return authToken, nil
}
