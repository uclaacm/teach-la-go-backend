package db

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestForkProgram(t *testing.T) {
	// TODO
}
