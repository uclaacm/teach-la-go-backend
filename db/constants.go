package db

import (
	"math/rand"
	"os"
	"time"
)

const (
	// DefaultEnvVar describes the default environment
	// variable used to open a connection to the database.
	DefaultEnvVar = "TLACFG"

	python = iota
	processing
	html
	react
	langCount

	// the number of program thumbnails available to choose from.
	thumbnailCount = 58

	// programsPath describes the path to the program
	// management endpoint.
	programsPath = "programs"

	// usersPath describes the path to the user management
	// endpoint
	usersPath = "users"

	// classesPath describes the path to the classes
	// management endpoint.
	classesPath = "classes"

	// classesAliasPath describes the path to the collection with 3 word id => hash mapping for classes
	classesAliasPath = "classes_alias"

	shardName    = "--shards--"
	numShards    = 8                                 // number of shards
	aliasSize    = int64(16777216)                   // number of total unique IDs we can allocate
	divider      = int64(1024)                       // a factor used to divide each shards into "blocks"
	maxSize      = aliasSize / (divider * numShards) // number of blocks per shard
	slotPerShard = aliasSize / numShards             // how many IDs we have per shard
	shardCap     = slotPerShard

	msgTypeRead  = "READ"
	msgTypeError = "ERROR"
)

var EnableBetaFeatures = os.Getenv("ENABLE_BETA_FEATURES")

func langString(langCode int) string {
	switch langCode {
	case python:
		return "python"
	case processing:
		return "processing"
	case html:
		return "html"
	case react:
		return "react"
	default:
		return "DNE"
	}
}

// defaultProgram returns a Program struct initialized to
// default values for a given Language.
// if the language does not exist, it returns nil.
func defaultProgram(language string) (defaultProg Program) {
	defaultCode := ""

	switch language {
	case "python":
		defaultCode = "import turtle\n\nt = turtle.Turtle()\n\nt.color('red')\nt.forward(75)\nt.left(90)\n\n\nt.color('blue')\nt.forward(75)\nt.left(90)\n"
	case "processing":
		defaultCode = "function setup() {\n  createCanvas(400, 400);\n}\n\nfunction draw() {\n  background(220);\n  ellipse(mouseX, mouseY, 100, 100);\n}"
	case "html":
		defaultCode = "<html>\n  <head>\n  </head>\n  <body>\n    <div style='width: 100px; height: 100px; background-color: black'>\n    </div>\n  </body>\n</html>"
	case "react":
		defaultCode = "const {\n  Button,\n} = MaterialUI;\n\nconst App = () => (\n  <LikeButton />\n);\n\nconst LikeButton = () => {\n  const [liked, setLiked] = React.useState(false);\n\n  if (liked) {\n    return 'You liked this.';\n  }\n\n  return <Button variant=\"contained\" onClick={() => setLiked(true)}>Like</Button>;\n}"
	default:
		return Program{}
	}

	defaultProg.Code = defaultCode
	defaultProg.Language = language
	defaultProg.Name = language
	defaultProg.DateCreated = time.Now().UTC().String()
	defaultProg.Thumbnail = rand.Int63n(thumbnailCount)
	return defaultProg
}

// defaultData is the factory function
// for constructing default UserData structs
// and its associated Programs. Associations
// between said UserData and Programs are not
// automatically applied in the database.
func defaultData() (User, []Program) {
	defaultProgs := make([]Program, 0)
	for i := python; i < langCount; i++ {
		defaultProgs = append(defaultProgs, defaultProgram(langString(i)))
	}

	u := User{
		DisplayName: "J Bruin",
		PhotoName:   "icecream",
	}
	return u, defaultProgs
}
