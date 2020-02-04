package lib

import (
	"../logger"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleUsers manages all requests pertaining to
// user information.
type HandleUsers struct {
	Client *firestore.Client
}

// HandleUsers.ServeHTTP is used by net/http to serve
// endpoints in accordance with the handler.
// Requests are handled appropriately based on request
// type.
func (h HandleUsers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// parse the /:uid field
	uid := r.URL.Path[len("/userData/"):]

	// parse the response body into a map[string]string
	var (
		bytesBody []byte
		err       error
	)
	body := make(map[string]string)
	if bytesBody, err = ioutil.ReadAll(r.Body); err != nil {
		logger.Errorf("failed to read request body.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if err := json.Unmarshal(bytesBody, &body); len(bytesBody) > 0 && err != nil {
		logger.Errorf("failed to marshal request body. %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// catch bad requests
	if uid == "" && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// handle based on request method
	var (
		response string
		code     int
	)
	switch r.Method {
	case http.MethodGet:
		response, code = h.getUserData(uid)

	case http.MethodPut:
		response, code = h.updateUserData(uid, bytesBody)

	case http.MethodPost:
		response, code = h.initializeUserData()
	}

	// handle errors.
	if code != http.StatusOK {
		logger.Errorf(response)
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

/**
 * getUserData
 * GET /userData/:uid
 *
 * Returns the UserData object for user {uid} in JSON.
 */
func (h *HandleUsers) getUserData(uid string) (string, int) {
	// acquire userdoc corresponding to uid
	userDoc, err := h.Client.Collection(UsersPath).Doc(uid).Get(context.Background())

	// catch document errors.
	if status.Code(err) == codes.NotFound {
		return "document does not exist.", http.StatusNotFound
	} else if err != nil {
		return fmt.Sprintf("error occurred in document retrieval: %s", err), http.StatusInternalServerError
	}

	// acquire only desired fields for response.
	var u User
	if err = userDoc.DataTo(&u); err != nil {
		return "error occurred in reading document.", http.StatusInternalServerError
	}

	// convert to JSON.
	var userJSON []byte

	// optional: pretty print our JSON response.
	userJSON, err = json.MarshalIndent(u, "", "    ")
	if err != nil {
		return "error occurred in reading document.", http.StatusInternalServerError
	}

	// return the user data as JSON.
	return string(userJSON), http.StatusOK
}

/**
 * updateUserData
 * PUT /userData/:uid {Body: JSON user data}
 *
 * Merges the JSON passed to it in the request body
 * with userDoc :uid
 */
func (h *HandleUsers) updateUserData(uid string, bytesBody []byte) (string, int) {
	// get userDoc
	userDoc := h.Client.Collection(UsersPath).Doc(uid)

	// unmarshal into an UserData struct.
	requestObj := User{}
	json.Unmarshal(bytesBody, &requestObj)

	// ensure all fields were filled.
	updateData := requestObj.ToFirestoreUpdate()

	if len(updateData) == 0 {
		return "missing fields from request.", http.StatusBadRequest
	}

	_, err := userDoc.Update(context.Background(), updateData)

	// check for errors.
	if status.Code(err) == codes.NotFound {
		return "document does not exist.", http.StatusNotFound
	} else if err != nil {
		return fmt.Sprintf("error occurred in document retrieval: %s", err), http.StatusInternalServerError
	}

	return "", http.StatusOK
}

/**
 * initializeUserData
 * POST /userData
 *
 * Creates a new user in the database and returns their data.
 */
func (h *HandleUsers) initializeUserData() (string, int) {
	newDoc := h.Client.Collection(UsersPath).NewDoc()

	newUser, newProgs := defaultData()

	// create all new programs and associate them to the user.
	for _, prog := range newProgs {
		// create program in database.
		newProg := h.Client.Collection(UsersPath).NewDoc()
		newProg.Set(context.Background(), prog)

		// establish association in user doc.
		newUser.Programs = append(newUser.Programs, newProg.ID)
	}

	// create user doc.
	newDoc.Set(context.Background(), newUser)

	result, err := json.Marshal(newUser)
	if err != nil {
		return "failed to initialize user data.", http.StatusInternalServerError
	}

	return string(result), http.StatusOK
}
