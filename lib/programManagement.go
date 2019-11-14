package lib

import (
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

// Program
// Go representation of a program document.
type Program struct {
	Code        string    `json:"code"`
	DateCreated time.Time `json:"dateCreated"`
	Language    string    `json:"language"`
	Name        string    `json:"name"`
	Thumbnail   uint16    `json:"thumbnail"`
}

// HandlePrograms
// Manages all requests pertaining to program information.
type HandlePrograms struct {
	Client *firestore.Client
}

// HandlePrograms.ServeHTTP
// Handle requests appropriately based on request type.
func (h HandlePrograms) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Path[len("/programs/"):]

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

/* getProgram
 * GET /programs/:docid
 * Returns the requested document, if it exists.
 */
func (h *HandlePrograms) getProgram(w http.ResponseWriter, r *http.Request, pid string) {
	w.WriteHeader(http.StatusNotImplemented)
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
