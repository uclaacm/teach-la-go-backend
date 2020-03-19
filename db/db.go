package db

import (
	"context"
	"fmt"
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

// OpenFromEnv returns a pointer to a database client based on
// JSON credentials given by the environment variable.
// Returns an error if it fails at any point.
func OpenFromEnv(ctx context.Context) (*DB, error) {
	cfg := os.Getenv("TLACFG")
	if cfg == "" {
		return nil, fmt.Errorf("no $TLACFG environment variable provided")
	}

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsJSON([]byte(cfg))
	app, err := firebase.NewApp(ctx, nil, opt)

	// acquire the firestore client, fail if we cannot.
	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}

// CreateUser creates the default user and program documents,
// then returns the object for said User.
func (d *DB) CreateUser(ctx context.Context) (*User, error) {
	// create new doc for user
	doc := d.Collection(UsersPath).NewDoc()

	// create structures to be used as default data
	newUser, newProgs := defaultData()
	newUser.UID = doc.ID

	// create all new programs and associate them to the user.
	for _, prog := range newProgs {
		// create program in database.
		newProg := d.Collection(ProgramsPath).NewDoc()
		newProg.Set(context.Background(), prog)

		// establish association in user doc.
		newUser.Programs = append(newUser.Programs, newProg.ID)
	}

	// create user doc.
	_, err := doc.Set(ctx, newUser)
	return &newUser, err
}

// GetUser returns a user document in struct form,
// with an error if one occurs.
func (d *DB) GetUser(ctx context.Context, uid string) (*User, error) {
	doc, err := d.Collection(UsersPath).Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	u := &User{UID: uid}
	return u, doc.DataTo(u)
}

// UpdateUser updates the user document with given uid to match
// the provided struct.
// An error is returned should one occur.
func (d *DB) UpdateUser(ctx context.Context, uid string, u *User) error {
	doc := d.Collection(UsersPath).Doc(uid)

	_, err := doc.Update(ctx, u.ToFirestoreUpdate())
	return err
}

// DeleteProgramFromUser takes a uid and a pid,
// and deletes the pid from the User with the given uid
func (d *DB) DeleteProgramFromUser(ctx context.Context, uid string, pid string) (err error) {
	doc := d.Collection(UsersPath).Doc(uid)

	var ud *firestore.DocumentSnapshot
	u := User{}
	if ud, err = doc.Get(ctx); err == nil {
		if err = ud.DataTo(&u); err != nil {
			return
		}
	} else {
		return
	}

	_, err = doc.Update(ctx, []firestore.Update{
		{Path: "mostRecentProgram", Value: int(-1)},
		{Path: "programs", Value: firestore.ArrayRemove(pid)},
	})

	return
}

// AddProgramToUser takes a uid and a pid,
// and adds the pid to the user's list of programs
func (d *DB) AddProgramToUser(ctx context.Context, uid string, pid string) error {

	//get the user doc
	doc := d.Collection(UsersPath).Doc(uid)

	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "programs", Value: firestore.ArrayUnion(pid)},
	})

	return err

}

// CreateProgram creates a new program document to match
// the provided struct.
// The program's UID is returned with an error, should one
// occur.
func (d *DB) CreateProgram(ctx context.Context, p *Program) (string, error) {
	doc := d.Collection(ProgramsPath).NewDoc()

	// update UID to match, then update doc.
	p.UID = doc.ID
	_, err := doc.Set(ctx, *p)
	return doc.ID, err
}

// GetProgram returns a program document in struct form,
// with an error if one occurs.
func (d *DB) GetProgram(ctx context.Context, uid string) (*Program, error) {
	doc, err := d.Collection(ProgramsPath).Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	p := &Program{UID: uid}
	return p, doc.DataTo(p)
}

// UpdateProgram updates the program with the given uid to match
// the program provided as an argument.
// An error is returned if any issues are encountered.
func (d *DB) UpdateProgram(ctx context.Context, uid string, p *Program) error {
	doc := d.Collection(ProgramsPath).Doc(uid)

	_, err := doc.Update(ctx, p.ToFirestoreUpdate())
	return err
}

// DeleteProgram deletes the program with the given uid. An error
// is returned should one occur.
func (d *DB) DeleteProgram(ctx context.Context, uid string) error {
	doc := d.Collection(ProgramsPath).Doc(uid)

	_, err := doc.Delete(ctx)
	return err
}
