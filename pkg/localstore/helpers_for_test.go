package localstore

import (
	"encoding/json"
	"testing"
)

func jsonEqual(t *testing.T, a, b interface{}) bool {
	return asJSON(t, a) == asJSON(t, b)
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
