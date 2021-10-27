package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DB implements the TLADB interface on a Firestore
// database.
type DB struct {
	// Primary database connection.
	*firestore.Client
}

func (d *DB) LoadProgram(ctx context.Context, pid string) (Program, error) {
	doc, err := d.Collection(programsPath).Doc(pid).Get(ctx)
	if err != nil {
		return Program{}, err
	}

	p := Program{}
	if err := doc.DataTo(&p); err != nil {
		return Program{}, err
	}
	return p, nil
}

func (d *DB) StoreProgram(ctx context.Context, p Program) error {
	if _, err := d.Collection(programsPath).Doc(p.UID).Set(ctx, &p); err != nil {
		return err
	}
	return nil
}

func (d *DB) LoadClass(ctx context.Context, cid string) (Class, error) {
	doc, err := d.Collection(classesPath).Doc(cid).Get(ctx)
	if err != nil {
		return Class{}, err
	}

	c := Class{}
	if err := doc.DataTo(&c); err != nil {
		return Class{}, err
	}
	return c, nil
}

func (d *DB) StoreClass(ctx context.Context, c Class) error {
	if _, err := d.Collection(classesPath).Doc(c.CID).Set(ctx, &c); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteClass(ctx context.Context, c Class) error {
	classRef := d.Collection(classesPath).Doc(c.CID)
	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		classSnap, err := tx.Get(classRef)
		if err != nil {
			return err
		}

		class := Class{}
		if err := classSnap.DataTo(&class); err != nil {
			return err
		}
		for _, prog := range class.Programs {
			progRef := d.Collection(programsPath).Doc(prog)
			// if we can't find a program, then it's not a problem.
			if err := tx.Delete(progRef); status.Code(err) != codes.NotFound {
				return err
			}
		}

		return tx.Delete(classRef)
	})
	if err != nil {
		return err
	}
	
	return nil
}

func (d *DB) LoadUser(ctx context.Context, uid string) (User, error) {
	doc, err := d.Collection(usersPath).Doc(uid).Get(ctx)
	if err != nil {
		return User{}, err
	}

	u := User{}
	if err := doc.DataTo(&u); err != nil {
		return User{}, err
	}
	return u, nil
}

func (d *DB) StoreUser(ctx context.Context, u User) error {
	if _, err := d.Collection(usersPath).Doc(u.UID).Set(ctx, &u); err != nil {
		return err
	}
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
