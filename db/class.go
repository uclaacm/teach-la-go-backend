package db

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Class is a struct representation of a class document.
// It provides functions for converting the struct
// to firebase-digestible types.
type Class struct {
	Thumbnail   int64    `firestore:"thumbnail" json:"thumbnail"`
	Name        string   `firestore:"name" json:"name"`
	Creator     string   `firestore:"creator" json:"creator"`
	Instructors []string `firestore:"instructors" json:"instructors"`
	Members     []string `firestore:"members" json:"members"`
	Programs    []string `firestore:"programs" json:"programs"`
	CID         string   `firestore:"CID" json:"cid"`
	WID         string   `firestore:"WID" json:"wid"`
}

// AddClassToUser takes a uid and a pid,
// and adds the pid to the user's list of programs
func (d *DB) AddClassToUser(ctx context.Context, uid string, cid string) error {
	//get the user doc
	ref := d.Collection(usersPath).Doc(uid)

	//add the class id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "classes", Value: firestore.ArrayUnion(cid)},
		})
	})
}

// AddUserToClass add an uid to a given class
func (d *DB) AddUserToClass(ctx context.Context, uid string, cid string) error {
	//get the class doc
	ref := d.Collection(classesPath).Doc(cid)

	//add the user id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "members", Value: firestore.ArrayUnion(uid)},
		})
	})
}

// RemoveUserFromClass removes an uid from a given class
func (d *DB) RemoveUserFromClass(ctx context.Context, uid string, cid string) error {
	//get the class doc
	ref := d.Collection(classesPath).Doc(cid)

	//remove the user id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "members", Value: firestore.ArrayRemove(uid)},
		})
	})
}

// RemoveClassFromUser removes a class from a given user
func (d *DB) RemoveClassFromUser(ctx context.Context, uid string, cid string) error {
	//get the user doc
	ref := d.Collection(usersPath).Doc(uid)

	//remove the class id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "classes", Value: firestore.ArrayRemove(cid)},
		})
	})

}

// loadClass takes a cid, and returns a Class struct with its parameters populated
// The returned value is a pointer to the struct instantiated in this function
func (d *DB) loadClass(ctx context.Context, cid string) (*Class, error) {
	// get the class doc
	doc, err := d.Collection(classesPath).Doc(cid).Get(ctx)
	if err != nil {
		return nil, err
	}

	// populate
	c := &Class{}
	if err := doc.DataTo(&c); err != nil {
		return nil, err
	}
	return c, err
}

// CreateClass is the handler for creating a new class.
// It takes the UID of the creator, the name of the class,
// and a thumbnail id.
func (d *DB) CreateClass(c echo.Context) error {
	// create an anonymous structure to handle requests
	req := struct {
		UID       string `json:"uid"`
		Name      string `json:"name"`
		Thumbnail int64  `json:"thumbnail"`
	}{}

	// read JSON from request body
	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	switch {
	case req.UID == "":
		return c.String(http.StatusBadRequest, "uid is required")
	case req.Name == "":
		return c.String(http.StatusBadRequest, "class name is required")
	case req.Thumbnail < 0 || req.Thumbnail >= thumbnailCount:
		return c.String(http.StatusBadRequest, "bad thumbnail id")
	}

	// structure for class info
	class := Class{
		Thumbnail:   req.Thumbnail,
		Name:        req.Name,
		Creator:     req.UID,
		Instructors: []string{req.UID},
		Members:     []string{},
		Programs:    []string{},
	}

	// create a new doc for this class
	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		ref := d.Collection(classesPath).NewDoc()
		class.CID = ref.ID // set the CID parameter
		return tx.Set(ref, class)
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not create class doc")
	}

	// create an wid for this class
	wid, err := d.MakeAlias(c.Request().Context(), class.CID, classesAliasPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		// update wid field
		cref := d.Collection(classesPath).Doc(class.CID)
		return tx.Update(cref, []firestore.Update{
			{Path: "WID", Value: wid},
		})
	}); err != nil {
		return c.String(http.StatusInternalServerError, "failed to create class alias")
	}

	class.WID = wid

	//add this class to the user's "Classes" list
	err = d.AddClassToUser(c.Request().Context(), req.UID, class.CID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to join user to class")
	}

	//return the class struct in the response
	return c.JSON(http.StatusOK, class)
}

