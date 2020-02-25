package lib

import (
	"context"
	"os"
	"strings"

	"../tools/tinycrypt"

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
	// check, using os.Stat(), that the file exists. If it does not exist,
	// then fail.
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsFile(path)
	app, err := firebase.NewApp(ctx, nil, opt)

	// acquire the firestore client, fail if we cannot.
	client, err := app.Firestore(ctx)
	return &DB{Client: client}, err
}

// OpenFromEnv calls OpenFromCreds with the path
// provided by your environment variable $CFGPATH.
func OpenFromEnv(ctx context.Context) (*DB, error) {
	return OpenFromCreds(ctx, os.Getenv(DefaultEnvVar))
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
func (d *DB) DeleteProgramFromUser(ctx context.Context, uid string, pid string) error {
	
	//get the user doc
	doc := d.Collection(UsersPath).Doc(uid)

	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "programs", Value: firestore.ArrayRemove(pid)},
	})

	return err
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
func (d *DB) CreateProgram(ctx context.Context, p *Program/*, uid string*/) (string, error) {
	
	//crate new program document
	doc := d.Collection(ProgramsPath).NewDoc()

	// update UID to match, then update doc.
	p.UID = doc.ID
	_, err := doc.Set(ctx, *p)
	return p.UID, err
}

// GetProgram returns a program document in struct form,
// with an error if one occurs.
func (d *DB) GetProgram(ctx context.Context, pid string) (*Program, error) {
	doc, err := d.Collection(ProgramsPath).Doc(pid).Get(ctx)
	if err != nil {
		return nil, err
	}

	p := &Program{UID: pid}
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


// CreateClass creates a new class document to match the provided struct.
// The class's UID is returned with an error, should one occur.
func (d *DB) CreateClass(ctx context.Context, c *Class) (string, error) {
	// create a new doc for this class
	doc := d.Collection(ClassesPath).NewDoc()

	// set the CID parameter
	c.CID = doc.ID

	// update database
	_, err := doc.Set(ctx, *c)

	//return the results
	return c.CID, err
}

// AddClassToUser takes a uid and a pid, 
// and adds the pid to the user's list of programs
func (d *DB) AddClassToUser(ctx context.Context, uid string, cid string) error {

	//get the user doc
	doc := d.Collection(UsersPath).Doc(uid) 

	//add the class id
	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "classes", Value: firestore.ArrayUnion(cid)},
	})

	return err

}

func (d *DB) AddUserToClass(ctx context.Context, uid string, cid string) error {

	//get the class doc
	doc := d.Collection(ClassesPath).Doc(cid) 

	//add the class id
	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "members", Value: firestore.ArrayUnion(uid)},
	})

	return err

}

func (d *DB) RemoveUserFromClass(ctx context.Context, uid string, cid string) error {

	//get the class doc
	doc := d.Collection(ClassesPath).Doc(cid) 

	//add the class id
	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "members", Value: firestore.ArrayRemove(uid)},
	})

	return err

}

func (d *DB) RemoveClassFromUser(ctx context.Context, uid string, cid string) error {

	//get the user doc
	doc := d.Collection(UsersPath).Doc(uid) 

	//remove the class id
	_, err := doc.Update(ctx, []firestore.Update{
		{Path: "classes", Value: firestore.ArrayRemove(cid)},
	})

	return err

}


func (d *DB) GetClass(ctx context.Context, cid string) (*Class, error) {

	//get document for specified class
	doc, err := d.Collection(ClassesPath).Doc(cid).Get(ctx)
	if err != nil {
		return nil, err
	}

	// create struct to populate
	c := &Class{}
	// populate struct
	if err := doc.DataTo(c); err != nil {
		return nil, err
	}
	
	return c, err
}


func (d *DB) MakeAlias(ctx context.Context, pid string) (string, error) {

	// convert pid into a 36 bit hash
	aid := MakeHash(pid) 

	//get the mapping collection
	col := d.Collection(AliasPath)
	//get the snapshot of the document
	snap, err := col.Doc(pid).Get(ctx)

	//if the doc id is taken, try different numbers
	for snap.Exists() == true {
		aid+=1
		if aid == 0xFFFFFFFFF{
			aid = 0
		}
		snap, err = col.Doc(pid).Get(ctx)
	}

	// create a new doc for this alias
	//convert the aid into a word
	s := strings.Join(tinycrypt.GenerateWord36(aid), ", ")
	//create mapping
	_, err = col.Doc(s).Set(ctx, map[string]interface{}{
		"target" : pid,
	})

	return s, err

}

// Take the 120 bit pid and makes a 36 bit aid
func MakeHash(pid string) (uint64) {
	
	//get first 6 chars
	pid = pid[0:6]
	index := uint64(0)
	aid := uint64(0)

	//convert their UNICODE into 6 bit ints
	for _, c := range pid{
		//if this is a number (0 ~ 9)
		if c >= 0x0030 && c <= 0x0039 {
			index = uint64(c - 0x0030) 
		} else if c >= 0x0041 && c <= 0x005A{ //else if this is uppercase char
			index = uint64(c - 0x0041) 
		} else if c >= 0x0061 && c <= 0x007A{ //else if this is lowercase char
			index = uint64(c - 0x0061)
		} else {
			index = 0
		}

		aid = (aid | index) << 6
	}

	return aid

}

