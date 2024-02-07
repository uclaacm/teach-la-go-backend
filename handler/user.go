package handler

import (
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type User struct {
	Classes           []string `firestore:"classes" json:"classes"`
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	Programs          []string `firestore:"programs" json:"programs"`
	UID               string   `json:"uid"`
	DeveloperAcc      bool     `firestore:"developerAcc" json:"developerAcc"`
}

// ToFirestoreUpdate returns the database update
// representation of its UserData struct.
func (u *User) ToFirestoreUpdate() []firestore.Update {
	f := []firestore.Update{
		{Path: "mostRecentProgram", Value: u.MostRecentProgram},
	}

	switch {
	case u.DisplayName != "":
		f = append(f, firestore.Update{Path: "displayName", Value: u.DisplayName})
	case u.PhotoName != "":
		f = append(f, firestore.Update{Path: "photoName", Value: u.PhotoName})
	case len(u.Programs) != 0:
		f = append(f, firestore.Update{Path: "programs", Value: firestore.ArrayUnion(u.Programs)})
	}

	return f
}

// GetUser acquires the user document with the given uid. The
// provided context must be a *db.DBContext.
//
// Query Parameters:
//   - uid string: UID of user to GET
//   - programs string: Whether to acquire programs.
//
// Returns: Status 200 with marshalled User and programs.
func GetUser(cc echo.Context) error {
	resp := struct {
		UserData db.User               `json:"userData"`
		Programs map[string]db.Program `json:"programs"`
	}{
		UserData: db.User{},
		Programs: make(map[string]db.Program),
	}

	c := cc.(*db.DBContext)

	// Lookup user information.
	uid, programsRequested := c.QueryParam("uid"), c.QueryParam("programs")
	if uid == "" {
		return c.String(http.StatusBadRequest, "`uid` is a required query parameter.")
	}
	user, err := c.LoadUser(c.Request().Context(), uid)
	if err != nil {
		c.Logger().Debugf("Failed to load user with uid `%s`: %v", uid, err)
		return c.String(http.StatusNotFound, "Failed to load user.")
	}

	if db.EnableBetaFeatures == "true" {
		user.DeveloperAcc = true
	}

	resp.UserData = user

	// Get programs, if requested.
	if programsRequested != "" {
		for _, p := range resp.UserData.Programs {
			// If error in retrieving a given program, ignore it.
			currentProg, err := c.LoadProgram(c.Request().Context(), p)
			if err != nil {
				c.Logger().Warnf("Failed to load program with pid `%s` for user with uid `%s`. User could be corrupted!", p, uid)
				continue
			}

			resp.Programs[p] = currentProg
		}
	}
	return c.JSON(http.StatusOK, &resp)
}

// DeleteUser deletes an user along with all their programs
// from the database.
//
// Request Body:
//
//	{
//	    "uid": REQUIRED
//	}
//
// Returns: status 200 on deletion.
func DeleteUser(cc echo.Context) error {
	resp := struct {
		UserData db.User               `json:"userData"`
		Programs map[string]db.Program `json:"programs"`
	}{
		UserData: db.User{},
		Programs: make(map[string]db.Program),
	}

	c := cc.(*db.DBContext)

	// Lookup user information.
	uid := c.QueryParam("uid")

	if uid == "" {
		return c.String(http.StatusBadRequest, "`uid` is a required query parameter.")
	}

	user, err := c.LoadUser(c.Request().Context(), uid)

	if err != nil {
		return c.String(http.StatusNotFound, "could not find user")
	}

	resp.UserData = user

	// Delete all programs
	for _, prog := range user.Programs {
		if err := c.RemoveProgram(c.Request().Context(), prog); err != nil && status.Code(err) != codes.NotFound {
			return c.String(http.StatusInternalServerError, "failed to delete user.")
		}
	}

	if err := c.DeleteUser(c.Request().Context(), uid); err != nil {
		return c.String(http.StatusInternalServerError, "failed to delete user.")
	}

	return c.String(http.StatusOK, "user deleted successfully")
}

// CreateUser creates a new user object corresponding to either
// the provided UID or a random new one if none is provided
// with the default data.
//
// Request Body:
//
//	{
//	    "uid": string <optional>
//	}
//
// Returns: Status 200 with a marshalled User struct on success.
func CreateUser(cc echo.Context) error {
	var body struct {
		UID string `json:"uid"`
	}

	if err := httpext.RequestBodyTo(cc.Request(), &body); err != nil {
		return cc.String(http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body").Error())
	}

	// create structures to be used as default data
	newUser, newProgs := db.DefaultData()
	newUser.UID = body.UID

	c := cc.(*db.DBContext)
	user, err := c.CreateUser(c.Request().Context(), newUser)
	if err != nil {
		if strings.Contains(err.Error(), "user document with uid") {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	for _, prog := range newProgs {
		// create program in database
		p, err := c.CreateProgram(c.Request().Context(), prog)
		if err != nil {
			return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to create user").Error())
		}

		// establish association in user doc.
		user.Programs = append(user.Programs, p.UID)
	}

	// set most recent program
	user.MostRecentProgram = user.Programs[0]
	if err := c.StoreUser(c.Request().Context(), user); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to create user").Error())
	}

	return c.JSON(http.StatusCreated, &user)
}

// UpdateUser updates the doc with specified UID's fields
// to match those of the request body.
//
// Request Body:
//
//	{
//		   "uid": [REQUIRED]
//	    [User object fields]
//	}
//
// Returns: Status 200 on success.
func UpdateUser(cc echo.Context) error {
	// unmarshal request body into an User struct.
	requestObj := User{}
	c := cc.(*db.DBContext)
	if err := httpext.RequestBodyTo(c.Request(), &requestObj); err != nil {
		return err
	}

	uid := requestObj.UID

	if uid == "" {
		return c.String(http.StatusBadRequest, "a uid is required")
	}
	if len(requestObj.Programs) != 0 {
		return c.String(http.StatusBadRequest, "program list cannot be updated via /program/update")
	}

	ref := c.QueryParam("uid")
	if err := c.UpdateUser(c.Request().Context(), ref, requestObj.ToFirestoreUpdate()); err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "could not find user")
		}
		return c.String(http.StatusInternalServerError, "failed to update user.")
	}

	return c.String(http.StatusOK, "user updated successfully")
}
