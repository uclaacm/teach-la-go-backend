package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uclaacm/teach-la-go-backend/db"
	"github.com/uclaacm/teach-la-go-backend/handler"
)

func TestGetClass(t *testing.T) {
	t.Run("missingUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("missingCID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"uid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("improperBody", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		}
	})
	t.Run("classDNE", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"uid\": \"test\", \"cid\": \"does not exist\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})
	t.Run("validClass", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:     "test",
			Members: []string{"test"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
	t.Run("withPrograms", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:      "test",
			Members:  []string{"test"},
			Programs: []string{"test"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		require.NoError(t, d.StoreProgram(context.Background(), db.Program{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodPost, "/?programs=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
	t.Run("withUsers", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:         "test",
			Instructors: []string{"test"},
			Members:     []string{"test"},
			Programs:    []string{"test"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID:         "test",
			DisplayName: "Test Account",
		}))
		req := httptest.NewRequest(http.MethodPost, "/?userData=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusOK, rec.Code)
			require.NotEmpty(t, rec.Body)
		}
	})
	t.Run("withBoth", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:         "test",
			Instructors: []string{"test"},
			Members:     []string{"test"},
			Programs:    []string{"test"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID:         "test",
			DisplayName: "Test Account",
		}))
		require.NoError(t, d.StoreProgram(context.Background(), db.Program{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodPost, "/programs=true&userData=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
}
