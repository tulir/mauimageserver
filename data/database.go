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
		database, err = sql.Open("mysql", fmt.Sprintf("%[1]s@%[2]s/%[3]s", conf.Authentication.ToString(), conf.Connection.ToString(), conf.Database))
	} else if sqlType == "sqlite" {
		database, err = sql.Open("sqlite3", fmt.Sprintf("%[1]s", conf.Database))
	} else {
		return fmt.Errorf("%[1]s is not yet supported", conf.Type)
	}

	if err != nil {
		return err
	} else if database == nil {
		return fmt.Errorf("Failed to open SQL connection!")
	}
	_, err = database.Query("CREATE TABLE IF NOT EXISTS users (username VARCHAR(16) PRIMARY KEY, password BINARY(60) NOT NULL, authtoken BINARY(60));")
	if err != nil {
		return err
	}
	_, err = database.Query("CREATE TABLE IF NOT EXISTS images (imgname VARCHAR(32) PRIMARY KEY, adder VARCHAR(16), adderip VARCHAR(64));")
	if err != nil {
		return err
	}
	return nil
}

// UnloadDatabase unloads the database.
func UnloadDatabase() {
	database.Close()
}

// GetOwner gets the owner of the image with the given name.
func GetOwner(imageName string) string {
	result, err := database.Query("SELECT adder FROM images WHERE imgname=?", imageName)
	if err != nil {
		return ""
	}
	for result.Next() {
		if result.Err() != nil {
			return ""
		}
		var adder string
		result.Scan(&adder)
		if len(adder) != 0 {
			return adder
		}
	}
	return ""
}

// Remove removes the image with the given name.
func Remove(imageName string) error {
	_, err := database.Query("DELETE FROM images WHERE imgname=?", imageName)
	return err
}

// Insert inserts the given image name and marks it owned by the given username.
func Insert(imageName, adder, adderip string) error {
	_, err := database.Query("INSERT INTO images VALUES(?, ?, ?);", imageName, adder, adderip)
	return err
}

// CheckAuthToken checks if the given auth token is valid for the given user.
func CheckAuthToken(username string, authtoken []byte) error {
	result, err := database.Query("SELECT authtoken FROM users WHERE username=?;", username)
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
				err = bcrypt.CompareHashAndPassword(hash, authtoken)
				if err != nil {
					return fmt.Errorf("invalid-authtoken")
				}
				return nil
			}
		}
	}
	return fmt.Errorf("invalid-authtoken")
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

	authToken, authHash := newAuthToken()
	if len(authToken) == 0 {
		return "", fmt.Errorf("authtoken-generror")
	}

	// Update database.
	database.Query("UPDATE users SET authtoken=? WHERE username=?;", authHash, username)
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

	authToken, authHash := newAuthToken()
	if len(authToken) == 0 {
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
	_, err = database.Query("INSERT INTO users VALUES(?, ?, ?)", username, hash, authHash)
	// Make sure nothing went wrong.
	if err != nil {
		// Something went wrong, return error.
		return "", fmt.Errorf("inserterror")
	}

	// Return the auth token.
	return authToken, nil
}

func newAuthToken() (string, []byte) {
	// Generate an authentication token.
	authToken := random.AuthToken()
	// Make sure it was generated.
	if authToken == "" {
		// Generation failed, return error.
		return "", nil
	}

	// Generate the bcrypt hash from the generated authentication token.
	authHash, err := bcrypt.GenerateFromPassword([]byte(authToken), bcrypt.DefaultCost-3)
	// Make sure nothing went wrong.
	if err != nil {
		// Something went wrong, return error.
		return "", nil
	}
	return authToken, authHash
}
