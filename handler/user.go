package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/httpext"
)

// GetUser acquires the user document with the given uid. The
// provided context must be a *db.DBContext.
//
// Query Parameters:
//  - uid string: UID of user to GET
//	- programs string: Whether to acquire programs.
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

// CreateUser creates a new user object corresponding to either
// the provided UID or a random new one if none is provided
// with the default data.
//
// Request Body:
// {
//     "uid": string <optional>
// }
//
// Returns: Status 200 with a marshalled User struct on success.
func CreateUser(cc echo.Context) error {
	var body struct {
		UID string `json:"uid"`
	}

	c := cc.(*db.DBContext)

	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to marshal request body").Error())
	}

	newUser, err := c.NewUser(c.Request().Context(), body.UID)
	if err != nil {
		if strings.Contains(err.Error(), "user document with uid '") {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to create user").Error())
	}

	return c.JSON(http.StatusCreated, &newUser)
}
