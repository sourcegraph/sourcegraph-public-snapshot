package httputil

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

// WriteJSON writes a JSON Content-Type header and a JSON-encoded
// object to the http.ResponseWriter.
func WriteJSON(w http.ResponseWriter, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusInternalServerError, Err: err}
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")

	_, err = w.Write(data)
	return err
}
