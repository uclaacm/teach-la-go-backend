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
}
