package db

import (
	"context"
	// "errors"
	"github.com/pkg/errors"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

/// NOTE:
/// The *Xyz*Transact() functions are equivalent in behavior to *Xyz*(),
/// except they operate on a transaction instead of a context and should
/// be called within a RunTransaction() callback

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

func (d *DB) CreateUser(ctx context.Context, u User) (User, error) {
	// create a new doc for the user if necessary
	ref := d.Collection(usersPath).NewDoc()
	if u.UID != "" {
		ref = d.Collection(usersPath).Doc(u.UID)
	}
	u.UID = ref.ID

	userSnap, _ := ref.Get(ctx)
	if userSnap.Exists() {
		return u, errors.Errorf("user document with uid '%s' already initialized", u.UID)
	}
	// If there was an error in creating the user, return the error
	if _, err := ref.Create(ctx, u); err != nil {
		return u, err
	}

	// Return the user
	return u, nil
}

func (d *DB) CreateProgram(ctx context.Context, p Program) (Program, error) {
	newProg := d.Collection(programsPath).NewDoc()
	p.UID = newProg.ID
	if _, err := newProg.Create(ctx, p); err != nil {
		return p, err
	}

	return p, nil
}

func (d *DB) CreateProgramTransact(tx *firestore.Transaction, p Program) (Program, error) {
	newProg := d.Collection(programsPath).NewDoc()
	p.UID = newProg.ID
	if err := tx.Create(newProg, p); err != nil {
		return p, err
	}

	return p, nil
}

// Create a program and associate it with a user and a class.
//
// If wid == "", will not attempt to join a class.
func (d *DB) CreateProgramAndAssociate(ctx context.Context, p Program, uid string, wid string) error {
	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// create program
		pRef, err := d.CreateProgramTransact(tx, p)
		if err != nil {
			return err
		}

		// associate to user, if they exist
		u, err := d.LoadUserTransact(tx, uid)
		if err != nil {
			return err
		}

		u.Programs = append(u.Programs, pRef.UID)

		if err := d.StoreUserTransact(tx, u); err != nil {
			return err
		}

		// associate to class, if they exist
		var cid string
		var class Class
		if wid != "" {
			cid, err = d.GetUIDFromWIDTransact(tx, wid, ClassesAliasPath)
			if err != nil {
				return err
			}

			class, err = d.LoadClassTransact(tx, cid)
			if err != nil {
				return err
			}

			class.Programs = append(class.Programs, pRef.UID)

			p.WID = class.WID

			err := d.StoreClassTransact(tx, class)
			if err != nil {
				return err
			}
		}

		p.UID = pRef.UID

		return nil
	})

	return err
}

func (d *DB) RemoveProgram(ctx context.Context, pid string) error {
	if _, err := d.Collection(programsPath).Doc(pid).Delete(ctx); err != nil {
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

func (d *DB) LoadClassTransact(tx *firestore.Transaction, cid string) (Class, error) {
	doc, err := tx.Get(d.Collection(classesPath).Doc(cid))
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

func (d *DB) StoreClassTransact(tx *firestore.Transaction, c Class) error {
	if err := tx.Set(d.Collection(classesPath).Doc(c.CID), &c); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteClass(ctx context.Context, cid string) error {
	if _, err := d.Collection(classesPath).Doc(cid).Delete(ctx); err != nil {
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

func (d *DB) LoadUserTransact(tx *firestore.Transaction, uid string) (User, error) {
	doc, err := tx.Get(d.Collection(usersPath).Doc(uid))
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

func (d *DB) StoreUserTransact(tx *firestore.Transaction, u User) error {
	if err := tx.Set(d.Collection(usersPath).Doc(u.UID), &u); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteUser(ctx context.Context, uid string) error {
	if _, err := d.Collection(usersPath).Doc(uid).Delete(ctx); err != nil {
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
