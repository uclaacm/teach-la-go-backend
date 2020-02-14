package tools

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// BodyTo reads the request body with ioutil.ReadAll
// and marshals it into the interface described by i. Returns
// an error on failure.
func BodyTo(r *http.Request, i interface{}) error {
	if bytesBody, err := ioutil.ReadAll(r.Body); err != nil {
		return err
	} else if err := json.Unmarshal(bytesBody, i); err != nil {
		return err
	}
	return nil
}
