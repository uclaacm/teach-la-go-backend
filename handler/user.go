package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uclaacm/teach-la-go-backend/db"
)

// GetUser acquires the userdoc with the given uid.
//
// Query Parameters:
//  - uid string: UID of user to get
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

	c, ok := cc.(*db.DBContext)
	ec := c.Context
	if !ok {
		return c.String(http.StatusInternalServerError, "Failed to acquire database connection!")
	}

	// Lookup user information.
	user, err := c.LoadUser(ec.Request().Context(), c.QueryParam("uid"))
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load user.")
	}
	resp.UserData = user

	// get programs, if requested
	if c.QueryParam("programs") != "" {
		for _, p := range resp.UserData.Programs {
			// If error in retrieving program, ignore it.
			currentProg, err := c.LoadProgram(ec.Request().Context(), p)
			if err != nil {
				continue
			}

			resp.Programs[p] = currentProg
		}
	}
	return c.JSON(http.StatusOK, &resp)
}
