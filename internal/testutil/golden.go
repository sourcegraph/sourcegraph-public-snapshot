package testutil

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func AssertGolden(t testing.TB, path string, update bool, want interface{}) {
	t.Helper()

	data, err := json.MarshalIndent(want, " ", " ")
	if err != nil {
		t.Fatal(err)
	}

	if update {
		if err = ioutil.WriteFile(path, data, 0640); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
	}

	golden, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %s", path, err)
	}

	if have, want := string(data), string(golden); have != want {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(have, want, false)
		t.Error(dmp.DiffPrettyText(diffs))
	}
}
