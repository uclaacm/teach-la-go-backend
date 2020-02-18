package lib

import (
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
func (d *DB) HandleCreateClass(w http.ResponseWriter, r *http.Request) {
	
	var (
		err error
	)

	//create an anonymous structure to handle requests
	req := struct {
		Uid 		string  	`json:"uid"`
		Name 		string		`json:"name"`
		Thumbnail 	int64 		`json:"thumbnail"`
	}{}

	//read JSON from request body
	if err = requests.BodyTo(r, &req); err != nil {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}
	if req.Uid == "" {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}
	if req.Name == "" {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}

	if req.Thumbnail < 0 || req.Thumbnail >= 50 {
		http.Error(w, "Bad thumbnail provided, Exiting", http.StatusInternalServerError)
		return
	}

	// structure for class info
	class := Class{
		Thumbnail: req.Thumbnail, 
		Name: req.Name, 
		Creator: req.Uid, 
		Instructors: []string{req.Uid},
		Members: []string{},
		Programs: []string{},
		CID: "",
	}
	
	
	// TODO create id using words, not hash
	//create the class
	cid, err := d.CreateClass(r.Context(), &class)

	//add this class to the user's "Classes" list
	err = d.AddClassToUser(r.Context(), req.Uid, cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return the result of this handler

	// read the class document just created and put it into a struct
	c, err := d.GetClass(r.Context(), cid)

	if err != nil || c == nil {
		http.Error(w, "class does not exist.", http.StatusNotFound)
		return
	}

	if resp, err := json.Marshal(c); err != nil {
		http.Error(w, "failed to marshal response.", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}
}

func (d *DB) HandleGetClass(w http.ResponseWriter, r *http.Request) {
	
	var (
		err error
	)

	//create an anonymous structure to handle requests
	req := struct {
		UID 		string  	`json:"uid"`
		CID 		string		`json:"cid"`
	}{}

	//read JSON from request body
	if err = requests.BodyTo(r, &req); err != nil {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}
	if req.UID == "" {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}
	if req.CID == "" {
		http.Error(w, "error occurred in reading body.", http.StatusInternalServerError)
		return
	}

	// get the class as a struct
	c, err := d.GetClass(r.Context(), req.CID)

	// check for error
	if err != nil || c == nil {
		http.Error(w, "class does not exist.", http.StatusNotFound)
		return
	}

	//check if the uid exists in the members list
	var is_member bool = false;

	for _, m := range c.Members {
		if m == req.UID {
			is_member = true
			break
		}
	}
	
	if !is_member {
		http.Error(w, "failed to marshal response.", http.StatusInternalServerError)
		return
	}

	if resp, err := json.Marshal(c); err != nil {
		http.Error(w, "failed to marshal response.", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}
}