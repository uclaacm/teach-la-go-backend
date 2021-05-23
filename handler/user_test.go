package handler_test

import (
	"context"
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
}
