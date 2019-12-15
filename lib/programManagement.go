package lib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ThumbnailCount describes the number of program
// thumbnails available to choose from.
const ThumbnailCount = 58

// Program is a representation of a program document.
type Program struct {
	Code        string    `json:"code" firestore:"code"`
	DateCreated time.Time `json:"dateCreated" firestore:"dateCreated"`
	Language    string    `json:"language" firestore:"language"`
	Name        string    `json:"name" firestore:"name"`
	Thumbnail   uint16    `json:"thumbnail" firestore:"thumbnail"`
}

// defaultProgram returns a Program struct initialized to
// default values for a given language.
func defaultProgram(language string) Program {
	var defaultCode string

	defaultProg := Program{
		DateCreated: time.Now().UTC(),
		Language:    language,
	}

	switch language {
	case "python":
		defaultCode = "import turtle\n\nt = turtle.Turtle()\n\nt.color('red')\nt.forward(75)\nt.left(90)\n\n\nt.color('blue')\nt.forward(75)\nt.left(90)\n"
	case "processing":
		defaultCode = "function setup() {\n  createCanvas(400, 400);\n}\n\nfunction draw() {\n  background(220);\n  ellipse(mouseX, mouseY, 100, 100);\n}"
	case "html":
		defaultCode = "<html>\n  <head>\n  </head>\n  <body>\n    <div style='width: 100px; height: 100px; background-color: black'>\n    </div>\n  </body>\n</html>"
	default:
		log.Printf("received request to create default program for unsupported language: %s.", language)
	}

	defaultProg.Code = defaultCode

	return defaultProg
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
		h.createProgram(w, r)

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
func (h *HandlePrograms) createProgram(w http.ResponseWriter, r *http.Request) {
	// createProgramRequest describes the anticipated JSON of
	// a request to the POST /programs/ endpoint.
	type createProgramRequest struct {
		UID       string `json:"uid"`
		Name      string `json:"name"`
		Thumbnail uint16 `json:"thumbnail"`
		Language  string `json:"language"`
	}

	// createProgramResponse describes the anticipated JSON of
	// a response from the POST /programs/ endpoint.
	type createProgramResponse struct {
		ProgramData Program `json:"programData"`
		UID         string  `json:"key"`
	}

	// get JSON body from request.
	var err error
	var bytesBody []byte
	var body createProgramRequest
	if bytesBody, err = ioutil.ReadAll(r.Body); err == nil {
		json.Unmarshal(bytesBody, &body)
	} else {
		log.Printf("error: could not read from request body properly.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check params.
	var missingFields []string
	if body.UID == "" {
		missingFields = append(missingFields, "UID")
	} else if body.Name == "" {
		missingFields = append(missingFields, "name")
	} else if body.Language == "" {
		missingFields = append(missingFields, "language")
	} else if body.Thumbnail >= ThumbnailCount {
		missingFields = append(missingFields, "thumbnail")
	}

	if len(missingFields) != 0 {
		log.Printf("error: request missing %s fields.", missingFields)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ensure that corresponding userdoc exists.
	userDoc, err := h.Client.Collection("users").Doc(body.UID).Get(r.Context())
	if !userDoc.Exists() {
		log.Printf("bad request: user with UID %s does not exist.", body.UID)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("error: failed to acquire userdoc. %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// initialize a program to match params supplied,
	// filling in default values when applicable.
	programData := defaultProgram(body.Language)
	if err := json.Unmarshal(bytesBody, &programData); err != nil {
		log.Printf("error: failed to unmarshal programData. %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create the new program document.
	programDoc := h.Client.Collection("programs").NewDoc()

	// write to programDoc and update associated user.
	programDoc.Set(r.Context(), &programData)
	h.Client.Collection("users").Doc(body.UID).Update(r.Context(), []firestore.Update{{Path: "programs", Value: firestore.ArrayUnion(programDoc.ID)}})
	log.Printf("created new program doc {%s} associated to user {%s}.", programDoc.ID, body.UID)

	// write response, return OK if nominal.
	response := createProgramResponse{
		ProgramData: programData,
		UID:         programDoc.ID,
	}
	if updateData, err := json.Marshal(response); err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(updateData)
	} else {
		log.Printf("error: failed to marshal createProgram response. %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

/**
 * updatePrograms
 * PUT /programs/:uid
 *
 * Updates the programs for the current user with {uid}.
 */
func (h *HandlePrograms) updatePrograms(w http.ResponseWriter, r *http.Request, pid string) {
	w.WriteHeader(http.StatusNotImplemented)
}

/**
 * deletePrograms
 * DELETE /programs/:uid
 *
 * Deletes the program with {name} from user {uid}.
 */
func (h *HandlePrograms) deletePrograms(w http.ResponseWriter, r *http.Request, pid string) {
	// acquire userdoc
	userDoc := h.Client.Collection("users").Doc(pid)
	userData, err := userDoc.Get(r.Context())
	if err != nil {
		log.Printf("error: failed to acquire user doc. %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse body to map, failing request if server fails any step.
	body := make(map[string]string)
	if bodyBytes, err := ioutil.ReadAll(r.Body); err == nil {
		if err = json.Unmarshal(bodyBytes, &body); err != nil {
			log.Printf("error: failed to parse body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// "name" field is required.
	programID := body["name"]
	if programID == "" {
		log.Printf("bad request: body missing program name.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// acquire progdoc.
	progDoc := h.Client.Collection("programs").Doc(programID)
	progData, err := progDoc.Get(r.Context())
	if err != nil {
		log.Printf("error: failed to acquire program doc.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check that both userdoc and progdoc exist.
	if !userData.Exists() || !progData.Exists() {
		log.Printf("bad request: userDoc or progDoc do not exist.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete the program from userdoc program list.
	userDoc.Update(r.Context(), []firestore.Update{{Path: "programs", Value: firestore.ArrayRemove(programID)}})

	// delete the program from the programs collection.
	progDoc.Delete(r.Context())

	w.WriteHeader(http.StatusOK)
}
