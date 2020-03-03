package db

import (
	"context"
	"encoding/json"
	"net/http"

	"../tools/requests"
)

/**
 * getProgram
 * Query parameters: programId
 *
 * Returns: Status 200 with a marshalled Program struct.
 */
func (d *DB) HandleGetProgram(w http.ResponseWriter, r *http.Request) {
	var pid string

	// attempt to acquire program from request context.
	// if missing, then check query parameters.
	if ctxID, ok := r.Context().Value("getProgram").(string); ok {
		pid = ctxID
	} else {
		pid = r.URL.Query().Get("programId")
	}

	// attempt to acquire doc.
	p, err := d.GetProgram(r.Context(), pid)

	// check that the pid is present and that the program exists.
	if err != nil || p == nil {
		http.Error(w, "program does not exist.", http.StatusNotFound)
		return
	}

	// otherwise, return the marshalled program.
	if resp, err := json.Marshal(&p); err != nil {
		http.Error(w, "failed to marshal response.", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

/**
 * initializeProgramData
 * Body:
 * {
 *   uid: UID for the user the program belongs to
 *   thumbnail: index of desired thumbnail
 *   language: language string
 *   name: name of program
 *   code: [optional code for program]
 * }
 *
 * Returns: Status 200 with a marshalled User struct.
 *
 * Creates a new program in the database and returns its data.
 */
func (d *DB) HandleInitializeProgram(w http.ResponseWriter, r *http.Request) {
	var (
		langCode int
		err      error
	)

	// unmarshal request body into a struct matching
	// what we expect.
	requestBody := Program{}
	if err := requests.BodyTo(r, &requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// check that language exists.
	if langCode, err = LanguageCode(requestBody.Language); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check that user exists.
	u, err := d.GetUser(r.Context(), requestBody.UID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// thumbnail should be within range.
	if requestBody.Thumbnail > ThumbnailCount || requestBody.Thumbnail < 0 {
		http.Error(w, "thumbnail index out of bounds.", http.StatusBadRequest)
		return
	}

	// add default code if none provided.
	if requestBody.Code == "" {
		requestBody.Code = defaultProgram(langCode).Code
	}

	// create the program doc.
	p, err := d.CreateProgram(r.Context(), &requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// associate program to user.
	if err = d.AddProgramToUser(r.Context(), u.UID, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//d.UpdateUser(r.Context(), u.UID, u)

	// pass control to getProgramData.
	ctx := context.WithValue(r.Context(), "getProgram", p)
	d.HandleGetProgram(w, r.WithContext(ctx))
}
 
/**
 * updateProgramData
 * Body:
 * {
 *     [Program object]
 * }
 *
 * Returns: Status 200 on success.
 *
 * Merges the JSON passed to it in the request body
 * with program uid.
 */
func (d *DB) HandleUpdateProgram(w http.ResponseWriter, r *http.Request) {
	// unmarshal request body into an Program struct.
	requestObj := Program{}
	if err := requests.BodyTo(r, &requestObj); err != nil {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}

	uid := requestObj.UID
	if uid == "" {
		http.Error(w, "a uid is required.", http.StatusBadRequest)
		return
	}

	d.UpdateProgram(r.Context(), uid, &requestObj)
	w.WriteHeader(http.StatusOK)
}

/**
 * deleteProgram
 * Query parameters: uid, pid
 *
 * Deletes the program identified by {pid} from user {uid}.
 */
func (d *DB) HandleDeleteProgram(w http.ResponseWriter, r *http.Request) {
	// acquire parameters.
	uid := r.URL.Query().Get("userId")
	pid := r.URL.Query().Get("programId")

	var (
		// u   *User
		err error
	)

	//TODO: Make this handler atomic

	// attempt to delete program doc.
	if err = d.DeleteProgram(r.Context(), pid); err != nil {
		http.Error(w, "failed to delete program doc.", http.StatusInternalServerError)
		return
	}

	// Remove this program from the user's list
	if err = d.DeleteProgramFromUser(r.Context(), uid, pid); err != nil {
		http.Error(w, "failed updating user's program list.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
