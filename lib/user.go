package lib

import "cloud.google.com/go/firestore"

// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type User struct {
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	Programs          []string `firestore:"programs" json:"programs"`
	Classes           []string `firestore:"classes" json:"classes"`
}

// ToFirestoreUpdate returns the database update
// representation of its UserData struct.
func (u *User) ToFirestoreUpdate() []firestore.Update {
	f := []firestore.Update{
		{Path: "mostRecentProgram", Value: u.MostRecentProgram},
		{Path: "programs", Value: u.Programs},
	}
	return f
}
