package db

import (
	"testing"

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
