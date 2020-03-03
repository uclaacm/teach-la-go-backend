package db

import (
	"fmt"

	"cloud.google.com/go/firestore"
)

// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type User struct {
	Classes           []string `firestore:"classes" json:"classes"`
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	Programs          []string `firestore:"programs" json:"programs"`
	UID               string   `json:"uid"`
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

// AddClass adds a class id 'cid' to an User's class list.
func (u *User) AddClass(cid string) {
	u.Classes = append(u.Classes, cid)
}

// RemoveClass removes a class id 'cid' from an User's
// class list. An error is thrown if the User is not
// a member of that class.
func (u *User) RemoveClass(cid string) error {
	// loop through all the elements.
	// when we find the index of a match, change
	// the array to exclude the element, and then
	// return nil.
	for idx, el := range u.Classes {
		// found the element.
		if el == cid {
			// remove it and update.
			half := u.Classes[:idx]

			// append elements of h2.
			for _, element := range u.Classes[idx:] {
				half = append(half, element)
			}

			u.Classes = half
			return nil
		}
	}

	// if we can't find it, throw an error.
	return fmt.Errorf("failed to find %s in User classes", cid)
}

// AddProgram adds the program id 'pid' to an User's
// program list.
func (u *User) AddProgram(pid string) {
	u.Programs = append(u.Programs, pid)
}

// DEPRICATED: No longer used
// RemoveProgram removes the program identified from the User's
// array.
func (u *User) RemoveProgram(pid string) error {
	// loop through all the elements.
	// when we find the index of a match, change
	// the array to exclude the element, and then
	// return nil.
	for idx, el := range u.Programs {
		// found the element.
		if el == pid {
			// remove it and update.
			half := u.Programs[:idx]

			// append elements of h2.
			for _, element := range u.Programs[idx:] {
				half = append(half, element)
			}

			u.Programs = half
			return nil
		}
	}

	// if we can't find it, throw an error.
	return fmt.Errorf("failed to find %s in User programs", pid)
}
