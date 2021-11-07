package httpext_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uclaacm/teach-la-go-backend/httpext"
)

func TestRequestBodyTo(t *testing.T) {
	t.Run("NilBody", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		s := ""
		assert.NoError(t, httpext.RequestBodyTo(r, &s))
	})
	t.Run("EmptyBody", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{}"))
		var s struct {
			SomeField bool `json:"someField"`
		}
		assert.NoError(t, httpext.RequestBodyTo(r, &s))
		assert.Zero(t, s.SomeField)
	})
	t.Run("ValidBody", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"f\":42}"))
		var s struct {
			F int `json:"f"`
		}
		assert.NoError(t, httpext.RequestBodyTo(r, &s))
	})
	t.Run("PartialFill", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"f\":42}"))
		var s struct {
			F     int  `json:"f"`
			Empty bool `json:"empty"`
		}
		assert.NoError(t, httpext.RequestBodyTo(r, &s))
		assert.Equal(t, 42, s.F)
		assert.Zero(t, s.Empty)
	})
}
