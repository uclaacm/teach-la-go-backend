package lib_test

import (
	"context"
	"os"
	"testing"

	"../lib"
)

type dbTester struct {
	*lib.DB
}

// runs all tests for the database interface in order.
func TestDB(t *testing.T) {
	if openResult := t.Run("open", testOpen); !openResult {
		t.Fatal("failed to pass DB.Open tests, terminating early")
	}

	// if open is verified to work, we can safely proceed
	// with the remaining tests with the dbTester suite.
	db := dbTester{}
	db.DB, _ = lib.OpenFromEnv(context.Background())

	// run DB.Get tests first to ensure that proceeding
	// tests function properly.
	if passed := t.Run("getUser", db.testGetUser) && t.Run("getProgram", db.testGetProgram); !passed {
		t.Fatal("failed to pass DB.Get tests, terminating early")
	}
}

// test that a database client can be opened properly.
func testOpen(t *testing.T) {
	// create a firestore client.
	ctx := context.Background()

	// test with empty credentials path.
	if d, err := lib.OpenFromCreds(ctx, ""); err == nil && d == nil {
		t.Fatalf("returned nil db client despite raising no error")
	}

	// should be able to open client from environment variables.
	t.Logf("using environment variable $%s with value %s", lib.DefaultEnvVar, os.Getenv(lib.DefaultEnvVar))
	if d, err := lib.OpenFromEnv(ctx); err != nil || d == nil {
		t.Fatalf("error raised with assumed config path")
	}
}

// test DB.GetUser for functionality.
func (d *dbTester) testGetUser(t *testing.T) {
	// test with nil uid.
	uid := ""
	if _, err := d.GetUser(context.Background(), uid); err == nil {
		t.Errorf("incorrectly returned successfully from nil UID")
	}

	// acquire uid from arguments.
	uid = os.Args[len(os.Args)-2]
	t.Logf("trying to open user document with UID '%s'", uid)
	if uid == "" {
		t.Fatalf("existing uid not provided, please run tests with `go test -args [existing user id] [existing program id]`")
	}

	// test with provided uid.
	if _, err := d.GetUser(context.Background(), uid); err != nil {
		t.Errorf("failed to retrieve user object for assumed real UID '%s'. %s", uid, err)
	}
}

func (d *dbTester) testGetProgram(t *testing.T) {
	// test with nil uid.
	uid := ""
	if _, err := d.GetProgram(context.Background(), uid); err == nil {
		t.Errorf("incorrectly returned successfully from nil UID")
	}

	// acquire uid from arguments.
	uid = os.Args[len(os.Args)-1]
	t.Logf("trying to open program document with UID '%s'", uid)
	if uid == "" {
		t.Fatalf("existing uid not provided, please run tests with `go test -args [existing user id] [existing program id]`")
	}

	// test with provided uid.
	if _, err := d.GetProgram(context.Background(), uid); err != nil {
		t.Errorf("failed to rerieve program for assumed real UID '%s'. %s", uid, err)
	}
}