// JoinClass takes a UID and cid(wid) as a JSON, and attempts to
// add the UID to the class given by cid. The updated struct of the class is returned as a
// JSON
// func (d *DB) JoinClass(c echo.Context) error {
// 	req := struct {
// 		UID string `json:"uid"`
// 		CID string `json:"cid"`
// 	}{}

// 	// read JSON from request body
// 	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	if req.UID == "" {
// 		return c.String(http.StatusBadRequest, "uid is required")
// 	}
// 	if req.CID == "" {
// 		return c.String(http.StatusBadRequest, "cid is required")
// 	}

// 	// get the class as a struct
// 	class, err := d.loadClass(c.Request().Context(), req.CID)
// 	if err != nil || class == nil {
// 		return c.String(http.StatusNotFound, "class does not exist")
// 	}

// 	// check if user exists
// 	if err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
// 		userDoc, err := tx.Get(d.Collection(usersPath).Doc(req.UID))
// 		if err != nil {
// 			return err
// 		}
// 		if !userDoc.Exists() {
// 			return errors.New("user does not exist")
// 		}
// 		return nil
// 	}); err != nil {
// 		return c.String(http.StatusNotFound, "user does not exist")
// 	}

// 	// add user to the class
// 	err = d.AddUserToClass(c.Request().Context(), req.UID, req.CID)
// 	if err != nil {
// 		return c.String(http.StatusNotFound, "failed to add user to class")
// 	}

// 	// add this class to the user's "Classes" list
// 	err = d.AddClassToUser(c.Request().Context(), req.UID, req.CID)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to add user to class list").Error())
// 	}

// 	return c.JSON(http.StatusOK, class)
// }

// LeaveClass takes a UID and CID through the request body, and
// attempts to remove user UID from the provided class CID.
func (d *DB) LeaveClass(c echo.Context) error {
	var req struct {
		UID string `json:"uid"`
		CID string `json:"cid"`
	}

	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.UID == "" {
		return c.String(http.StatusBadRequest, "uid is required")
	}
	if req.CID == "" {
		return c.String(http.StatusBadRequest, "cid is required")
	}

	class, err := d.loadClass(c.Request().Context(), req.CID)
	if err != nil || class == nil {
		return c.String(http.StatusNotFound, "class does not exist")
	}

	// check if user exists
	err = d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		_, err := tx.Get(d.Collection(usersPath).Doc(req.UID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "user does not exist")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "unexecpted error occurred!").Error())
	}

	// remove user from the class
	err = d.RemoveUserFromClass(c.Request().Context(), req.UID, req.CID)
	if err != nil {
		return c.String(http.StatusNotFound, errors.Wrap(err, "failed to remove user from class").Error())
	}

	// remove cid from user list
	err = d.RemoveClassFromUser(c.Request().Context(), req.UID, req.CID)
	if err != nil {
		return c.String(http.StatusNotFound, errors.Wrap(err, "failed to remove class ID from user").Error())
	}

	// return the latest state of the user
	return c.String(http.StatusOK, "")
}

// GetClassMembers returns the user IDs and display names of each member in the requested class.
// It takes a class id and checks that the given uid is in the class first
func (d *DB) GetClassMembers(c echo.Context) error {
	var req struct {
		CID string `json:"cid"`
		UID string `json:"uid"`
	}

	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.UID == "" || req.CID == "" {
		return c.String(http.StatusBadRequest, "uid and cid fields are both required")
	}
	uid := req.UID
	cid := req.CID

	// get the class as a struct (pointer)
	class, err := d.loadClass(c.Request().Context(), cid)
	if err != nil || class == nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("failed to get class: %s", err))
	}

	// Check if user is in class
	isIn := false
	for _, m := range class.Members {
		if m == uid {
			isIn = true
			break
		}
	}

	if !isIn {
		return c.String(http.StatusNotFound, "given user not in class")
	}

	res := make(map[string]User)

	for _, uid := range class.Members {
		userSnap, err := d.Collection(usersPath).Doc(uid).Get(c.Request().Context())
		if err != nil {
			continue
		}

		tmpUser := User{}
		err = userSnap.DataTo(&tmpUser)
		if err != nil {
			continue
		}

		res[uid] = tmpUser
	}

	// convert to JSON and return
	return c.JSON(http.StatusOK, res)
}
