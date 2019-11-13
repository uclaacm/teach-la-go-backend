package lib

import "time"

// User: used for getUserInfo calls and the like
type UserData struct {
	MostRecentProgram string   `firestore:"mostRecentProgram"`
	Programs          []string `firestore:"programs"`
}

// Program: used for identifying programs.
type Program struct {
	Code        string    `json:"code"`
	DateCreated time.Time `json:"dateCreated"`
	Language    string    `json:"language"`
	Name        string    `json:"name"`
	Thumbnail   uint16    `json:"thumbnail"`
}
