package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Parameters for additional data.
	withPrograms, withUserData := c.QueryParam("programs"), c.QueryParam("userData")

	if !isIn {
		return c.String(http.StatusBadRequest, "given user not in class")
	}

	// If program data is requested.
	partial := false
	if withPrograms != "" && withPrograms != "false" {
		for _, p := range class.Programs {
			program, err := c.LoadProgram(c.Request().Context(), p)
			if err != nil {
				partial = true
			}
			res.ProgramData = append(res.ProgramData, program)
		}
	}

	// Retrieve userData if requested.
	if withUserData != "" && withUserData != "false" {
		if isInstructor {
			for _, uid := range class.Members {
				user, err := c.LoadUser(c.Request().Context(), uid)
				if err != nil {
					partial = true
				}
				res.UserData[user.UID] = user
			}
		}

		// Students should see Instructor data
		for _, uid := range class.Instructors {
			user, err := c.LoadUser(c.Request().Context(), uid)
			if err != nil {
				partial = true
			}
			res.UserData[user.UID] = user
		}
	}

	// Indicate whether the response is partial.
	if partial {
		return c.JSON(http.StatusPartialContent, res)
	} else {
		return c.JSON(http.StatusOK, res)
	}
}

// DeleteClass takes a wid and deletes it.
// Any programs associated with the class will also be deleted.
// Users that are in the class will still contain a reference to this class,
// thus it is the user's responsibility to remove references to a deleted class.
func DeleteClass(cc echo.Context) error {
	var req struct {
		CID string `json:"cid"`
	}
	
	c := cc.(*db.DBContext)
	
	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.CID == "" {
		return c.String(http.StatusBadRequest, "cid is required")
	}

	class, err := c.LoadClass(c.Request().Context(), req.CID)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	for _, prog := range class.Programs {
		if err := c.RemoveProgram(c.Request().Context(), prog); 
		// if we can't find a program, then it's not a problem.
		err != nil && status.Code(err) != codes.NotFound {
			return err
		}
	}

	/* 	TODO: since this is no longer a transaction, potential for Class to not be
		deleted, but all the programs were still deleted 	*/
	if err := c.DeleteClass(c.Request().Context(), class); err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "could not find class")
		}

		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to delete class").Error())
	}

	return c.String(http.StatusOK, "")
}