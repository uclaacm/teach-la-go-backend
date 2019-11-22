package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserData
// Go representation of a user document.
type UserData struct {
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	Programs          []string `firestore:"programs" json:"programs"`
	Classes           []string `firestore:"classes" json:"classes"`
}

// defaultUserData
// Factory function for default user data objects.
func defaultUserData() *UserData {
	u := UserData{DisplayName: "J Bruin"}
	return &u
}

// ToFirebaseUpdate
// Returns the Firebase update representation of this struct.
func (u *UserData) ToFirebaseUpdate() []firestore.Update {
	f := []firestore.Update{
		{Path: "mostRecentProgram", Value: u.MostRecentProgram},
		{Path: "programs", Value: u.Programs},
	}
	return f
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
// PUT /userData/:uid {Body: JSON user data}
// Merges user data with the JSON passed to it in the request body.
func (h *HandleUsers) updateUserData(w http.ResponseWriter, r *http.Request, uid string) {
	// get userDoc
	userDoc := h.Client.Collection("users").Doc(uid)

	// parse data into object.
	requestData, err := ioutil.ReadAll(r.Body)

	// check for errors.
	if err != nil {
		log.Printf("failed in reading request body: %s", err)
		http.Error(w, "error occurred in reading request body.", http.StatusInternalServerError)
		return
	}
	if requestData == nil {
		http.Error(w, "nothing to update.", http.StatusBadRequest)
		return
	}

	// unmarshal into an UserData struct.
	requestObj := UserData{}
	json.Unmarshal(requestData, &requestObj)

	// ensure all fields were filled.
	updateData := requestObj.ToFirebaseUpdate()

	if len(updateData) == 0 {
		http.Error(w, "missing fields from request.", http.StatusBadRequest)
	}

	_, err = userDoc.Update(r.Context(), updateData)

	// check for errors.
	if status.Code(err) == codes.NotFound {
		log.Printf("document does not exist.")
		http.Error(w, "document does not exist.", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("error occurred in document retrieval: %s", err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// initializeUserData
// POST /userData
// Creates a new user in the database and returns their data.
func (h *HandleUsers) initializeUserData(w http.ResponseWriter, r *http.Request) {
	newDoc := h.Client.Collection("users").NewDoc()

	newUser := defaultUserData()

	newDoc.Set(r.Context(), newUser)

	result, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("error: failed marshalling new user object: %s", err)
		http.Error(w, "failed to initialize user data.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
