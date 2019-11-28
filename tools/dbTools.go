package tools

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// GetDB returns a pointer to a firestore *Client based on the
// JSON credentials pointed to by the environment variable
// $CFGPATH.
func GetDB() (client *firestore.Client) {
	// load environment variable $CFGPATH as a string: the path to our config.
	configPath := os.Getenv("CFGPATH")

	// check, using os.Stat(), that the file exists. If it does not exist, fail.
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("could not find firebase config file! Did you set your CFGPATH variable? %s", err)
	}
	log.Printf("using application credentials located at %s", configPath)

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	// acquire the firestore client, fail if we cannot.
	client, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	// naked return of "client" if successful.
	return
}
