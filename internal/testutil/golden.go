package testutil

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertGolden(t testing.TB, path string, update bool, want interface{}) {
	t.Helper()

	data := marshal(t, want)

	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
		if err := ioutil.WriteFile(path, data, 0640); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
	}

	golden, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %s", path, err)
	}

	if diff := cmp.Diff(string(golden), string(data)); diff != "" {
		t.Error(diff)
	}
}

func marshal(t testing.TB, v interface{}) []byte {
	t.Helper()

	switch v2 := v.(type) {
	case string:
		return []byte(v2)
	case []byte:
		return v2
	default:
		data, err := json.MarshalIndent(v, " ", " ")
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
}
