package config

import (
	"path/filepath"
	"runtime"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

func TestTree_validate(t *testing.T) {
	var absPath string
	if runtime.GOOS == "windows" {
		absPath = "C:\\foo"
	} else {
		absPath = "/foo"
	}

	tests := map[string]*Tree{
		"absolute path":                &Tree{SourceUnits: []*unit.SourceUnit{{Files: []string{absPath}}}},
		"relative path above root":     &Tree{SourceUnits: []*unit.SourceUnit{{Files: []string{filepath.Join("..", "foo")}}}},
		"bad path after being cleaned": &Tree{SourceUnits: []*unit.SourceUnit{{Files: []string{filepath.Join("foo", "bar", "..", "..", "..", "..", "baz")}}}},
	}

	for label, tree := range tests {
		if err := tree.validate(); err != ErrInvalidFilePath {
			t.Errorf("%s: got err %v, want ErrInvalidFilePath", label, err)
		}
	}
}
