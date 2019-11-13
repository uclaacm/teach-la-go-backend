package lib

import (
	"context"

	"cloud.google.com/go/firestore"
)

/* Handler
Used to manage all routes under a single firestore client.
Contains the running context for the firestore client, as well
as a pointer to a firestore client.
*/
type Handler struct {
	Context    context.Context
	FireClient *firestore.Client
}
