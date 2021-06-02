package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uclaacm/teach-la-go-backend/db"
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
