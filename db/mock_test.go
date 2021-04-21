package db_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/uclaacm/teach-la-go-backend/db"
	"testing"
)

func TestMockDB_StoreClass(t *testing.T) {
	t.Run("stores", func(t *testing.T) {
		d := db.OpenMock()
		err := d.StoreClass(nil, db.Class{})
		assert.NoError(t, err)
	})
}
