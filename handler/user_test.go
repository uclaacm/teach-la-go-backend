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
	t.Run("MissingProgram", func(t *testing.T) {
		d := db.OpenMock()

		// Insert an example user missing a program.
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID:      "testuser",
			Programs: []string{"doesnotexist"},
		}))

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
			assert.Empty(t, resp.Programs)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("MissingUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.DeleteUser(&db.DBContext{
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

		if assert.NoError(t, handler.DeleteUser(&db.DBContext{
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

		if assert.NoError(t, handler.DeleteUser(&db.DBContext{
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

		if assert.NoError(t, handler.DeleteUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			require.Equal(t, http.StatusOK, rec.Code)
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
		}
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("emptyUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.CreateUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)

			// try to marshall result into an user struct
			u := db.User{}
			if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &u)) {
				assert.NotZero(t, u)
			}
		}
	})

	t.Run("providedUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"uid": "abcdef123"}`))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.CreateUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Result().Body)

			// try to marshall result into an user struct
			u := db.User{}
			if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &u)) {
				assert.NotZero(t, u)
				assert.Equal(t, u.UID, "abcdef123")
			}
		}
	})

	t.Run("repeatedUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"uid": "abcdef123"}`))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		assert.NoError(t, handler.CreateUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		}))

		req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"uid": "abcdef123"}`))
		rec = httptest.NewRecorder()
		c = echo.New().NewContext(req, rec)
		if assert.NoError(t, handler.CreateUser(&db.DBContext{
			Context: c,
			TLADB:   d,
		})) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

}

func TestUpdateUser(t *testing.T) {
	t.Run("MissingUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.UpdateUser(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		d := db.OpenMock()
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{\"uid\":\"fakeUID\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, handler.UpdateUser(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	// dbConsistencyWarning(t)

	t.Run("TypicalRequests", func(t *testing.T) {
		iter := d.Collection(usersPath).Documents(context.Background())
		defer iter.Stop()
		randomDoc, err := iter.Next()
		if !assert.NoError(t, err) {
			return
		}
		t.Logf("using doc ID (%s)", randomDoc.Ref.ID)

		u := User{}
		if err := randomDoc.DataTo(&u); !assert.NoError(t, err) {
			t.Fatalf("encountered a fatal error when converting random user doc to object: %s", err)
		}
		u.UID = randomDoc.Ref.ID
		u.Programs = []string{}

		t.Run("DisplayName", func(t *testing.T) {
			uCopy := u
			uCopy.DisplayName = "test"

			bytes, err := json.Marshal(&uCopy)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(bytes)))
			rec := httptest.NewRecorder()
			assert.NotNil(t, req, rec)
			c := echo.New().NewContext(req, rec)

			if assert.NoError(t, handler.UpdateUser(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				// TODO: more tests required.
			}

		})
		t.Run("MostRecentProgram", func(t *testing.T) {}) // TODO
		t.Run("PhotoName", func(t *testing.T) {})
		t.Run("Programs", func(t *testing.T) {})
	})
}
