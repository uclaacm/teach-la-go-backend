package lib

import (
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Program is a representation of a program document.
type Program struct {
	Code        string `json:"code"`
	DateCreated string `json:"dateCreated"`
	Language    string `json:"language"`
	Name        string `json:"name"`
	Thumbnail   uint16 `json:"thumbnail"`
}

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
	pid := r.URL.Path[len("/programs/"):]

	// handle based on method.
	switch r.Method {
	case http.MethodGet:
		h.getProgram(w, r, pid)

	case http.MethodPost:
		h.createProgram(w, r, pid)

	case http.MethodPut:
		h.updatePrograms(w, r, pid)

	case http.MethodDelete:
		h.deletePrograms(w, r, pid)
	}
}

/**
 * getProgram
 * GET /programs/:docid
 *
 * Returns the requested document by its {docid}, if it exists,
 * in JSON.
 */
func (h *HandlePrograms) getProgram(w http.ResponseWriter, r *http.Request, pid string) {
	// acquire program doc corresponding to its ID.
	progDoc, err := h.Client.Collection("programs").Doc(pid).Get(r.Context())

	// catch errors.
	if status.Code(err) == codes.NotFound {
		http.Error(w, "document does not exist.", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("error occurred in document retrieval: %s", err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	// move doc data to struct representation.
	var p Program
	if err = progDoc.DataTo(&p); err != nil {
		log.Printf("error occurred in writing document to %T struct: %s", p, err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	// move to JSON.
	var progJSON []byte
	if progJSON, err = json.Marshal(p); err != nil {
		log.Printf("failed marshalling struct to JSON: %s", err)
		http.Error(w, "error occurred in document retrieval.", http.StatusInternalServerError)
		return
	}

	// return result.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(progJSON)
}

/* createProgram
 * POST /programs
 * Creates and returns a program document.
 */
func (h *HandlePrograms) createProgram(w http.ResponseWriter, r *http.Request, pid string) {
	w.WriteHeader(http.StatusNotImplemented)
}

/* updatePrograms
 * PUT /programs/:uid
 * Updates the programs for the current user.
 */
func (h *HandlePrograms) updatePrograms(w http.ResponseWriter, r *http.Request, pid string) {
	w.WriteHeader(http.StatusNotImplemented)
}

/* deletePrograms
 * DELETE /programs/:uid
 * Deletes the program with given uid.
 */
func (h *HandlePrograms) deletePrograms(w http.ResponseWriter, r *http.Request, pid string) {
	w.WriteHeader(http.StatusNotImplemented)
}
