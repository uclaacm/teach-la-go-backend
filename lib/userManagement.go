package lib

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"io"

	"cloud.google.com/go/firestore"
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
	uid := r.URL.Path[len("/userData/"):]

	switch (r.Method) {
	case http.MethodGet:
		h.getUserData(w, r, uid)

	case http.MethodPost:
		h.updateUserData(w, r, uid)
		
	case http.MethodPut:
		h.initializeUserData(w, r)
	}
}

// InitializeUserData
// router.GET("/initializeUserData/:uid", ...)
func (h *HandleUsers) initializeUserData(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, string(http.StatusNotImplemented))
}

// GetUserData
// router.GET("/getUserData/:uid", ...)
// Used to retrieve a user object as JSON.
func (h *HandleUsers) getUserData(w http.ResponseWriter, r *http.Request, uid string) {
	// acquire userdoc corresponding to uid
	userDoc, err := h.Client.Collection("users").Doc(uid).Get(context.Background())
	
	if err != nil {
		log.Printf("error occurred in document retrieval: %s", err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	if !userDoc.Exists() {
		http.Error(w, "document does not exist.", http.StatusBadRequest)
		return
	}

	log.Printf("acquired user doc %s.", uid)

	// acquire only desired fields for response.
	var u UserData
	if err = userDoc.DataTo(&u); err != nil {
		log.Printf("error occurred in writing document to %T object: %s", u, err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	// convert to JSON.
	var userJSON []byte
	userJSON, err = json.MarshalIndent(u, "", "    ")
	if err != nil {
		log.Printf("error occurred in writing document to %T object.", u)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	// return the user data as JSON.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(userJSON)
}

// UpdateUserData
// router.POST("/updateUserData/:uid", ...)
func (h *HandleUsers) updateUserData(w http.ResponseWriter, r *http.Request, uid string) {
	io.WriteString(w, string(http.StatusNotImplemented))
}
