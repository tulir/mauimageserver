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

// Package data contains all data storage things (config, database, etc...)
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

// MISDatabase is the interface for MIS databases.
type MISDatabase interface {
	// Open a connection and initialize the underlying database.
	Load() error
	// Unload the underlying database.
	Unload() error
	// GetInternalDB returns the underlying sql.DB pointer.
	GetInternalDB() *sql.DB

	// Insert the given image name and marks it owned by the given username.
	Insert(imageName, imageFormat, mimeType, adder, adderip, client string, hidden bool) error
	// Update the image with the given name giving it the given information.
	Update(imageName, imageFormat, mimeType, adderip, client string, hidden bool) error

	// Remove the image with the given name.
	Remove(imageName string) error
	// SetHidden changes the hidden status of the image.
	SetHidden(imageName string, hidden bool) error

	// Query for basic details of the given image.
	Query(imageName string) (ImageEntry, error)
	// GetOwner gets the owner of the image with the given name.
	GetOwner(imageName string) string
	// Search the database with the given arguments.
	Search(format, adder, client string, timeMin, timeMax int64) ([]ImageEntry, error)
}

type mis struct {
	conf SQLConfig
	db   *sql.DB
}

// CreateDatabase creates an instance of MISDatabase
func CreateDatabase(config SQLConfig) MISDatabase {
	return mis{conf: config}
}

func (data mis) GetInternalDB() *sql.DB {
	return data.db
}

func (data mis) Load() error {
	var err error
	data.db, err = sql.Open("mysql", fmt.Sprintf("%[1]s@%[2]s/%[3]s", data.conf.Authentication.ToString(), data.conf.Connection.ToString(), data.conf.Database))

	if err != nil {
		return err
	} else if data.db == nil {
		return fmt.Errorf("Failed to open SQL connection!")
	}

	_, err = data.db.Exec("CREATE TABLE IF NOT EXISTS images (" +
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
		return err
	}
	return nil
}

func (data mis) Unload() error {
	return data.db.Close()
}

func (data mis) GetOwner(imageName string) string {
	result, err := data.db.Query("SELECT adder FROM images WHERE imgname=?", imageName)
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

func (data mis) Search(format, adder, client string, timeMin, timeMax int64) ([]ImageEntry, error) {
	var result *sql.Rows
	var err error
	if len(format) == 0 {
		if len(adder) == 0 {
			if len(client) == 0 {
				if timeMin <= 0 || timeMax <= 0 {
					result, err = data.db.Query("SELECT * FROM images;")
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE timestamp BETWEEN ? AND ?;", timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin <= 0 || timeMax <= 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE (client LIKE ?);", client)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", client, timeMin, timeMax)
				}
			}
		} else {
			adder = "%" + adder + "%"
			if len(client) == 0 {
				if timeMin <= 0 || timeMax <= 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE (adder LIKE ?);", adder)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (timestamp BETWEEN ? AND ?);", adder, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin <= 0 || timeMax <= 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (client LIKE ?);", adder, client)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE (adder LIKE ?) AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", adder, client, timeMin, timeMax)
				}
			}
		}
	} else {
		if len(adder) == 0 {
			if len(client) == 0 {
				if timeMin == 0 || timeMax == 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE format=?;", format)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (timestamp BETWEEN ? AND ?);", format, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin == 0 || timeMax == 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (client LIKE ?);", format, client)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, client, timeMin, timeMax)
				}
			}
		} else {
			adder = "%" + adder + "%"
			if len(client) == 0 {
				if timeMin == 0 || timeMax == 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?);", format, adder)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, adder, timeMin, timeMax)
				}
			} else {
				client = "%" + client + "%"
				if timeMin == 0 || timeMax == 0 {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (client LIKE ?);", format, adder, client)
				} else {
					result, err = data.db.Query("SELECT * FROM images WHERE format=? AND (adder LIKE ?) AND (client LIKE ?) AND (timestamp BETWEEN ? AND ?);", format, adder, client, timeMin, timeMax)
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

func (data mis) Remove(imageName string) error {
	_, err := data.db.Exec("DELETE FROM images WHERE imgname=?", imageName)
	return err
}

func (data mis) SetHidden(imageName string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := data.db.Exec("UPDATE images SET hidden=? WHERE imgname=?", hid, imageName)
	return err
}

func (data mis) Insert(imageName, imageFormat, mimeType, adder, adderip, client string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := data.db.Exec("INSERT INTO images (imgname, format, mimetype, adder, adderip, client, timestamp, hidden) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", imageName, imageFormat, mimeType, adder, adderip, client, time.Now().Unix(), hid)
	return err
}

func (data mis) Update(imageName, imageFormat, mimeType, adderip, client string, hidden bool) error {
	var hid int
	if hidden {
		hid = 1
	} else {
		hid = 0
	}
	_, err := data.db.Exec("UPDATE images SET format=?,mimetype=?,adderip=?,client=?,timestamp=?,hidden=? WHERE imgname=?", imageFormat, mimeType, adderip, client, time.Now().Unix(), hid, imageName)
	return err
}

func (data mis) Query(imageName string) (ImageEntry, error) {
	result, err := data.db.Query("SELECT format, mimetype, adder, adderip, client, timestamp, id, hidden FROM images WHERE imgname=?", imageName)
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
