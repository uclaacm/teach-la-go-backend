package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/handler"
)

func TestGetUser(t *testing.T) {
	t.Run("MissingUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodGet, "/?uid=doesnotexist", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	t.Run("ValidUID", func(t *testing.T) {
		d := db.OpenMock()

		// Insert an example user for testing.
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodGet, "/?uid=test", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
	t.Run("WithPrograms", func(t *testing.T) {
		d := db.OpenMock()

		// Insert an example user and program.
		prog := db.Program{
			UID: "testprog",
		}
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID:      "testuser",
			Programs: []string{prog.UID},
		}))
		require.NoError(t, d.StoreProgram(context.Background(), prog))

		req := httptest.NewRequest(http.MethodGet, "/?uid=testuser&programs=true", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		resp := struct {
			UserData db.User               `json:"userData"`
			Programs map[string]db.Program `json:"programs"`
		}{
			UserData: db.User{},
			Programs: make(map[string]db.Program),
		}

		if assert.NoError(t, handler.GetUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			require.Equal(t, http.StatusOK, rec.Code)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.NotEmpty(t, resp.Programs)
			prog, ok := resp.Programs["testprog"]
			assert.True(t, ok)
			assert.Equal(t, "testprog", prog.UID)
		}
	})
}
