package lib

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// DB describes a common interface for any database
// in use by our application.
type DB struct {
	*firestore.Client
}

// OpenFromCreds returns a pointer to a database client based on
// JSON credentials pointed to by the provided path.
// Returns an error if it fails at any point.
func OpenFromCreds(ctx context.Context, path string) (*DB, error) {
	// load environment variable $CFGPATH as a string: the path to our config.
	configPath := path

	// check, using os.Stat(), that the file exists. If it does not exist,
	// then fail.
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(ctx, nil, opt)

	// acquire the firestore client, fail if we cannot.
	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}

// OpenFromEnv calls OpenFromCreds with the path
// provided by your environment variable $CFGPATH.
func OpenFromEnv(ctx context.Context) (*DB, error) {
	return OpenFromCreds(ctx, os.Getenv("CFGPATH"))
}

// CreateUser creates the default user and program documents,
// then returns the uid for said user.
func (d *DB) CreateUser(ctx context.Context) (string, error) {
	doc := d.Collection(UserEndpt).NewDoc()

	newUser, newProgs := defaultData()

	// create all new programs and associate them to the user.
	for _, prog := range newProgs {
		// create program in database.
		newProg := d.Collection(ProgEndpt).NewDoc()
		newProg.Set(context.Background(), prog)

		// establish association in user doc.
		newUser.Programs = append(newUser.Programs, newProg.ID)
	}

	// create user doc.
	_, err := doc.Set(ctx, newUser)
	return doc.ID, err
}

// GetUser returns a user document in struct form,
// with an error if one occurs.
func (d *DB) GetUser(ctx context.Context, uid string) (*User, error) {
	doc, err := d.Collection(UserEndpt).Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	u := &User{}
	return u, doc.DataTo(u)
}

// UpdateUser updates the user document with given uid to match
// the provided struct.
// An error is returned should one occur.
func (d *DB) UpdateUser(ctx context.Context, uid string, u *User) error {
	doc := d.Collection(UserEndpt).Doc(uid)

	_, err := doc.Update(ctx, u.ToFirestoreUpdate())
	return err
}

// CreateProgram creates a new program document to match
// the provided struct.
// The program's UID is returned with an error, should one
// occur.
func (d *DB) CreateProgram(ctx context.Context, p *Program) (string, error) {
	doc := d.Collection(ProgEndpt).NewDoc()

	_, err := doc.Set(ctx, *p)
	return doc.ID, err
}

// GetProgram returns a program document in struct form,
// with an error if one occurs.
func (d *DB) GetProgram(ctx context.Context, uid string) (*Program, error) {
	doc, err := d.Collection(ProgEndpt).Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	p := &Program{}
	return p, doc.DataTo(p)
}

// UpdateProgram updates the program with the given uid to match
// the program provided as an argument.
// An error is returned if any issues are encountered.
func (d *DB) UpdateProgram(ctx context.Context, uid string, p *Program) error {
	doc := d.Collection(ProgEndpt).Doc(uid)

	_, err := doc.Update(ctx, p.ToFirestoreUpdate())
	return err
}

// DeleteProgram deletes the program with the given uid. An error
// is returned should one occur.
func (d *DB) DeleteProgram(ctx context.Context, uid string) error {
	doc := d.Collection(ProgEndpt).Doc(uid)

	_, err := doc.Delete(ctx)
	return err
}
