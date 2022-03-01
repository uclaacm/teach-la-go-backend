package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLangString(t *testing.T) {
	assert.Equal(t, langString(python), "python")
	assert.Equal(t, langString(processing), "processing")
	assert.Equal(t, langString(html), "html")
	assert.Equal(t, langString(langCount), "DNE")
}

func TestDefaultProgram(t *testing.T) {
	p := DefaultProgram(langString(python))
	assert.NotEmpty(t, p)
	p = DefaultProgram(langString(processing))
	assert.NotEmpty(t, p)
	p = DefaultProgram(langString(html))
	assert.NotEmpty(t, p)
	p = DefaultProgram("not a language")
	assert.Empty(t, p)
}

func TestDefaultData(t *testing.T) {
	u, p := DefaultData()
	assert.NotEmpty(t, p)
	assert.NotEmpty(t, u)
}
