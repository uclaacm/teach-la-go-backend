package db

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserToFirestoreUpdate(t *testing.T) {
	t.Run("MostRecentProgram", func(t *testing.T) {
		u := User{MostRecentProgram: "someHash"}
		update := u.ToFirestoreUpdate()
		assert.Len(t, update, 1)
		assert.Equal(t, "someHash", update[0].Value)
	})
	t.Run("DisplayName", func(t *testing.T) {
		u := User{DisplayName: "test"}
		update := u.ToFirestoreUpdate()
		assert.Len(t, update, 2)
		assert.Equal(t, "mostRecentProgram", update[0].Path)
		assert.Equal(t, "test", update[1].Value)
	})
	t.Run("PhotoName", func(t *testing.T) {
		u := User{PhotoName: "icecream"}
		update := u.ToFirestoreUpdate()
		assert.Len(t, update, 2)
		assert.Equal(t, "icecream", update[1].Value)
	})
	t.Run("Programs", func(t *testing.T) {
		u := User{Programs: []string{"hash0", "hash1"}}
		update := u.ToFirestoreUpdate()
		assert.Len(t, update, 2)
		assert.Equal(t, "programs", update[1].Path)
		// TODO: value cannot be easily verified.
	})
}

func TestGetUser(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	if !assert.NoError(t, err) {
		return
	}

	t.Run("MissingUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetUser(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?uid=fakeUID", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.GetUser(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	dbConsistencyWarning(t)

	t.Run("TypicalRequests", func(t *testing.T) {
		// get some random doc
		iter := d.Collection(usersPath).Documents(context.Background())
		defer iter.Stop()
		randomDoc, err := iter.Next()
		assert.NoError(t, err)
		t.Logf("using doc ID (%s)", randomDoc.Ref.ID)

		t.Run("NoPrograms", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?uid="+randomDoc.Ref.ID, nil)
			rec := httptest.NewRecorder()
			assert.NotNil(t, req, rec)
			c := echo.New().NewContext(req, rec)

			if assert.NoError(t, d.GetUser(c)) {
				assert.NotEmpty(t, rec.Result().Body)
				assert.Equal(t, http.StatusOK, rec.Code)
				// TODO: check for the rest of the response's validity
			}
		})
		t.Run("WithPrograms", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?uid="+randomDoc.Ref.ID+"&programs=true", nil)
			rec := httptest.NewRecorder()
			assert.NotNil(t, req, rec)
			c := echo.New().NewContext(req, rec)

			if assert.NoError(t, d.GetUser(c)) {
				assert.NotEmpty(t, rec.Result().Body)
				assert.Equal(t, http.StatusOK, rec.Code)
				// TODO: check for the rest of the response's validity
			}
		})
	})
}

func TestUpdateUser(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	if !assert.NoError(t, err) {
		return
	}

	t.Run("MissingUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateUser(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{\"uid\":\"fakeUID\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.UpdateUser(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	dbConsistencyWarning(t)

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

			if assert.NoError(t, d.UpdateUser(c)) {
				assert.Equal(t, http.StatusOK, rec.Code)
				// TODO: more tests required.
			}

		})
		t.Run("MostRecentProgram", func(t *testing.T) {}) // TODO
		t.Run("PhotoName", func(t *testing.T) {})
		t.Run("Programs", func(t *testing.T) {})
	})
}

func TestCreateUser(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	if !assert.NoError(t, err) {
		return
	}

	req := httptest.NewRequest(http.MethodPut, "/", nil)
	rec := httptest.NewRecorder()
	assert.NotNil(t, req, rec)
	c := echo.New().NewContext(req, rec)

	if assert.NoError(t, d.CreateUser(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.NotEmpty(t, rec.Result().Body)

		// try to marshall result into an user struct
		u := User{}
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &u)) {
			assert.NotZero(t, u)
		}
	}
}

func TestDeleteUser(t *testing.T) {
	d, err := Open(context.Background(), os.Getenv("TLACFG"))
	if !assert.NoError(t, err) {
		return
	}

	t.Run("MissingUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("BadUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("{\"uid\":\"fakeUID\"}"))
		rec := httptest.NewRecorder()
		assert.NotNil(t, req, rec)
		assert.NotNil(t, req.Body)
		c := echo.New().NewContext(req, rec)

		if assert.NoError(t, d.DeleteUser(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
		}
	})

	// TODO: live tests
}
