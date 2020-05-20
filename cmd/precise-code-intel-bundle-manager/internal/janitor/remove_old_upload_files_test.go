package janitor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveOldUploadFiles(t *testing.T) {
	bundleDir := testRoot(t)
	mtimes := map[string]time.Time{
		"u1": time.Now().Local().Add(-time.Minute * 3),  // older than 1m
		"u2": time.Now().Local().Add(-time.Minute * 2),  // older than 1m
		"u3": time.Now().Local().Add(-time.Second * 30), // newer than 1m
		"u4": time.Now().Local().Add(-time.Second * 20), // newer than 1m
	}

	for name, mtime := range mtimes {
		path := filepath.Join(bundleDir, "uploads", name)
		if err := makeFile(path, mtime); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	j := &Janitor{
		bundleDir:    bundleDir,
		maxUploadAge: time.Minute,
		metrics:      NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOldUploadFiles(); err != nil {
		t.Fatalf("unexpected error cleaning failed uploads: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir, "uploads"))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	expected := []string{"u3", "u4"}
	if diff := cmp.Diff(expected, names); diff != "" {
		t.Errorf("unexpected directory contents (-want +got):\n%s", diff)
	}
}
