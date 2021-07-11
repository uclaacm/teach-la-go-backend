package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/httpext"
)

// GetClass takes the UID (either of a member or an instructor)
// and a CID (wid) as a JSON, and returns an object representing the class.
// If the given UID is not a member or an instructor, an error is returned.
func GetClass(cc echo.Context) error {
	var (
		req struct {
			UID string `json:"uid"`
			CID string `json:"cid"`
		}
		res struct {
			*db.Class
			ProgramData []db.Program       `json:"programData"`
			UserData    map[string]db.User `json:"userData"`
		}
		err error
	)
	res.UserData = make(map[string]db.User)

	c := cc.(*db.DBContext)

	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.UID == "" || req.CID == "" {
		return c.String(http.StatusBadRequest, "uid and cid fields are both required")
	}

	class, err := c.LoadClass(c.Request().Context(), req.CID)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	res.Class = &class

	// Check if the uid exists in the members list or instructor list
	isIn, isInstructor := false, false
	for _, m := range class.Members {
		if m == req.UID {
			isIn = true
			break
		}
	}
	for _, i := range class.Instructors {
		if i == req.UID {
			isIn, isInstructor = true, true
			break
		}
	}

	if !isIn {
		return c.String(http.StatusBadRequest, "given user not in class")
	}

	// If program data is requested.
	partial := false
	if withPrograms := c.QueryParam("programs"); withPrograms != "" && withPrograms != "false" {
		for _, p := range class.Programs {
			program, err := c.LoadProgram(c.Request().Context(), p)
			if err != nil {
				partial = true
			}
			res.ProgramData = append(res.ProgramData, program)
		}
	}

	if withUserData := c.QueryParam("userData"); isInstructor && withUserData != "" && withUserData != "false" {
		for _, uid := range class.Members {
			user, err := c.LoadUser(c.Request().Context(), uid)
			if err != nil {
				partial = true
			}
			res.UserData[user.UID] = user
		}
	}

	if partial {
		return c.JSON(http.StatusPartialContent, res)
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
