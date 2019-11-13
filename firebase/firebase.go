package firebase

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

type FirestoreConfig struct {
	CredentialType      string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url"`
}

/* GetDB
Returns a pointer to a firestore Client as well as
the context for said client. Requires that the environment
variable GOOGLE_APPLICATION_CREDENTIALS be properly set
and pointing to the desired *.json credentials file for
a service account for the application.
*/
func GetDB(config FirestoreConfig) (*firestore.Client, context.Context) {
	// setup ctx
	ctx := context.Background()

	// setup client
	client, err := firestore.NewClient(ctx, config.ProjectID)

	if err != nil {
		log.Fatal("failed to create client.")
		return nil, nil
	}

	log.Printf("created client for projectID: %s", config.ProjectID)
	return client, ctx
}
