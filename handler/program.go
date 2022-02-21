package handler

import (
	"context"
	"net/http"
	
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProgram takes partial program fields and a user
// ID for the owner, then creates it.
//
// Request Body:
// {
//    uid: UID for the user the program belongs to
//	  wid: [optional WID for the class the program should be added to]
//    program: {
//        thumbnail: index of the desired thumbnail
//        language: language string
//        name: name of the program
//        code: [optional code for the program]
//    }
// }
//
// Returns 201 created on success. TODO: postman docs
func (d *DB) CreateProgramTemp(c echo.Context) error {
	var requestBody struct {
		UID  string  `json:"uid"`
		WID  string  `json:"wid"`
		Prog Program `json:"program"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &requestBody); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}

	// check that language exists.
	p := defaultProgram(requestBody.Prog.Language)
	if p.Code == "" {
		return c.String(http.StatusBadRequest, "language does not exist")
	}

	// thumbnail should be within range.
	if requestBody.Prog.Thumbnail > thumbnailCount || requestBody.Prog.Thumbnail < 0 {
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
	var class *Class
	if wid != "" {
		var err error
		cid, err = d.GetUIDFromWID(c.Request().Context(), wid, classesAliasPath)
		if err != nil {
			return err
		}

		class, err = d.loadClass(c.Request().Context(), cid)
		if err != nil {
			return err
		}
	}

	// create the program doc.
	err := d.RunTransaction(c.Request().Context()) error {
		// create program
		pRef := d.CreateProgram(c.Request().Context(), Prog)

		// associate to user, if they exist
		u := d.LoadUser(c.Request().Context(), requestBody.UID)

		//***may need to write a new function for this code!!
		u.Programs = append(u.Programs, pRef.ID)
		if wid != "" {
			classRef := d.loadClass(c.Request().Context(), cid)
			p.programs = classRef;

			p.WID = class.WID
		}
		err := d.StoreClass(c.Request().Context(), u) error;

		p.UID = pRef.ID
		return pRef;
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, errors.Wrap(err, "failed to find user document").Error())
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to create program and associate to user or class").Error())
	}

	return c.JSON(http.StatusCreated, p)
}