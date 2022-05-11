package database

import (
	"encoding/json"
	"testing"
)

func assertJSONEqual(t *testing.T, want, got any) {
	wantJ := asJSON(t, want)
	gotJ := asJSON(t, got)
	if wantJ != gotJ {
		t.Errorf("Wanted %s, but got %s", wantJ, gotJ)
	}
}

func jsonEqual(t *testing.T, a, b any) bool {
	return asJSON(t, a) == asJSON(t, b)
}

func asJSON(t *testing.T, v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
