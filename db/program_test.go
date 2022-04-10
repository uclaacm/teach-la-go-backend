package db

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateProgram(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	t.Run("MissingUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{\"programs\":{}}"))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateProgram(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{\"uid\":\"badUID\",\"programs\":{}}"))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateProgram(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	t.Run("EmptyRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(""))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateProgram(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("Bad JSON"))
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateProgram(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
	// TODO: more rigorous integration tests
}

func TestDeleteProgram(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	require.NoError(t, err)

	dbConsistencyWarning(t)
	t.Log("this test will assume that the first userdoc pulled from staging has at least one program")

	t.Run("TypicalRequest", func(t *testing.T) {
		randomUser := User{}
		err = d.RunTransaction(context.Background(), func(ctx context.Context, tx *firestore.Transaction) error {
			// get some random user with more than one program
			userDoc, err := d.Collection(usersPath).DocumentRefs(context.Background()).Next()
			if err != nil {
				return err
			}
			userSnap, err := tx.Get(userDoc)
			if err != nil {
				return err
			}
			return userSnap.DataTo(&randomUser)
		})
		require.NoError(t, err)

		// request struct
		request := struct {
			UID string `json:"uid"`
			PID string `json:"pid"`
		}{
			UID: randomUser.UID,
			PID: randomUser.Programs[0],
		}
		t.Logf("trying to delete program with pid (%s)", request.PID)
		programToDelete := d.Collection(programsPath).Doc(request.PID)
		b, err := json.Marshal(&request)
		require.NoError(t, err)

		// try to delete it
		req, rec := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(string(b))), httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		if assert.NoError(t, d.DeleteProgram(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// check that the program actually was deleted
			_, err := programToDelete.Get(context.Background())
			assert.Equal(t, codes.NotFound, status.Code(err))
		}
	})
}

func TestForkProgram(t *testing.T) {
	// TODO
}
