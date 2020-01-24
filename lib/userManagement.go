package lib

import (
	"context"
	"encoding/json"
	"net/http"

	t "../tools"
)

/**
 * getUserData
 * Parameters:
 * {
 *     uid: ...
 * }
 *
 * Returns: Status 200 with marshalled User object.
 *
 * Acquire the userdoc with the given uid.
 */
func (d *DB) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	var (
		u *User
		userJSON []byte
		err error
	)

	// if the current request does not have an User struct
	// in its context (e.g. referred from createUser), then
	// acquire the User struct assuming the uid was provided
	// in the request body.
	if ctxUser := r.Context().Value("user"); ctxUser == nil {
		// attempt to acquire UID from request body.
		if err := t.RequestBodyTo(r, u); err != nil {
			http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
			return
		}

		// attempt to get the complete user struct.
		u, err = d.GetUser(r.Context(), u.UID)
		if err != nil {
			http.Error(w, "error occurred in reading document.", http.StatusInternalServerError)
			return
		}
	} else if _, isUser := ctxUser.(*User); isUser {
		// otherwise, the current request has an User struct in its context.
		// proceed with that user.
		u = ctxUser.(*User)
	}

	// convert to JSON.
	if userJSON, err = json.Marshal(u); err != nil {
		http.Error(w, "error occurred in writing response.", http.StatusInternalServerError)
		return
	}

	// return the user data as JSON.
	w.Write(userJSON)
}

/**
 * updateUserData
 * Parameters:
 * {
 *     [User object]
 * }
 *
 * Returns: Status 200 on success.
 *
 * Merges the JSON passed to it in the request body
 * with userDoc uid.
 */
func (d *DB) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// unmarshal request body into an User struct.
	requestObj := User{}
	if err := t.RequestBodyTo(r, &requestObj); err != nil {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}

	uid := requestObj.UID
	if uid == "" {
		http.Error(w, "a uid is required.", http.StatusBadRequest)
		return
	}

	d.UpdateUser(r.Context(), uid, &requestObj)
	w.WriteHeader(http.StatusOK)
}

/**
 * initializeUserData
 * Parameters: none
 *
 * Returns: Status 200 with a marshalled User struct.
 *
 * Creates a new user in the database and returns their data.
 */
func (d *DB) HandleInitializeUser(w http.ResponseWriter, r *http.Request) {
	u, err := d.CreateUser(r.Context())
	if err != nil {
		http.Error(w, "failed to initialize user data.", http.StatusInternalServerError)
		return
	}

	// pass control to getUserData.
	ctx := context.WithValue(r.Context(), "user", u)
	d.HandleGetUser(w, r.WithContext(ctx))
}
