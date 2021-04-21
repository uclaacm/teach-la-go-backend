package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// DB implements the TLADB interface on a Firestore
// database.
type DB struct {
	// Primary database connection.
	*firestore.Client
}

type MockDB struct {
	// "Users, Programs, Class" collection
	db map[string]map[string]interface{}
}

func (d *MockDB) LoadProgram(_ context.Context, pid string) (Program, error) {
	p, ok := d.db[programsPath][pid].(Program)
	if !ok {
		return Program{}, errors.New("program has not been created")
	}
	return p, nil
}

func (d *MockDB) StoreProgram(_ context.Context, p Program) error {
	d.db[programsPath][p.UID] = p
	return nil
}

func (d *MockDB) LoadClass(_ context.Context, cid string) (c Class, err error) {
	c, ok := d.db[classesPath][cid].(Class)
	if !ok {
		err = errors.New("invalid class ID")
	}
	return
}

func (d *MockDB) StoreClass(_ context.Context, c Class) error {
	d.db[classesPath][c.CID] = c
	return nil
}

func (d *MockDB) LoadUser(_ context.Context, uid string) (u User, err error) {
	u, ok := d.db[usersPath][uid].(User)
	if !ok {
		err = errors.New("invalid user ID")
	}
	return
}

func (d *MockDB) StoreUser(_ context.Context, u User) error {
	d.db[usersPath][u.UID] = u
	return nil
}

func OpenMock() MockDB {
	m := MockDB{db: make(map[string]map[string]interface{})}
	m.db[usersPath] = make(map[string]interface{})
	m.db[programsPath] = make(map[string]interface{})
	m.db[classesPath] = make(map[string]interface{})
	return m
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
