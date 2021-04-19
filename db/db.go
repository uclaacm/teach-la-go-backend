package db

import (
	"context"
	"errors"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// DB implements the TLADB interface on a Firestore
// database. It is thread-safe.
type DB struct {
	// Primary database connection.
	*firestore.Client

	// Enforces atomicity at a lower level.
	*sync.Map
}

func (d *DB) LoadProgram(ctx context.Context, p string) (Program, error) {
	return Program{}, nil
}

func (d *DB) StoreProgram(ctx context.Context, p *Program) error {
	return nil
}

func (d *DB) LoadClass(context.Context, string) (Class, error) {
	return Class{}, nil
}

func (d *DB) StoreClass(context.Context, *Class) error {
	return nil
}

func (d *DB) LoadUser(context.Context, string) (User, error) {
	return User{}, nil
}

func (d *DB) StoreUser(context.Context, *User) error {
	return nil
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
