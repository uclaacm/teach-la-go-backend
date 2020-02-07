package lib

import (
	"context"
	"encoding/json"
	//"fmt"
	//"io/ioutil"
	//"log"
	"net/http"
	//"strconv"

	t "../tools"

	//"cloud.google.com/go/firestore"
)

/**
 * getProgram
 * Parameters:
 * {
 *     uid: ...
 * }
 *
 * Returns: Status 200 with marshalled Program object.
 *
 * Acquire the program doc with the given uid.
 */

 func (d *DB) HandleGetProgram(w http.ResponseWriter, r *http.Request) {
	var (
		p *Program
		progJSON []byte
		err error
	)

	// if the current request does not have an Program struct
	// in its context (e.g. referred from createProgram), then
	// acquire the Program struct assuming the uid was provided
	// in the request body.
	if ctxProgram := r.Context().Value("program"); ctxProgram == nil {
		// attempt to acquire UID from request body.
		if err := t.RequestBodyTo(r, p); err != nil {
			http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
			return
		}

		// attempt to get the complete program struct.
		p, err = d.GetProgram(r.Context(), p.UID)
		if err != nil {
			http.Error(w, "error occurred in reading document.", http.StatusInternalServerError)
			return
		}
	} else if _, isProgram := ctxProgram.(*Program); isProgram {
		// otherwise, the current request has a Program struct in its context.
		// proceed with that program.
		p = ctxProgram.(*Program)
	}

	// convert to JSON.
	if progJSON, err = json.Marshal(p); err != nil {
		http.Error(w, "error occurred in writing response.", http.StatusInternalServerError)
		return
	}

	// return the user data as JSON.
	w.Write(progJSON)
 }


/**
 * initializeProgramData
 * Parameters: none
 *
 * Returns: Status 200 with a marshalled User struct.
 *
 * Creates a new user in the database and returns their data.
 */
 func (d *DB) HandleInitializeProgram(w http.ResponseWriter, r *http.Request) {
	 // unmarshal request body into an Program struct.
	requestObj := Program{}
	if err := t.RequestBodyTo(r, &requestObj); err != nil {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}
	// put program struct into db
	p, err := d.CreateProgram(r.Context(), &requestObj) 
	if err != nil {
		http.Error(w, "failed to initialize program data.", http.StatusInternalServerError)
		return
	}

	// pass control to getProgramData.
	ctx := context.WithValue(r.Context(), "program", p)
	d.HandleGetProgram(w, r.WithContext(ctx))
}


/**
 * updateProgramData
 * Parameters:
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
	if err := t.RequestBodyTo(r, &requestObj); err != nil {
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
 * Parameters:
 * {
 *     pid: ...
 * }
 *
 *
 * Deletes the program identified by {pid}. Did not make {uid} required. 
 */
 func (d *DB) HandleDeleteProgram(w http.ResponseWriter, r *http.Request) {

	//if _,err = d.getProgram(r.Context(), r.Context().Value("pid"))
	var pr = Program{}
	p := &pr
	// if the current request does not have an Program struct
	// in its context, then acquire the Program struct assuming 
	// the uid was provided in the request body.
	if ctxProgram := r.Context().Value("program"); ctxProgram == nil {
		// program does not exist
		// attempt to acquire UID from request body.
		if err := t.RequestBodyTo(r, p); err != nil {
			http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
			return
		}

		// attempt to get the complete program struct.
		_, err := d.GetProgram(r.Context(), p.UID)
		if err != nil {
			http.Error(w, "error occurred in reading document.", http.StatusInternalServerError)
			return
		}

	} else if _, isProgram := ctxProgram.(*Program); isProgram {
		// program does exist
		// otherwise, the current request has a Program struct in its context.
		// proceed with that program.
		p = ctxProgram.(*Program)
	}

	d.DeleteProgram(r.Context(), p.UID)
 }
