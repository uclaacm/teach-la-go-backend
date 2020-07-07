package db

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
