package db

import (
	"errors"
	"time"

	"github.com/uclaacm/teach-la-go-backend/tools/log"
)

const (
	notALanguage = -1
	python       = iota
	processing
	html

	// LanguageCount is the number of programming languages
	// available.
	LanguageCount

	// ThumbnailCount describes the number of program
	// thumbnails available to choose from.
	ThumbnailCount = 58

	// DefaultEnvVar describes the default environment
	// variable used by the server.
	DefaultEnvVar = "CFGPATH"

	// ProgramsPath describes the path to the program
	// management endpoint.
	ProgramsPath = "programs"

	// UsersPath describes the path to the user management
	// endpoint
	UsersPath = "users"

	// ClassesPath describes the path to the classes
	// management endpoint.
	ClassesPath = "classes"

	// ProgramsAliasPath describes the path to the collection with 3 word id => hash mapping for programs
	ProgramsAliasPath = "programs_alias"

	// ClassesAliasPath describes the path to the collection with 3 word id => hash mapping for classes
	ClassesAliasPath = "classes_alias"

	// num_shards = 32
	// alias_size = 16777216
	// divider = 1024
	num_shards = 8
	alias_size = int64(64)
	divider = int64(4)
	max_size = alias_size / (divider * num_shards)
	slot_per_shard = alias_size / num_shards
	shard_cap = slot_per_shard

	
)

// LanguageName acquires the name for the language desecribed
// by the code, returning an error if such a language does not
// exist.
func LanguageName(code int) (string, error) {
	switch code {
	case python:
		return "python", nil

	case processing:
		return "processing", nil

	case html:
		return "html", nil
	}

	return "", errors.New("language does not exist")
}

// LanguageCode acquires the code for the language described
// by the string, returning an error if such a language does not
// exist.
func LanguageCode(name string) (int, error) {
	switch name {
	case "python":
		return python, nil

	case "processing":
		return processing, nil

	case "html":
		return html, nil
	}

	return notALanguage, errors.New("language does not exist")
}

// defaultProgram returns a Program struct initialized to
// default values for a given Language.
func defaultProgram(languageCode int) (defaultProg Program) {
	var defaultCode string

	switch languageCode {
	case python:
		defaultCode = "import turtle\n\nt = turtle.Turtle()\n\nt.color('red')\nt.forward(75)\nt.left(90)\n\n\nt.color('blue')\nt.forward(75)\nt.left(90)\n"
	case processing:
		defaultCode = "function setup() {\n  createCanvas(400, 400);\n}\n\nfunction draw() {\n  background(220);\n  ellipse(mouseX, mouseY, 100, 100);\n}"
	case html:
		defaultCode = "<html>\n  <head>\n  </head>\n  <body>\n    <div style='width: 100px; height: 100px; background-color: black'>\n    </div>\n  </body>\n</html>"
	case notALanguage:
		log.Debugf("language does not exist.")
		return
	}

	//defaultProg.UID = uid
	defaultProg.Code = defaultCode
	defaultProg.Language, _ = LanguageName(languageCode)
	defaultProg.DateCreated = time.Now().UTC().String()

	return
}

// defaultData is the factory function
// for constructing default UserData structs
// and its associated Programs. Associations
// between said UserData and Programs are not
// automatically applied in the database.
func defaultData() (User, []Program) {
	var defaultProgs []Program
	for k := 0; k < LanguageCount; k++ {
		defaultProgs = append(defaultProgs, defaultProgram(k))
	}

	u := User{
		DisplayName: "J Bruin",
		//UID:		 	uid,
	}
	return u, defaultProgs
}
