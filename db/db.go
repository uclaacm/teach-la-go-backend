package db

import (
	"context"
	"fmt"
	"os"
	"strings"

	tinycrypt "github.com/uclaacm/teach-la-go-backend-tinycrypt"

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
	if err != nil {
		return nil, err
	}

	// acquire the firestore client, fail if we cannot.
	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}

// CreateUser creates the default user and program documents,
// then returns the object for said User.
func (d *DB) CreateUser(ctx context.Context) (*User, error) {
	// create new doc for user
	ref := d.Collection(UsersPath).NewDoc()

	// create structures to be used as default data
	newUser, newProgs := defaultData()
	newUser.UID = ref.ID

	// create all new programs and associate them to the user.
	for _, prog := range newProgs {
		// create program in database.
		newProg := d.Collection(ProgramsPath).NewDoc()
		newProg.Set(context.Background(), prog)

		// establish association in user doc.
		newUser.Programs = append(newUser.Programs, newProg.ID)
	}

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Create(ref, newUser)
	})

	return &newUser, err
}

// GetUser returns a user document in struct form,
// with an error if one occurs.
func (d *DB) GetUser(ctx context.Context, uid string) (*User, error) {
	ref := d.Collection(UsersPath).Doc(uid)
	u := &User{UID: uid}

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}

		return doc.DataTo(u)
	})

	return u, err
}

// UpdateUser updates the user document with given uid to match
// the provided struct.
// An error is returned should one occur.
func (d *DB) UpdateUser(ctx context.Context, uid string, u *User) error {
	ref := d.Collection(UsersPath).Doc(uid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, u.ToFirestoreUpdate())
	})
}

// DeleteProgramFromUser takes a uid and a pid,
// and deletes the pid from the User with the given uid
func (d *DB) DeleteProgramFromUser(ctx context.Context, uid string, pid string) error {

	//get the user doc
	ref := d.Collection(UsersPath).Doc(uid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "programs", Value: firestore.ArrayRemove(pid)},
		})
	})
}

// AddProgramToUser takes a uid and a pid,
// and adds the pid to the user's list of programs
func (d *DB) AddProgramToUser(ctx context.Context, uid string, pid string) error {

	//get the user doc
	ref := d.Collection(UsersPath).Doc(uid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "programs", Value: firestore.ArrayUnion(pid)},
		})
	})
}

// CreateProgram creates a new program document to match
// the provided struct.
// The program's UID is returned with an error, should one
// occur.
func (d *DB) CreateProgram(ctx context.Context, p *Program) (string, error) {
	ref := d.Collection(ProgramsPath).NewDoc()

	// update UID to match, then update doc.
	p.UID = ref.ID

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Set(ref, *p)
	})

	return p.UID, err
}

// GetProgram returns a program document in struct form,
// with an error if one occurs.
func (d *DB) GetProgram(ctx context.Context, pid string) (*Program, error) {
	ref := d.Collection(ProgramsPath).Doc(pid)

	p := &Program{UID: pid}

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}

		return doc.DataTo(p)
	})

	return p, err
}

// UpdateProgram updates the program with the given uid to match
// the program provided as an argument.
// An error is returned if any issues are encountered.
func (d *DB) UpdateProgram(ctx context.Context, uid string, p *Program) error {
	ref := d.Collection(ProgramsPath).Doc(uid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, p.ToFirestoreUpdate())
	})
}

// DeleteProgram deletes the program with the given uid. An error
// is returned should one occur.
func (d *DB) DeleteProgram(ctx context.Context, uid string) error {
	ref := d.Collection(ProgramsPath).Doc(uid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Delete(ref)
	})
}

// CreateClass creates a new class document to match the provided struct.
// The class's UID is returned with an error, should one occur.
func (d *DB) CreateClass(ctx context.Context, c *Class) (string, error) {
	// create a new doc for this class
	ref := d.Collection(ClassesPath).NewDoc()

	//set the CID parameter
	c.CID = ref.ID

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Set(ref, *c)
	})

	//return the results
	return ref.ID, err
}

func (d *DB) UpdateClassWID(ctx context.Context, cid string, wid string) error {
	ref := d.Collection(ClassesPath).Doc(cid)

	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "WID", Value: wid},
		})
	})
}

// AddClassToUser takes a uid and a pid,
// and adds the pid to the user's list of programs
func (d *DB) AddClassToUser(ctx context.Context, uid string, cid string) error {
	//get the user doc
	ref := d.Collection(UsersPath).Doc(uid)

	//add the class id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "classes", Value: firestore.ArrayUnion(cid)},
		})
	})
}

// AddUserToClass add an uid to a given class
func (d *DB) AddUserToClass(ctx context.Context, uid string, cid string) error {
	//get the class doc
	ref := d.Collection(ClassesPath).Doc(cid)

	//add the user id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "members", Value: firestore.ArrayUnion(uid)},
		})
	})
}

// RemoveUserFromClass removes an uid from a given class
func (d *DB) RemoveUserFromClass(ctx context.Context, uid string, cid string) error {
	//get the class doc
	ref := d.Collection(ClassesPath).Doc(cid)

	//remove the user id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "members", Value: firestore.ArrayRemove(uid)},
		})
	})
}

// RemoveClassFromUser removes a class from a given user
func (d *DB) RemoveClassFromUser(ctx context.Context, uid string, cid string) error {
	//get the user doc
	ref := d.Collection(UsersPath).Doc(uid)

	//remove the class id
	return d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(ref, []firestore.Update{
			{Path: "classes", Value: firestore.ArrayRemove(cid)},
		})
	})

}

// GetClass takes a cid, and returns a Class struct with its parameters populated
// The retuned value is a pointer to the struct instantiated in this function
func (d *DB) GetClass(ctx context.Context, cid string) (*Class, error) {
	//get the class doc
	ref := d.Collection(ClassesPath).Doc(cid)

	//create struct to populate
	c := &Class{}

	//populate struct
	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}

		return doc.DataTo(c)
	})

	return c, err
}

// MakeAlias takes an id (usually pid or cid), allocates a 3 word id(wid), and
// stores it in Firebase. The generated wid is a string, with words comma seperated
func (d *DB) MakeAlias(ctx context.Context, uid string, path string) (string, error) {

	// convert uid into a 36 bit hash
	//aid := tinycrypt.MakeHash(uid)
	aid := tinycrypt.GenerateHash()

	// convert that to a 3 word id
	wid_list := tinycrypt.GenerateWord36(aid)
	// the result is an array,so concat into a single string
	wid := strings.Join(wid_list, ",")

	// get the mapping collection
	col := d.Collection(path)

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(col.Doc(wid))
		if err != nil {
			return err
		}

		//if the doc id is taken, generate a different wid
		for doc.Exists() == true {

			aid++
			if aid >= 0xFFFFFFFFF {
				aid = 0
			}

			wid_list = tinycrypt.GenerateWord36(aid)
			wid = strings.Join(wid_list, ",")

			doc, err = tx.Get(col.Doc(wid))
			if err != nil {
				return err
			}
		}

		//create mapping
		return tx.Set(col.Doc(wid), map[string]interface{}{
			"target": uid,
		})
	})

	return strings.Join(wid_list, ","), err

}

// GetUIDFromWID returns the UID given a WID
func (d *DB) GetUIDFromWID(ctx context.Context, wid string, path string) (string, error) {

	// get the document with the mapping
	ref := d.Collection(path).Doc(wid)

	t := struct {
		Target string `firestore:target`
	}{}

	err := d.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}

		return doc.DataTo(&t)
	})

	return t.Target, err
}
