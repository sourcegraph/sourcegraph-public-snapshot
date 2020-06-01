package paths

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMigrate(t *testing.T) {
	bundleDir := testRoot(t)

	files := []string{
		filepath.Join(bundleDir, "dbs", "123.lsif.db"),
		filepath.Join(bundleDir, "dbs", "123.456.lsif.db"),
		filepath.Join(bundleDir, "dbs", "123.457.lsif.db"),
		filepath.Join(bundleDir, "dbs", "unknown-file"),
		filepath.Join(bundleDir, "uploads", "123.lsif.gz"),
		filepath.Join(bundleDir, "uploads", "123.456.lsif.gz"),
		filepath.Join(bundleDir, "uploads", "123.457.lsif.gz"),
		filepath.Join(bundleDir, "uploads", "unknown-file"),
	}
	for _, file := range files {
		if err := makeFile(file); err != nil {
			t.Fatalf("unexpected error creating file: %s", err)
		}
	}

	if err := Migrate(bundleDir); err != nil {
		t.Fatalf("unexpected error running migrations: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	expected := []string{
		"db-parts/123.456.gz",
		"db-parts/123.457.gz",
		"dbs/123/sqlite.db",
		"dbs/unknown-file",
		"upload-parts/123.456.gz",
		"upload-parts/123.457.gz",
		"uploads/123.gz",
		"uploads/unknown-file",
	}
	if diff := cmp.Diff(expected, names); diff != "" {
		t.Errorf("unexpected directory contents (-want +got):\n%s", diff)
	}
}

func TestGetIDAndPartIndex(t *testing.T) {
	testCases := []struct {
		filename  string
		id        int
		partIndex int
	}{
		{"123.lsif.db", 123, -1},
		{"123.456.lsif.db", 123, 456},
		{"123.456.sqlite", -1, -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.filename, func(t *testing.T) {
			id, partIndex := getIDAndPartIndex(testCase.filename)
			if id != testCase.id {
				t.Errorf("unexpected id. want=%d have=%d", testCase.id, id)
			}
			if partIndex != testCase.partIndex {
				t.Errorf("unexpected part index. want=%d have=%d", testCase.partIndex, partIndex)
			}
		})
	}
}
