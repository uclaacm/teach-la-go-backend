package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// DB describes a common type for our database operations.
type DB struct {
	*firestore.Client
}

// Open returns a pointer to a new database client based on
// JSON credentials given by the environment variable.
// Returns an error if it fails at any point.
func Open(ctx context.Context, cfg string) (*DB, error) {
	if cfg == "" {
		return nil, errors.New("config variable is required")
	}

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsJSON([]byte(cfg))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	// acquire the firestore client, fail if we cannot.
	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}

// OpenFromJSON returns a pointer to a new database client based
// on a JSON file given by the provided path.
// Returns an error if it fails at any point.
func OpenFromJSON(ctx context.Context, path string) (*DB, error) {
	opt := option.WithCredentialsFile(path)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}
