package root

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSGHome(t *testing.T) {
	testHome := os.TempDir()
	actualHome, err := createSGHome(testHome)
	defer func() {
		os.Remove(actualHome)
	}()

	if err != nil {
		t.Fatalf("error creating SG Home dir(.sourcegraph) at %q: %q", testHome, err)
	}

	wantedHome := filepath.Join(testHome, ".sourcegraph")
	_, err = os.Stat(wantedHome)
	if err != nil {
		t.Errorf("failed to stat SG Home %q. Expected directory to be created\n", err)
	}
}
