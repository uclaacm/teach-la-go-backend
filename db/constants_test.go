package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
