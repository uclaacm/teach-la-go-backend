package tools

import (
	"../logger"
	"context"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// GetDB returns a pointer to a firestore *Client based on the
// JSON credentials pointed to by the environment variable
// $CFGPATH.
func GetDB(ctx *context.Context) (client *firestore.Client) {
	// load environment variable $CFGPATH as a string: the path to our config.
	configPath := os.Getenv("CFGPATH")

	// check, using os.Stat(), that the file exists. If it does not exist, fail.
	if _, err := os.Stat(configPath); err != nil {
		logger.Fatalf("could not find firebase config file! Did you set your CFGPATH variable? %s", err)
	}
	logger.Debugf("using application credentials located at %s", configPath)

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(*ctx, nil, opt)
	if err != nil {
		logger.Fatalf("failed to create client: %s", err)
	}

	// acquire the firestore client, fail if we cannot.
	client, err = app.Firestore(*ctx)
	if err != nil {
		logger.Fatalf("failed to create client: %s", err)
	}

	// naked return of client if successful.
	return
}
