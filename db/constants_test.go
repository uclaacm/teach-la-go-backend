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
	p := defaultProgram(langString(python))
	assert.NotEmpty(t, p)
	p = defaultProgram(langString(processing))
	assert.NotEmpty(t, p)
	p = defaultProgram(langString(html))
	assert.NotEmpty(t, p)
	p = defaultProgram("not a language")
	assert.Empty(t, p)
}

func TestDefaultData(t *testing.T) {
	u, p := defaultData()
	assert.NotEmpty(t, p)
	assert.NotEmpty(t, u)
}
