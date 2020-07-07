package httpext

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// RequestBodyTo reads the request body with ioutil.ReadAll
// and marshals it into the interface described by i.
// As opposed to binding (see echo.Bind), BodyTo will return
// successfully in the event of partial filling. Empty bodies
// are also accepted.
// Returns error on failure.
func RequestBodyTo(r *http.Request, i interface{}) error {
	if bytesBody, err := ioutil.ReadAll(r.Body); err != nil {
		return err
	} else if len(bytesBody) == 0 {
		return nil
	} else if err := json.Unmarshal(bytesBody, i); err != nil {
		return err
	}
	return nil
}
