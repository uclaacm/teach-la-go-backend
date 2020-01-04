package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
)

// HandlePrograms manages all requests pertaining to
// program information.
type HandlePrograms struct {
	Client *firestore.Client
}

// HandlePrograms.ServeHTTP is used by net/http to serve
// endpoints in accordance with the handler.
// Requests are handled appropriately based on request
// type.
func (h HandlePrograms) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// acquire pid.
	pid := r.URL.Path[len("/programs/"):]

	// handle based on method.
	var (
		response string
		code     int
	)

	switch r.Method {
	case http.MethodGet:
		response, code = h.getProgram(pid)

	case http.MethodPost:
		// body is map[string]string for create requests.
		body := make(map[string]string)
		if bytesBody, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(bytesBody, body)
		}

		response, code = h.createProgram(body)

	case http.MethodPut:
		// body is map[string]Program for update requests.
		body := make(map[string]Program)
		if bytesBody, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(bytesBody, body)
		}

		response, code = h.updatePrograms(pid, body)

	case http.MethodDelete:
		// body is map[string]string for deletion requests.
		body := make(map[string]string)
		if bytesBody, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(bytesBody, body)
		}

		response, code = h.deletePrograms(pid, body["uid"])
	}

	// log only if an error occurred.
	if code != http.StatusOK {
		log.Println(response)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
	w.WriteHeader(code)
}

/**
 * getProgram
 * GET /programs/:docid
 *
 * Returns the requested document by its {docid}, if it exists,
 * in JSON.
 */
func (h *HandlePrograms) getProgram(pid string) (string, int) {
	// acquire program doc corresponding to its ID.
	progDoc, err := h.Client.Collection(ProgramsPath).Doc(pid).Get(context.Background())

	// catch errors.
	if !progDoc.Exists() {
		return "document does not exist.", http.StatusNotFound
	} else if err != nil {
		return "error occurred in document retrieval.", http.StatusInternalServerError
	}

	// move doc data to struct representation.
	var p Program
	if err = progDoc.DataTo(&p); err != nil {
		log.Printf("error occurred in writing document to %T struct: %s", p, err)
		return "error occurred in document retrieval.", http.StatusInternalServerError
	}

	// move to JSON.
	var progJSON []byte
	if progJSON, err = json.Marshal(p); err != nil {
		log.Printf("failed marshalling struct to JSON: %s", err)
		return "error occurred in document retrieval.", http.StatusInternalServerError
	}

	return string(progJSON), http.StatusOK
}

/**
 * createProgram
 * POST /programs/
 *
 * Creates and returns a program document. Takes the following
 * parameters inside of the request's JSON body:
 * uid: string representing which user we should be updating
 * name: name of the new document
 * language: language the program will use
 * thumbnail: index of the thumbnail picture
 *
 * uid, name, language are required.
 */
func (h *HandlePrograms) createProgram(body map[string]string) (string, int) {
	// check params.
	uid := body["uid"]
	name := body["name"]
	language := body["language"]
	thumbnail, err := strconv.ParseInt(body["thumbnail"], 10, 64)
	if err != nil {
		return fmt.Sprintf("error: improper thumbnail value."), http.StatusBadRequest
	}

	var missingFields []string
	if uid == "" {
		missingFields = append(missingFields, "UID")
	} else if name == "" {
		missingFields = append(missingFields, "name")
	} else if _, err := LanguageCode(language); language == "" || err != nil {
		missingFields = append(missingFields, "language")
	} else if thumbnail >= ThumbnailCount {
		missingFields = append(missingFields, "thumbnail")
	}

	if len(missingFields) != 0 {
		return fmt.Sprintf("error: request missing valid %s fields.", missingFields), http.StatusBadRequest
	}

	// ensure that corresponding userdoc exists.
	userDoc := h.Client.Collection(UsersPath).Doc(uid)
	userDocSnap, err := userDoc.Get(context.Background())
	if !userDocSnap.Exists() {
		return fmt.Sprintf("bad request: user with UID %s does not exist.", uid), http.StatusBadRequest
	} else if err != nil {
		return fmt.Sprintf("error: failed to acquire userdoc. %s", err), http.StatusBadRequest
	}

	// initialize a program to match params supplied,
	// filling in default values when applicable.
	langCode, _ := LanguageCode(language)
	programData := defaultProgram(langCode)
	if name != "" {
		programData.Name = name
	}
	if language != "" {
		programData.Language = language
	}
	if thumbnail >= 0 {
		programData.Thumbnail = thumbnail
	}

	// create the new program document.
	programDoc := h.Client.Collection(ProgramsPath).NewDoc()

	// write to programDoc and update associated user.
	if _, err := programDoc.Set(context.Background(), &programData); err != nil {
		return fmt.Sprintf("error: failed to write to new program document. %s", err), http.StatusInternalServerError
	}
	if _, err := userDoc.Update(context.Background(), []firestore.Update{{Path: "programs", Value: firestore.ArrayUnion(programDoc.ID)}}); err != nil {
		return fmt.Sprintf("error: failed to update user document. %s", err), http.StatusInternalServerError
	}

	log.Printf("created new program doc {%s} associated to user {%s}.", programDoc.ID, uid)

	// write response, return OK if nominal.
	response := map[string]interface{}{
		"ProgramData": programData,
		"UID":         uid,
	}
	if updateData, err := json.Marshal(response); err == nil {
		return string(updateData), http.StatusOK
	}

	return "error: failed to marshal createProgram response.", http.StatusInternalServerError
}

