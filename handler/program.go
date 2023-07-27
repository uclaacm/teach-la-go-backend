// handler shouldn't know what firebase is
// anything that says collection or doc is a reference to firebase
// 		instead call a function to call it out of firebase
// 		replace the majority of it with load program

// move out of the logic related to firebase

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

func CreateProgram(cc echo.Context) error {
	var requestBody struct {
		UID  string     `json:"uid"`
		WID  string     `json:"wid"`
		Prog db.Program `json:"program"`
	}

	c := cc.(*db.DBContext)

	if err := httpext.RequestBodyTo(c.Request(), &requestBody); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}

	// check that language exists.
	p := db.DefaultProgram(requestBody.Prog.Language)
	if p.Code == "" {
		return c.String(http.StatusBadRequest, "language does not exist")
	}

	// thumbnail should be within range.
	if requestBody.Prog.Thumbnail > db.ThumbnailCount || requestBody.Prog.Thumbnail < 0 {
		return c.String(http.StatusBadRequest, "thumbnail index out of bounds")
	}
	p.Thumbnail = requestBody.Prog.Thumbnail

	// add code if provided.
	if requestBody.Prog.Code != "" {
		p.Code = requestBody.Prog.Code
	}

	// add name if provided.
	if requestBody.Prog.Name != "" {
		p.Name = requestBody.Prog.Name
	}

	// use the composite function to guarantee consistency
	err := c.CreateProgramAndAssociate(c.Request().Context(), p, requestBody.UID, requestBody.WID)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, errors.Wrap(err, "failed to find user document").Error())
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to create program and associate to user or class").Error())
	}

	return c.JSON(http.StatusCreated, &p)

}

func DeleteProgram(cc echo.Context) error {
	c := cc.(*db.DBContext)
	// acquire parameters via anonymous struct.
	var req struct {
		UID string `json:"uid"`
		PID string `json:"pid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.UID == "" || req.PID == "" {
		return c.String(http.StatusBadRequest, "uid and idx fields are both required")
	}

	u, err := c.LoadUser(c.Request().Context(), req.UID)
	if err != nil {
		return err
	}

	// get pid to delete then remove the entry
	idx := 0
	for i, p := range u.Programs {
		if p == req.PID {
			idx = i
			break
		}
	}
	if idx >= len(u.Programs) {
		return errors.New("invalid PID")
	}
	toDelete := u.Programs[idx]
	u.Programs = append(u.Programs[:idx], u.Programs[idx+1:]...)

	err = c.StoreUser(c.Request().Context(), u)
	if err != nil {
		return err
	}

	// remove program from class if is in class
	p, err := c.LoadProgram(c.Request().Context(), toDelete)

	if err != nil {
		return err
	}

	if p.WID != "" {
		cid, err := c.GetUIDFromWID(c.Request().Context(), p.WID, db.ClassesAliasPath)
		if err != nil {
			return err
		}
		cls, err := c.LoadClass(c.Request().Context(), cid)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return c.String(http.StatusNotFound, "class or program does not exist")
			}
			return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to commit transaction to database").Error())
		}
		idx := 0
		for i, p := range cls.Programs {
			if p == req.PID {
				idx = i
				break
			}
		}
		if idx >= len(cls.Programs) {
			return errors.New("invalid PID")
		}
		cls.Programs = append(cls.Programs[:idx], cls.Programs[idx+1:]...)

		err = c.StoreClass(c.Request().Context(), cls)
		if err != nil {
			return err
		}
	}

	err = c.RemoveProgram(c.Request().Context(), req.PID)

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "user or program does not exist")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to commit transaction to database").Error())
	}

	return c.String(http.StatusOK, "")
}
