package data

import (
	"database/sql"
	"fmt"
	"time"
)

// ImageEntry is an image entry.
type ImageEntry struct {
	ImageName string `json:"image-name"`
	Format    string `json:"image-format,omitempty"`
	MimeType  string `json:"mime-type,omitempty"`
	Adder     string `json:"adder,omitempty"`
	AdderIP   string `json:"adder-ip,omitempty"`
	Client    string `json:"client-name,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	ID        int    `json:"id,omitempty"`
	Hidden    bool   `json:"hidden,omitempty"`
}

var database *sql.DB

// LoadDatabase loads the database based on the given configuration.
func LoadDatabase(conf SQLConfig) (*sql.DB, error) {
	var err error
	database, err = sql.Open("mysql", fmt.Sprintf("%[1]s@%[2]s/%[3]s", conf.Authentication.ToString(), conf.Connection.ToString(), conf.Database))

	if err != nil {
		return nil, err
	} else if database == nil {
		return nil, fmt.Errorf("Failed to open SQL connection!")
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS images (" +
		"imgname VARCHAR(32) PRIMARY KEY," +
		"format VARCHAR(16)," +
		"mimetype VARCHAR(16)," +
		"adder VARCHAR(16) NOT NULL," +
		"adderip VARCHAR(64) NOT NULL," +
		"client VARCHAR(64) NOT NULL," +
		"timestamp BIGINT NOT NULL," +
		"hidden TINYINT(1) NOT NULL," +
		"id MEDIUMINT UNIQUE KEY AUTO_INCREMENT" +
		");")
	if err != nil {
		return nil, err
	}
	return database, nil
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
	defer result.Close()
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

// Search searches the database with the given arguments.
func Search(format, adder, client string, timeMin, timeMax int64) ([]ImageEntry, error) {
	var result *sql.Rows
	var err error
	if len(format) == 0 {
		if len(adder) == 0 {
			if len(client) == 0 {
				if timeMin <= 0 || timeMax <= 0 {
					result, err = database.Query("SELECT * FROM images;")
				} else {
					result, err = database.Query("SELECT * FROM images WHERE timestamp BETWEEN ? AND ?;", timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin <= 0 || timeMax <= 0 {
					result, err = database.Query("SELECT * FROM images WHERE (client LIKE ?);", client)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", client, timeMin, timeMax)
				}
			}
		} else {
			adder = "%" + adder + "%"
			if len(client) == 0 {
				if timeMin <= 0 || timeMax <= 0 {
					result, err = database.Query("SELECT * FROM images WHERE (adder LIKE ?);", adder)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (timestamp BETWEEN ? AND ?);", adder, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin <= 0 || timeMax <= 0 {
					result, err = database.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (client LIKE ?);", adder, client)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", adder, client, timeMin, timeMax)
				}
			}
		}
	} else {
		if len(adder) == 0 {
			if len(client) == 0 {
				if timeMin == 0 || timeMax == 0 {
					result, err = database.Query("SELECT * FROM images WHERE format=?;", format)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (timestamp BETWEEN ? AND ?);", format, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin == 0 || timeMax == 0 {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (client LIKE ?);", format, client)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, client, timeMin, timeMax)
				}
			}
		} else {
			adder = "%" + adder + "%"
			if len(client) == 0 {
				if timeMin == 0 || timeMax == 0 {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?);", format, adder)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, adder, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin == 0 || timeMax == 0 {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (client LIKE ?);", format, adder, client)
				} else {
					result, err = database.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, adder, client, timeMin, timeMax)
				}
			}
		}
	}
	defer result.Close()
	var results []ImageEntry
	if err != nil {
		return results, err
	}
	for result.Next() {
		if result.Err() != nil {
			continue
		}
		var imageName, format, mimeType, adder, adderip, client string
		var timestamp int64
		var id, hidden int

		err = result.Scan(&imageName, &format, &mimeType, &adder, &adderip, &client, &timestamp, &hidden, &id)
		if err != nil || hidden != 0 {
			continue
		}

		results = append(results, ImageEntry{ImageName: imageName, Format: format, MimeType: mimeType, Adder: adder, Client: client, Timestamp: timestamp, ID: id, Hidden: false})
	}
	return results, nil
}

// Remove removes the image with the given name.
func Remove(imageName string) error {
	_, err := database.Exec("DELETE FROM images WHERE imgname=?", imageName)
	return err
}

// SetHidden changes the hidden status of the image.
func SetHidden(imageName string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := database.Exec("UPDATE images SET hidden=? WHERE imgname=?", hid, imageName)
	return err
}

// Insert inserts the given image name and marks it owned by the given username.
func Insert(imageName, imageFormat, mimeType, adder, adderip, client string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := database.Exec("INSERT INTO images (imgname, format, mimetype, adder, adderip, client, timestamp, hidden) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", imageName, imageFormat, mimeType, adder, adderip, client, time.Now().Unix(), hid)
	return err
}

// Update updates the image with the given name giving it the given information.
func Update(imageName, imageFormat, mimeType, adderip, client string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := database.Exec("UPDATE images SET format=?,mimetype=?,adderip=?,client=?,timestamp=?,hidden=? WHERE imgname=?", imageFormat, mimeType, adderip, client, time.Now().Unix(), hid, imageName)
	return err
}

// Query for basic details of the given image.
func Query(imageName string) (ImageEntry, error) {
	result, err := database.Query("SELECT format, mimetype, adder, adderip, client, timestamp, id, hidden FROM images WHERE imgname=?", imageName)
	if err != nil {
		return ImageEntry{}, err
	}
	defer result.Close()
	for result.Next() {
		if result.Err() != nil {
			return ImageEntry{}, result.Err()
		}
		var format, mimeType, adder, adderip, client string
		var timestamp int64
		var id, hid int
		err = result.Scan(&format, &mimeType, &adder, &adderip, &client, &timestamp, &id, &hid)

		var hidden bool
		if hid == 0 {
			hidden = false
		} else {
			hidden = true
		}

		if err != nil {
			return ImageEntry{}, err
		} else if len(adder) == 0 || len(adderip) == 0 || len(client) == 0 || timestamp < 1 || id < 1 {
			return ImageEntry{ImageName: imageName, Format: format, MimeType: mimeType, Adder: adder, AdderIP: adderip, Client: client, Timestamp: timestamp, ID: id, Hidden: hidden}, fmt.Errorf("Invalid data")
		}
		return ImageEntry{ImageName: imageName, Format: format, MimeType: mimeType, Adder: adder, AdderIP: adderip, Client: client, Timestamp: timestamp, ID: id, Hidden: hidden}, nil
	}
	return ImageEntry{}, fmt.Errorf("No data found")
}
