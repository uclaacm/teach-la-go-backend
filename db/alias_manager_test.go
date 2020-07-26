package db_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	tinycrypt "github.com/uclaacm/teach-la-go-backend-tinycrypt"
	"github.com/uclaacm/teach-la-go-backend/db"
)

// Runs series of test to test functionality of database
func TestAliasManagement(t *testing.T) {
	// Test opening connection with database
	d, err := db.Open(context.Background(), os.Getenv("TLACFG"))
	if !assert.Nil(t, err) {
		return
	}

	// Initialize the shards
	t.Run("CreateShards", func(t *testing.T) {
		err := d.InitShards(context.Background(), "classes_alias")
		assert.Nil(t, err)
	})

	// Request unique ID numbers from the counter
	t.Run("GetID", func(t *testing.T) {
		for i := 0; i < 32; i++ {
			u, err := d.GetID(context.Background(), "classes_alias")
			assert.Nil(t, err)

			t.Logf("Unique ID + hashing:\t\t%d\n", u)
			w := tinycrypt.GenerateWord24(uint64(u))
			wid := strings.Join(w, ",")
			t.Logf("Unique ID + hashing + words:\t%s\n======", wid)
		}
	})

}
