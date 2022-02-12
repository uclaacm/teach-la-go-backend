// handler shouldn't know what firebase is
// anything that says collection or doc is a reference to firebase
// 		instead call a function to call it out of firebase
// 		replace the majority of it with load program

// move out of the logic related to firebase

package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uclaacm/teach-la-go-backend/db"
)

// GetProgram retrieves information about a single program.
//
// Query parameters: pid
//
// Returns status 200 OK with a marshalled Program struct.
func GetProgram(cc echo.Context) error {
	c := cc.(*db.DBContext)
	pid := c.QueryParam("pid")
	p, err := c.LoadProgram(c.Request().Context(), pid)
	if err != nil {
		c.Logger().Debugf("Failed to load program with pid `%s`: %v", pid, err)
		return c.String(http.StatusNotFound, "Failed to load program.")
	}

	return c.JSON(http.StatusOK, &p)
}