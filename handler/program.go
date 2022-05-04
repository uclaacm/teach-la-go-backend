// handler shouldn't know what firebase is
// anything that says collection or doc is a reference to firebase
// 		instead call a function to call it out of firebase
// 		replace the majority of it with load program

// move out of the logic related to firebase

package handler

import (
	"context"
	"net/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/pkg/errors"
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

	wid := requestBody.WID
	var cid string
	var class db.Class
	var err error

	if wid != "" {
		cid, err = c.GetUIDFromWID(c.Request().Context(), wid, db.ClassesAliasPath)
		if err != nil {
			return err
		}

		class, err = c.LoadClass(c.Request().Context(), cid)
		if err != nil {
			return err
		}
	}

	// create program
	pRef, _ := c.CreateProgram(c.Request().Context(), requestBody.Prog)

	// associate to user, if they exist
	u, _ := c.LoadUser(c.Request().Context(), requestBody.UID)

	u.Programs = append(u.Programs, pRef.UID)
	if wid != "" {
		classRef, _ := c.LoadClass(c.Request().Context(), cid)
		classRef.Programs = append(classRef.Programs, pRef.UID)

		p.WID = class.WID

		c.StoreClass(c.Request().Context(), classRef)
	}

	p.UID = pRef.UID

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

	// err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		// remove program from user list
		// uref := d.Collection(usersPath).Doc(req.UID)
		// uSnap, err := tx.Get(uref)
		// if err != nil {
		// 	return err
		// }
		// userDoc := User{}
		// if err := uSnap.DataTo(&userDoc); err != nil {
		// 	return err
		// }
		u,err := c.LoadUser(c.Request().Context(), req.UID)

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

		// pref := d.Collection(programsPath).Doc(toDelete)
		c.StoreUser(c.Request().Context(), u)

		// remove program from class if is in class
		p, err :=c.LoadProgram(c.Request().Context(), toDelete)
		//pSnap, err := tx.Get(pref)
		// if err != nil {
		// 	return err
		// }
		// programDoc := Program{}
		// if err := pSnap.DataTo(&programDoc); err != nil {
		// 	return err
		// }
		if p.WID != "" {
			cid, err := d.GetUIDFromWID(c.Request().Context(), programDoc.WID, classesAliasPath)
			if err != nil {
				return err
			}
			c, err := c.LoadClass(c.Request().Context(), cid)
			c.RemoveProgram(context.Context, toDelete)
			classRef := c.Collection(classesPath).Doc(cid)
			if err := tx.Update(classRef, []firestore.Update{
				{Path: "programs", Value: firestore.ArrayRemove(toDelete)},
			}); err != nil {
				return err
			}
		}

		// attempt to delete program doc
		// if err := tx.Set(uref, &userDoc); err != nil {
		// 	return err
		// }
		return c.RemoveProgram(c.Request().Context(), req.PID)
	// })
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "user or program does not exist")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to commit transaction to database").Error())
	}

	return c.String(http.StatusOK, "")
}
