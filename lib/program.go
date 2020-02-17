package lib

import (
	"time"

	"cloud.google.com/go/firestore"
)

// Program is a representation of a program document.
type Program struct {
	Code        string    `firestore:"code" json:"code"`
	DateCreated time.Time `firestore:"dateCreated" json:"dateCreated"`
	Language    string    `firestore:"language" json:"language"`
	Name        string    `firestore:"name" json:"name"`
	Thumbnail   int64     `firestore:"thumbnail" json:"thumbnail"`
	UID         string    `json:"uid"`
	//	PID			string	  `json:"pid"`
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
