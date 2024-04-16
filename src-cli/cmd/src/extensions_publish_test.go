package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadExtensionIDFromManifest(t *testing.T) {
	tests := map[string]string{
		`{"name": "a", "publisher": "b"}`:                     "b/a",
		`{"name": "a", "publisher": "b", "extensionID": "c"}`: "c",
		`{"extensionID": "c"}`:                                "c",
	}
	for manifest, want := range tests {
		t.Run(manifest, func(t *testing.T) {
			got, err := readExtensionIDFromManifest([]byte(manifest))
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}

	t.Run("no name", func(t *testing.T) {
		if _, err := readExtensionIDFromManifest([]byte(`{}`)); err == nil {
			t.Fatal()
		}
	})

	t.Run("no publisher", func(t *testing.T) {
		if _, err := readExtensionIDFromManifest([]byte(`{"name":"a"}`)); err == nil {
			t.Fatal()
		}
	})
}

func TestUpdatePropertyInManifest(t *testing.T) {
	tests := map[string]string{
		`{}`:        `{"p": "x"}`,
		`{"a":1}`:   `{"a":1, "p": "x"}`,
		`{"p":"a"}`: `{"p": "x"}`,
	}
	for manifest, want := range tests {
		t.Run(manifest, func(t *testing.T) {
			got, err := updatePropertyInManifest([]byte(manifest), "p", "x")
			if err != nil {
				t.Fatal(err)
			}
			if !jsonDeepEqual(string(got), want) {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}

	t.Run("remove property", func(t *testing.T) {
		manifest := `{"extensionID":"sourcegraph/typescript", "url":"https://sourcegraph.com"}`
		want := `{"extensionID": "sourcegraph/typescript"}`
		got, err := updatePropertyInManifest([]byte(manifest), "url", "")
		if err != nil {
			t.Fatal(err)
		}
		if !jsonDeepEqual(string(got), want) {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func jsonDeepEqual(a, b string) bool {
	var va, vb interface{}
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		panic(err)
	}
	return reflect.DeepEqual(va, vb)
}
