package db

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Program is a representation of a program document.
type Program struct {
	Code        string `firestore:"code" json:"code"`
	DateCreated string `firestore:"dateCreated" json:"dateCreated"`
	Language    string `firestore:"language" json:"language"`
	Name        string `firestore:"name" json:"name"`
	Thumbnail   int64  `firestore:"thumbnail" json:"thumbnail"`
	UID         string `json:"uid"`
	WID         string `json:"wid"` // Optional WID of class associated with program
}

// ToFirestoreUpdate returns the []firestore.Update representation
// of this struct. Any fields that are non-zero valued are included
// in the update, save for the date of creation.
func (p *Program) ToFirestoreUpdate() (up []firestore.Update) {
	if p.Code != "" {
		up = append(up, firestore.Update{Path: "code", Value: p.Code})
	}
	if p.Language != "" {
		up = append(up, firestore.Update{Path: "language", Value: p.Language})
	}
	if p.Name != "" {
		up = append(up, firestore.Update{Path: "name", Value: p.Name})
	}
	if p.Thumbnail != 0 {
		up = append(up, firestore.Update{Path: "thumbnail", Value: p.Thumbnail})
	}

	return
}

// UpdateProgram expects an array of partial Program structs
// and a UID of the user they belong to. If the user pointed
// to by UID does not own the programs passed to update,
// no programs are updated.
//
// Request Body:
// {
//     "uid": [REQUIRED],
//     "programs": [array of partial program objects as indexed in user]
// }
//
// Returns status 200 OK on nominal request.
func (d *DB) UpdateProgram(c echo.Context) error {
	var body struct {
		UID      string             `json:"uid"`
		Programs map[string]Program `json:"programs"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, "failed to read request body")
	}
	if body.UID == "" {
		return c.String(http.StatusBadRequest, "a uid is required")
	}

	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		usnap, err := tx.Get(d.Collection(usersPath).Doc(body.UID))
		if err != nil {
			return err
		}
		owner := User{}
		if err := usnap.DataTo(&owner); err != nil {
			return err
		}

		for id, p := range body.Programs {
			// confirm that the program specified is owned by UID.
			belongsTo := false
			for _, userProg := range owner.Programs {
				if id == userProg {
					belongsTo = true
					break
				}
			}
			if !belongsTo {
				return errors.Errorf("specified program is out of bounds for user %s", body.UID)
			}

			// update the program
			pref := d.Collection(programsPath).Doc(id)
			if err := tx.Update(pref, p.ToFirestoreUpdate()); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, errors.Wrap(err, "program ID could not be found").Error())
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to write update(s) to database").Error())
	}

	return c.String(http.StatusOK, "")
}

// ForkProgram forks a program `pid` to the user `uid`.
//
// Request Body:
// {
//    uid string
//    pid string
// }
//
// Returns status 201 created on success.
func (d *DB) ForkProgram(c echo.Context) error {
	// validate request structure
	var body struct {
		UID string `json:"uid"`
		PID string `json:"pid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &body); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if body.UID == "" || body.PID == "" {
		return c.String(http.StatusBadRequest, "uid and pid are both required")
	}

	forkedProgram := Program{}
	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		uref := d.Collection(usersPath).Doc(body.UID)
		pref := d.Collection(programsPath).Doc(body.PID)
		pSnap, err := tx.Get(pref)
		if err != nil {
			return err
		}

		// copy program
		newProgram := d.Collection(programsPath).NewDoc()
		if err := tx.Create(newProgram, pSnap); err != nil {
			return err
		}

		// TODO: strong potential for code lifting here.
		userSnap, err := tx.Get(uref)
		u := User{}
		if err != nil {
			return err
		}
		if err := userSnap.DataTo(&u); err != nil {
			return err
		}

		u.Programs = append(u.Programs, newProgram.ID)
		if err := tx.Set(uref, u.ToFirestoreUpdate()); err != nil {
			return err
		}
		return pSnap.DataTo(&forkedProgram)
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "could not find the program or user")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to fork program").Error())
	}

	return c.JSON(http.StatusCreated, forkedProgram)
}
