package db

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uclaacm/teach-la-go-backend/httpext"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
)

// User is a struct representation of a user document.
// It provides functions for converting the struct
// to firebase-digestible types.
type User struct {
	Classes           []string `firestore:"classes" json:"classes"`
	DisplayName       string   `firestore:"displayName" json:"displayName"`
	MostRecentProgram string   `firestore:"mostRecentProgram" json:"mostRecentProgram"`
	PhotoName         string   `firestore:"photoName" json:"photoName"`
	Programs          []string `firestore:"programs" json:"programs"`
	UID               string   `json:"uid"`
	DeveloperAcc      bool     `firestore:"developerAcc" json:"developerAcc"`
}

// ToFirestoreUpdate returns the database update
// representation of its UserData struct.
func (u *User) ToFirestoreUpdate() []firestore.Update {
	f := []firestore.Update{
		{Path: "mostRecentProgram", Value: u.MostRecentProgram},
	}

	switch {
	case u.DisplayName != "":
		f = append(f, firestore.Update{Path: "displayName", Value: u.DisplayName})
	case u.PhotoName != "":
		f = append(f, firestore.Update{Path: "photoName", Value: u.PhotoName})
	case len(u.Programs) != 0:
		f = append(f, firestore.Update{Path: "programs", Value: firestore.ArrayUnion(u.Programs)})
	}

	return f
}

// UpdateUser updates the doc with specified UID's fields
// to match those of the request body.
//
// Request Body:
// {
//	   "uid": [REQUIRED]
//     [User object fields]
// }
//
// Returns: Status 200 on success.
func (d *DB) UpdateUser(c echo.Context) error {
	// unmarshal request body into an User struct.
	requestObj := User{}
	if err := httpext.RequestBodyTo(c.Request(), &requestObj); err != nil {
		return err
	}

	uid := requestObj.UID
	if uid == "" {
		return c.String(http.StatusBadRequest, "a uid is required")
	}
	if len(requestObj.Programs) != 0 {
		return c.String(http.StatusBadRequest, "program list cannot be updated via /program/update")
	}

	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		ref := d.Collection(usersPath).Doc(uid)
		return tx.Update(ref, requestObj.ToFirestoreUpdate())
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "user could not be found")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to update user data").Error())
	}

	return c.String(http.StatusOK, "user updated successfully")
}

// DeleteUser deletes an user along with all their programs
// from the database.
//
// Request Body:
// {
//     "uid": REQUIRED
// }
//
// Returns: status 200 on deletion.
func (d *DB) DeleteUser(c echo.Context) error {
	var req struct {
		UID string `json:"uid"`
	}
	if err := httpext.RequestBodyTo(c.Request(), &req); err != nil {
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to read request body").Error())
	}
	if req.UID == "" {
		return c.String(http.StatusBadRequest, "uid is required")
	}
	userRef := d.Collection(usersPath).Doc(req.UID)

	err := d.RunTransaction(c.Request().Context(), func(ctx context.Context, tx *firestore.Transaction) error {
		userSnap, err := tx.Get(userRef)
		if err != nil {
			return err
		}

		usr := User{}
		if err := userSnap.DataTo(&usr); err != nil {
			return err
		}

		for _, prog := range usr.Programs {
			progRef := d.Collection(programsPath).Doc(prog)
			// if we can't find a program, then it's not a problem.
			if err := tx.Delete(progRef); err != nil && status.Code(err) != codes.NotFound {
				return err
			}
		}

		return tx.Delete(userRef)
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return c.String(http.StatusNotFound, "could not find user")
		}
		return c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to delete user").Error())
	}
	return c.String(http.StatusOK, "user deleted successfully")
}
