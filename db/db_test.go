package db

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// quick warning message for when we start
// doing "hot integration" tests.
func dbConsistencyWarning(t *testing.T) {
	t.Log("-------")
	t.Log("WARNING")
	t.Log("-------")
	t.Log("The following tests assume the database is consistent.")
	t.Log("Random docs are pulled for testing. If those docs are not valid, tests *will* fail.")
}

func TestOpen(t *testing.T) {
	t.Run("NoConfig", func(t *testing.T) {
		_, err := Open(context.Background(), "")
		assert.Error(t, err)
	})
	t.Run("InvalidJSON", func(t *testing.T) {
		_, err := Open(context.Background(), "{}")
		assert.Error(t, err)
	})
	t.Run("ValidJSON", func(t *testing.T) {
		_, err := Open(context.Background(), os.Getenv("TLACFG"))
		assert.NoError(t, err)
	})
}
