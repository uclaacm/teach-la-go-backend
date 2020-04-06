package db

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/uclaacm/teach-la-go-backend/tools/requests"
)

/**
 * getUserData
 * Acquire the userdoc with the given uid.
 *
 * Query Parameters:
 *  - id string: ID of user to get
 *  - programs bool: whether to include user's programs in response.
 *
 * Returns: Status 200 with marshalled User and optional programs.
 *
 * Example Response:
 * {
 *   "userData": [User object]
 *   "programs": [array of Program objects]
 * }
 */
func (d *DB) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	var (
		u        *User
		userJSON []byte
		err      error
	)

	query := r.URL.Query()

	// if the current request does not have an User struct
	// in its context (e.g. referred from createUser), then
	// acquire the User struct assuming the uid was provided
	// in the request body.
	if ctxUser := r.Context().Value("user"); ctxUser == nil {
		// attempt to get the complete user struct using URL
		// params.
		u, err = d.GetUser(r.Context(), query.Get("id"))
		if err != nil {
			http.Error(w, "error occurred in reading document.", http.StatusInternalServerError)
			return
		}
	} else if _, isUser := ctxUser.(*User); isUser {
		// otherwise, the current request has an User struct in its context.
		// proceed with that user.
		u = ctxUser.(*User)
	}

	// response structure
	resp := struct {
		UserData *User     `json:"userData"`
		Programs []Program `json:"programs"`
	}{UserData: u}

	// retrieve user's program information if requested.
	if query.Get("programs") == "true" {
		// get all the programs for the user.
		for _, pid := range u.Programs {
			// attempt to get program, failing if we cannot.
			p, err := d.GetProgram(r.Context(), pid)
			if err != nil {
				http.Error(w, "error occurred in retrieving programs.", http.StatusInternalServerError)
				return
			}

			// append to response.
			resp.Programs = append(resp.Programs, *p)
		}
	}

	// convert to JSON.
	if userJSON, err = json.Marshal(resp); err != nil {
		http.Error(w, "error occurred in writing response.", http.StatusInternalServerError)
		return
	}

	// return the user data as JSON.
	w.WriteHeader(http.StatusOK)
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
	if err := requests.BodyTo(r, &requestObj); err != nil {
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
