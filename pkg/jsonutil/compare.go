package jsonutil

import (
	"encoding/json"
	"testing"
)

func JSONEqual(t *testing.T, a, b interface{}) bool {
	return AsJSON(t, a) == AsJSON(t, b)
}

func AsJSON(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
