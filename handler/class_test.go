package handler_test

import (
	"context"
	"encoding/json"
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
	t.Run("userNotInClass", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:     "test",
			Members: []string{"test"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "notInClass",
		}))
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"uid\": \"notInClass\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "given user not in class", rec.Body.String())
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
			require.Equal(t, http.StatusOK, rec.Code)
			res := struct {
				*db.Class
				ProgramData []db.Program       `json:"programData"`
				UserData    map[string]db.User `json:"userData"`
			}{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			assert.NotZero(t, res.CID)
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
			require.Equal(t, http.StatusOK, rec.Code)
			res := struct {
				*db.Class
				ProgramData []db.Program       `json:"programData"`
				UserData    map[string]db.User `json:"userData"`
			}{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			assert.NotZero(t, res.CID)
			assert.NotEmpty(t, res.ProgramData)
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
			require.Equal(t, http.StatusOK, rec.Code)
			res := struct {
				*db.Class
				ProgramData []db.Program       `json:"programData"`
				UserData    map[string]db.User `json:"userData"`
			}{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			assert.NotZero(t, res.CID)
			assert.NotEmpty(t, res.UserData)
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
		req := httptest.NewRequest(http.MethodPost, "/?programs=true&userData=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			require.Equal(t, http.StatusOK, rec.Code)
			res := struct {
				*db.Class
				ProgramData []db.Program       `json:"programData"`
				UserData    map[string]db.User `json:"userData"`
			}{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
			assert.NotZero(t, res.CID)
			assert.NotEmpty(t, res.ProgramData)
			assert.NotEmpty(t, res.UserData)
		}
	})
	t.Run("partialPrograms", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:      "test",
			Members:  []string{"test"},
			Programs: []string{"doesNotExist"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
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
			require.Equal(t, http.StatusPartialContent, rec.Code)
		}
	})
	t.Run("partialUsers", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:         "test",
			Instructors: []string{"test"},
			Members:     []string{"test", "doesNotExist"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodPost, "/?userData=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			require.Equal(t, http.StatusPartialContent, rec.Code)
		}
	})
	t.Run("userIsNotInstructor", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID:     "test",
			Members: []string{"test", "doesNotExist"},
		}))
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		req := httptest.NewRequest(http.MethodPost, "/?userData=true", strings.NewReader("{\"uid\": \"test\", \"cid\": \"test\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.GetClass(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, rec.Body.String(), "user is not an instructor")
		}
	})
}
