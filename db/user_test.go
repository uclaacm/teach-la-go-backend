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
