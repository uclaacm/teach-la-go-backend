package db

import (
	"encoding/json"
	"net/http"

	"../tools/requests"
)

// HandleCreateClass is the handler for creating a new class.
// It takes the UID of the creator, the name of the class, and a thunmbnail id. 
func (d *DB) HandleCreateClass(w http.ResponseWriter, r *http.Request) {
	
	var (
		err error
	)

	//create an anonymous structure to handle requests
	req := struct {
		UID 		string  	`json:"uid"`
		Name 		string		`json:"name"`
		Thumbnail 	int64 		`json:"thumbnail"`
	}{}
	//read JSON from request body
	if err = requests.BodyTo(r, &req); err != nil {
		http.Error(w, "Error occurred in reading body.", http.StatusInternalServerError)
		return
	}

	if req.UID == "" {
		http.Error(w, "Error occurred in reading UID.", http.StatusInternalServerError)
		return
	}
	
	if req.Name == "" {
		http.Error(w, "Error occurred in reading Name.", http.StatusInternalServerError)
		return
	}

	if req.Thumbnail < 0 || req.Thumbnail >= ThumbnailCount  {
		http.Error(w, "Bad thumbnail provided, Exiting", http.StatusInternalServerError)
		return
	}

	

	// structure for class info
	class := Class{
		Thumbnail: req.Thumbnail, 
		Name: req.Name, 
		Creator: req.UID, 
		Instructors: []string{req.UID},
		Members: []string{},
		Programs: []string{},
	}	

	

	// create the class
	cid, err := d.CreateClass(r.Context(), &class)
	if err != nil {
		http.Error(w, "Error updating class in Firebase", http.StatusInternalServerError)
		return
	}

	//create an wid for this class
	wid, err := d.MakeAlias(r.Context(), cid, ClassesAliasPath)

	//Update class info 
	// create the class
	err = d.UpdateClassWID(r.Context(), cid, wid)

	//add this class to the user's "Classes" list
	err = d.AddClassToUser(r.Context(), req.UID, cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// read the class document just created 
	c, err := d.GetClass(r.Context(), cid)
	if err != nil || c == nil {
		http.Error(w, "Class does not exist", http.StatusNotFound)
		return
	}
	//return the class struct in the response
	if resp, err := json.Marshal(c); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}
}

// HandleGetClass takes the UID (either of a member or an instructor) 
// and a CID (wid) as a JSON, and returns an object representing the class. 
// If the given UID is not a member or an instructor, error is returned
func (d *DB) HandleGetClass(w http.ResponseWriter, r *http.Request) {
	
	var err error

	//create an anonymous structure to handle requests
	req := struct {
		UID 		string  	`json:"uid"`
		WID 		string		`json:"cid"`
	}{}

	//read JSON from request body
	if err = requests.BodyTo(r, &req); err != nil {
		http.Error(w, "Error occurred in reading body", http.StatusInternalServerError)
		return
	}
	if req.UID == "" {
		http.Error(w, "Error occurred in reading uid", http.StatusInternalServerError)
		return
	}
	if req.WID == "" {
		http.Error(w, "Error occurred in reading cid", http.StatusInternalServerError)
		return
	}

	cid, err := d.GetUIDFromWID(r.Context(), req.WID, ClassesAliasPath)

	// get the class as a struct (pointer)
	c, err := d.GetClass(r.Context(), cid)

	// check for error
	if err != nil || c == nil {
		http.Error(w, "Failed to get class (class does not exist or failed to unmarshal data)", http.StatusNotFound)
		return
	}

	//check if the uid exists in the members list or instructor list
	is_in := false;

	for _, m := range c.Members {
		if m == req.UID {
			is_in = true
			break
		}
	}

	for _, i := range c.Instructors{
		if i == req.UID {
			is_in = true
			break
		}
	}
	
	// if UID was not in class, return error
	if !is_in {
		http.Error(w, "Couldn't find user in class", http.StatusInternalServerError)
		return
	}

	// otherwise, convert the class struct into JSON and send it back
	if resp, err := json.Marshal(c); err != nil {
		http.Error(w, "Failed to marshal response.", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}
}

// HandleJoinClass takes a UID and cid(wid) as a JSON, and attempts to 
// add the UID to the class given by cid. The updated struct of the class is returned as a 
// JSON

func (d *DB) HandleJoinClass(w http.ResponseWriter, r *http.Request) {

	var err error

	//create an anonymous structure to handle requests
	req := struct {
		UID 		string  	`json:"uid"`
		WID 		string		`json:"cid"`
	}{}

	//read JSON from request body
	if err = requests.BodyTo(r, &req); err != nil {
		http.Error(w, "Error occurred in reading body", http.StatusInternalServerError)
		return
	}
	if req.UID == "" {
		http.Error(w, "Error occurred in reading UID", http.StatusInternalServerError)
		return   
	}
	if req.WID == "" {
		http.Error(w, "error occurred in reading WID", http.StatusInternalServerError)
		return
	}

	//TODO
	cid, err := d.GetUIDFromWID(r.Context(), req.WID, ClassesAliasPath)

	// get the class as a struct
	c, err := d.GetClass(r.Context(), cid)
	// check for error
	if err != nil || c == nil {
		http.Error(w, "Class does not exist.", http.StatusNotFound)
		return
	}

	//check if the user exists
	_, err = d.GetUser(r.Context(), req.UID)
	if err != nil {
		http.Error(w, "User does not exist.", http.StatusNotFound)
		return
	}

	//add user to the class
	err = d.AddUserToClass(r.Context(), req.UID, cid)
	if err != nil {
		http.Error(w, "Failed to add user to class", http.StatusNotFound)
		return
	}

	//add this class to the user's "Classes" list
	err = d.AddClassToUser(r.Context(), req.UID, cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get the updated class as a struct
	c, err = d.GetClass(r.Context(), cid)
	if err != nil || c == nil {
		http.Error(w, "Class does not exist.", http.StatusNotFound)
		return
	}

	// return the updated class struct as JSON
	if resp, err := json.Marshal(c); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}

}


func (d *DB) HandleLeaveClass(w http.ResponseWriter, r *http.Request) {

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

	//check if the user exists
	_, err = d.GetUser(r.Context(), req.UID)
	if err != nil {
		http.Error(w, "user does not exist.", http.StatusNotFound)
		return
	}

	//remove user from the class
	err = d.RemoveUserFromClass(r.Context(), req.UID, req.CID)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusNotFound)
		return
	}

	//remove cid from user list
	err = d.RemoveClassFromUser(r.Context(), req.UID, req.CID)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusNotFound)
		return
	}

	// return the latest state of the user
	u, err := d.GetUser(r.Context(), req.UID)

	if resp, err := json.Marshal(u); err != nil {
		http.Error(w, "failed to marshal response.", http.StatusInternalServerError)
	} else {
		w.Write(resp)
	}

}
