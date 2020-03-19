package db

import (
	"cloud.google.com/go/firestore"
)

// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type User struct {
	Classes           []string `firestore:"classes" json:"classes"`
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	MostRecentProgram int      `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	Programs          []string `firestore:"programs" json:"programs"`
	UID               string   `json:"uid"`
}

// ToFirestoreUpdate returns the database update
// representation of its UserData struct.
func (u *User) ToFirestoreUpdate() (f []firestore.Update) {
	if u.DisplayName != "" {
		f = append(f, firestore.Update{Path: "displayName", Value: u.DisplayName})
	}

	f = append(f, firestore.Update{Path: "mostRecentProgram", Value: u.MostRecentProgram})

	if u.PhotoName != "" {
		f = append(f, firestore.Update{Path: "photoName", Value: u.PhotoName})
	}

	if len(u.Programs) != 0 {
		f = append(f, firestore.Update{Path: "programs", Value: firestore.ArrayUnion(u.Programs)})
	}

	return f
}
