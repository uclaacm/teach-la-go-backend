package lib


// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type Class struct {
	Thumbnail           int64 `firestore:"thumbnail" json:"thumbnail"`
	Name       			string   `firestore:"name" json:"name"`
	Creator 			string   `firestore:"creator" json:"creator"`
	Instructors         []string   `firestore:"instructors" json:"instructors"`
	Members          	[]string `firestore:"members" json:"members"`
	Programs           	[]string   `json:"programs"`
	CID					string 		`json:"cid"`
}


