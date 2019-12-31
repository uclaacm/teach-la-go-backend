package lib

import (
	"time"

	"cloud.google.com/go/firestore"
)

// Program is a representation of a program document.
type Program struct {
	Code        string    `json:"code" firestore:"code"`
	DateCreated time.Time `json:"dateCreated" firestore:"dateCreated"`
	Language    string    `json:"language" firestore:"language"`
	Name        string    `json:"name" firestore:"name"`
	Thumbnail   int64     `json:"thumbnail" firestore:"thumbnail"`
}

// ToFirestoreUpdate returns the []firestore.Update representation
// of this struct. Any fields that are non-zero valued are included
// in the update, save for the date of creation.
func (p *Program) ToFirestoreUpdate() (up []firestore.Update) {
	if p.Code != "" {
		up = append(up, firestore.Update{Path: "code", Value: p.Code})
	}
	if p.Language != "" {
		up = append(up, firestore.Update{Path: "language", Value: p.Language})
	}
	if p.Name != "" {
		up = append(up, firestore.Update{Path: "name", Value: p.Name})
	}
	if p.Thumbnail != 0 {
		up = append(up, firestore.Update{Path: "thumbnail", Value: p.Thumbnail})
	}

	return
}

// UserData is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type UserData struct {
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	Programs          []string `firestore:"programs" json:"programs"`
	Classes           []string `firestore:"classes" json:"classes"`
}

// ToFirestoreUpdate returns the database update
// representation of its UserData struct.
func (u *UserData) ToFirestoreUpdate() []firestore.Update {
	f := []firestore.Update{
		{Path: "mostRecentProgram", Value: u.MostRecentProgram},
		{Path: "programs", Value: u.Programs},
	}
	return f
}
