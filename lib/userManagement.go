package lib

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserData
// Go representation of a user document.
type UserData struct {
	MostRecentProgram string   `firestore:"mostRecentProgram"`
	Programs          []string `firestore:"programs"`
}

// HandleUsers
// Manages all requests pertaining to user information.
type HandleUsers struct {
	Client *firestore.Client
}

// HandleUsers.ServeHTTP
// Handle requests appropriately based on request type.
func (h HandleUsers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// parse the /:uid field
	uid := r.URL.Path[len("/userData/"):]

	// catch bad requests
	if uid == "" && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// handle based on request method
	switch r.Method {
	case http.MethodGet:
		h.getUserData(w, r, uid)

	case http.MethodPut:
		h.updateUserData(w, r, uid)

	case http.MethodPost:
		h.initializeUserData(w, r)
	}
}

// getUserData
// GET /userData/:uid
// Used to retrieve a user data object as JSON.
func (h *HandleUsers) getUserData(w http.ResponseWriter, r *http.Request, uid string) {
	// acquire userdoc corresponding to uid
	userDoc, err := h.Client.Collection("users").Doc(uid).Get(r.Context())

	// catch document errors.
	if status.Code(err) == codes.NotFound {
		http.Error(w, "document does not exist.", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("error occurred in document retrieval: %s", err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	// acquire only desired fields for response.
	var u UserData
	if err = userDoc.DataTo(&u); err != nil {
		log.Printf("error occurred in writing document to %T object: %s", u, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// convert to JSON.
	var userJSON []byte

	// optional: pretty print our JSON response.
	userJSON, err = json.MarshalIndent(u, "", "    ")
	if err != nil {
		log.Printf("error occurred in writing document to %T object.", u)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// return the user data as JSON.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(userJSON)
}

// updateUserData
// PUT /updateUserData/:uid
func (h *HandleUsers) updateUserData(w http.ResponseWriter, r *http.Request, uid string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// initializeUserData
// POST /userData
func (h *HandleUsers) initializeUserData(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
