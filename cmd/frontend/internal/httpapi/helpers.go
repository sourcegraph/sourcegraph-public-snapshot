package httpapi

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// writeJSON writes a JSON Content-Type header and a JSON-encoded object to the
// http.ResponseWriter.
func writeJSON(w http.ResponseWriter, v any) error {
	// Return "[]" instead of "null" if v is a nil slice.
	if reflect.TypeOf(v).Kind() == reflect.Slice && reflect.ValueOf(v).IsNil() {
		v = []any{}
	}

	// MarshalIndent takes about 30-50% longer, which
	// significantly increases the time it takes to handle and return
	// large HTTP API responses.
	w.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(v)
}
