package httpext

import (
	"encoding/json"
	"io"
	"net/http"
)

// RequestBodyTo reads the request body with ioutil.ReadAll
// and marshals it into the interface described by i.
// As opposed to binding (see echo.Bind), BodyTo will return
// successfully in the event of partial filling. Empty bodies
// are also accepted.
// Returns error on failure.
// If body is empty, nil is returned, and i is untouched.
func RequestBodyTo(r *http.Request, i interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(i); err == io.EOF {
		return nil
	} else {
		return err
	}
}
