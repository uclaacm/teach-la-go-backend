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
	t.Run("ValidBody", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("{\"f\":42}"))
		var s struct {
			F int `json:"f"`
		}
		assert.NoError(t, httpext.RequestBodyTo(r, &s))
	})
}
