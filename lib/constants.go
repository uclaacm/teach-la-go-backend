package lib

import (
	"errors"
	"log"
	"time"
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

	// ProgEndpt describes the path to the program
	// management endpoint.
	ProgEndpt = "programs"

	// UserEndpt describes the path to the user management
	// endpoint
	UserEndpt = "users"
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
		log.Printf("language does not exist.")
		return
	}

	defaultProg.Code = defaultCode
	defaultProg.Language, _ = LanguageName(languageCode)
	defaultProg.DateCreated = time.Now().UTC()

	return
}

// defaultData is the factory function
// for constructing default UserData structs
// and its associated Programs. Associations
// between said UserData and Programs are not
// automatically applied in the database.
func defaultData() (*UserData, []Program) {
	var defaultProgs []Program
	for k := 0; k < LanguageCount; k++ {
		defaultProgs = append(defaultProgs, defaultProgram(k))
	}

	u := UserData{
		DisplayName: "J Bruin",
	}
	return &u, defaultProgs
}
