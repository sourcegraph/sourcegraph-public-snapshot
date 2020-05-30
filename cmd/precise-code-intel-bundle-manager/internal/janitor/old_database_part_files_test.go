package janitor

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

func TestRemoveOldDatabasePartFiles(t *testing.T) {
	bundleDir := testRoot(t)
	mtimes := map[string]time.Time{
		"42.0.lsif.gz": time.Now().Local().Add(-time.Minute * 3),  // older than 1m
		"42.1.lsif.gz": time.Now().Local().Add(-time.Minute * 2),  // older than 1m
		"43.0.lsif.gz": time.Now().Local().Add(-time.Second * 30), // newer than 1m
		"43.1.lsif.gz": time.Now().Local().Add(-time.Second * 20), // newer than 1m
		"50.lsif.gz":   time.Now().Local().Add(-time.Minute * 3),  // not a part file
		"51.lsif.gz":   time.Now().Local().Add(-time.Minute * 2),  // not a part file
	}

	for name, mtime := range mtimes {
		path := filepath.Join(bundleDir, "dbs", name)
		if err := makeFile(path, mtime); err != nil {
			t.Fatalf("unexpected error creating file %s: %s", path, err)
		}
	}

	j := &Janitor{
		bundleDir:          bundleDir,
		maxDatabasePartAge: time.Minute,
		metrics:            NewJanitorMetrics(metrics.TestRegisterer),
	}

	if err := j.removeOldDatabasePartFiles(); err != nil {
		t.Fatalf("unexpected error cleaning failed dbs: %s", err)
	}

	names, err := getFilenames(filepath.Join(bundleDir, "dbs"))
	if err != nil {
		t.Fatalf("unexpected error listing directory: %s", err)
	}

	expected := []string{"43.0.lsif.gz", "43.1.lsif.gz", "50.lsif.gz", "51.lsif.gz"}
	if diff := cmp.Diff(expected, names); diff != "" {
		t.Errorf("unexpected directory contents (-want +got):\n%s", diff)
	}
}
