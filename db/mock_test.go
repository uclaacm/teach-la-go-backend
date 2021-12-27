package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uclaacm/teach-la-go-backend/db"
)

func TestMockUser(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		d := db.OpenMock()
		assert.NoError(t, d.StoreUser(context.Background(), db.User{}))
	})
	t.Run("load", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		_, err := d.LoadUser(context.Background(), "test")
		assert.NoError(t, err)
	})
	t.Run("create", func(t *testing.T) {
		d := db.OpenMock()
	})
	t.Run("invalidLoad", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		_, err := d.LoadUser(context.Background(), "invalid")
		assert.Error(t, err)
	})
	t.Run("delete", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreUser(context.Background(), db.User{
			UID: "test",
		}))
		assert.NoError(t, d.DeleteUser(context.Background(), "test"))
	})
}

func TestMockProgram(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		d := db.OpenMock()
		assert.NoError(t, d.StoreProgram(context.Background(), db.Program{}))
	})
	t.Run("load", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreProgram(context.Background(), db.Program{
			UID: "test",
		}))
		_, err := d.LoadProgram(context.Background(), "test")
		assert.NoError(t, err)
	})
	t.Run("invalidLoad", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreProgram(context.Background(), db.Program{
			UID: "test",
		}))
		_, err := d.LoadProgram(context.Background(), "invalid")
		assert.Error(t, err)
	})
	// Add tests if there is a DeleteProgram
}

func TestMockClass(t *testing.T) {
	t.Run("store", func(t *testing.T) {
		d := db.OpenMock()
		assert.NoError(t, d.StoreClass(context.Background(), db.Class{}))
	})
	t.Run("load", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID: "test",
		}))
		_, err := d.LoadClass(context.Background(), "test")
		assert.NoError(t, err)
	})
	t.Run("create", func(t *testing.T) {
		d := db.OpenMock()
		assert.NoError(t, d.StoreClass(context.Background(), db.Class{}))
	})
	t.Run("invalidLoad", func(t *testing.T) {
		d := db.OpenMock()
		require.NoError(t, d.StoreClass(context.Background(), db.Class{
			CID: "test",
		}))
		_, err := d.LoadClass(context.Background(), "invalid")
		assert.Error(t, err)
	})
	// Add tests if there is a DeleteClass
}
