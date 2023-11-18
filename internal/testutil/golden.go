package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func AssertGolden(t testing.TB, path string, update bool, want any) {
	t.Helper()

	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
		if err := os.WriteFile(path, marshal(t, want), 0o640); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
	}

	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %s", path, err)
	}

	compare(t, golden, want)
}

func compare(t testing.TB, got []byte, want any) {
	t.Helper()

	switch wantIs := want.(type) {
	case string:
		if diff := cmp.Diff(string(got), wantIs); diff != "" {
			t.Errorf("(-want, +got):\n%s", diff)
		}
	case []byte:
		if diff := cmp.Diff(string(got), string(wantIs)); diff != "" {
			t.Errorf("(-want, +got):\n%s", diff)
		}
	default:
		wantBytes, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		require.JSONEq(t, string(got), string(wantBytes))
	}
}

func marshal(t testing.TB, v any) []byte {
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
