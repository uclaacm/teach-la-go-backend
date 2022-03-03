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

// GetProgram retrieves information about a single program.
//
// Query parameters: pid
//
// Returns status 200 OK with a marshalled Program struct.
func (d *DB) GetProgram(c echo.Context) error {
	pid := c.QueryParam("pid")
	if pid == "" {
		return c.String(http.StatusBadRequest, "pid is required")
	}

	// attempt to acquire doc.
	p := Program{}
	ref := d.Collection(programsPath).Doc(pid)
	doc, err := ref.Get(c.Request().Context())
	if err != nil {
		return c.String(http.StatusNotFound, errors.Wrap(err, "failed to locate program").Error())
	}
	if err := doc.DataTo(&p); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to marshal data").Error())
	}

	// update UID field and respond.
	p.UID = ref.ID
	return c.JSON(http.StatusOK, &p)
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

// DeleteProgram deletes a program entry from a user.
//
// Request Body:
// {
//    uid: string
//    pid: string
// }
//
// Returns status 200 OK on deletion.
func (d *DB) DeleteProgram(c echo.Context) error {
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

	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		// remove program from user list
		uref := d.Collection(usersPath).Doc(req.UID)
		uSnap, err := tx.Get(uref)
		if err != nil {
			return err
		}
		userDoc := User{}
		if err := uSnap.DataTo(&userDoc); err != nil {
			return err
		}

		// get pid to delete then remove the entry
		idx := 0
		for i, p := range userDoc.Programs {
			if p == req.PID {
				idx = i
				break
			}
		}
		if idx >= len(userDoc.Programs) {
			return errors.New("invalid PID")
		}
		toDelete := userDoc.Programs[idx]
		userDoc.Programs = append(userDoc.Programs[:idx], userDoc.Programs[idx+1:]...)

		pref := d.Collection(programsPath).Doc(toDelete)

		// remove program from class if is in class
		pSnap, err := tx.Get(pref)
		if err != nil {
			return err
		}
		programDoc := Program{}
		if err := pSnap.DataTo(&programDoc); err != nil {
			return err
		}
		if programDoc.WID != "" {
			cid, err := d.GetUIDFromWID(c.Request().Context(), programDoc.WID, ClassesAliasPath)
			if err != nil {
				return err
			}
			classRef := d.Collection(classesPath).Doc(cid)
			if err := tx.Update(classRef, []firestore.Update{
				{Path: "programs", Value: firestore.ArrayRemove(toDelete)},
			}); err != nil {
				return err
			}
		}

		// attempt to delete program doc
		if err := tx.Set(uref, &userDoc); err != nil {
			return err
		}
		return tx.Delete(pref)
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "user or program does not exist")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to commit transaction to database").Error())
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
