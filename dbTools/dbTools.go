package dbTools

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

/* GetDB
Returns a pointer to a firestore Client based on the
JSON credentials pointed to by the environment variable
$CFGPATH.
*/
func GetDB() (*firestore.Client) {
	// get client config. Fails early
	// if we cannot find our config.
	configPath := os.Getenv("CFGPATH")
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("could not find firebase config file! %s", err)
	}
	log.Printf("using application credentials located at %s", configPath)

	// set up the app through which our client will be
	// acquired.
	opt := option.WithCredentialsFile(configPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	// acquire the firestore client.
	var client *firestore.Client
	client, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	return client
}