/**
 * updatePrograms
 * PUT /programs/:uid
 *
 * Updates the programs for the current user with {uid}.
 * Takes a map of program IDs to program data within the
 * request body. For example:
 * {
 *   "programID1": { PROGRAM DATA },
 *   "programID2": { PROGRAM DATA }
 * }
 */
func (h *HandlePrograms) updatePrograms(pid string, body map[string]Program) (string, int) {
	// acquire userdoc
	userDoc := h.Client.Collection(UsersPath).Doc(pid)
	userData, err := userDoc.Get(context.Background())
	if err != nil || !userData.Exists() {
		return fmt.Sprintf("error: failed to acquire user doc. %s", err), http.StatusInternalServerError
	}

	// merge documents as appropriate.
	for id, programData := range body {
		// acquire program doc.
		programDoc := h.Client.Collection(ProgramsPath).Doc(id)

		// check for existence.
		if data, err := programDoc.Get(context.Background()); err != nil || !data.Exists() {
			return fmt.Sprintf("error: failed to acquire program doc."), http.StatusInternalServerError
		}

		// update data.
		programDoc.Update(context.Background(), programData.ToFirestoreUpdate())
	}

	return "", http.StatusOK
}

/**
 * deletePrograms
 * DELETE /programs/:pid
 *
 * Deletes the program identified by {pid} from user {uid}. {uid} must
 * be provided in the request body.
 */
func (h *HandlePrograms) deletePrograms(uid string, pid string) (string, int) {
	// acquire userdoc
	userDoc := h.Client.Collection(UsersPath).Doc(pid)
	userData, err := userDoc.Get(context.Background())
	if err != nil {
		return fmt.Sprintf("error: failed to acquire user doc. %s", err), http.StatusInternalServerError
	}

	// the pid and uid are required.
	if pid == "" || uid == "" {
		return "bad request: body missing parameter.", http.StatusBadRequest
	}

	// acquire progdoc.
	progDoc := h.Client.Collection(ProgramsPath).Doc(pid)
	progData, err := progDoc.Get(context.Background())
	if err != nil {
		return "error: failed to acquire program doc.", http.StatusInternalServerError
	}

	// check that both userdoc and progdoc exist.
	if !userData.Exists() || !progData.Exists() {
		return "bad request: userDoc or progDoc do not exist.", http.StatusBadRequest
	}

	// delete the program from userdoc program list.
	if _, err := userDoc.Update(context.Background(), []firestore.Update{{Path: "programs", Value: firestore.ArrayRemove(pid)}}); err != nil {
		return "error: failed to update userDoc.", http.StatusInternalServerError
	}

	// delete the program from the programs collection.
	if _, err := progDoc.Delete(context.Background()); err != nil {
		return "error: failed to delete progDoc.", http.StatusInternalServerError
	}

	return "", http.StatusOK
}
