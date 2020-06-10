package indexer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractTarfile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	tarfile, err := os.Open("./testdata/clone.tar")
	if err != nil {
		t.Fatalf("unexpected opening test tarfile: %s", err)
	}

	if err := extractTarfile(tempDir, tarfile); err != nil {
		t.Fatalf("unexpected extracting tarfile: %s", err)
	}

	sizes, err := readFiles(tempDir)
	if err != nil {
		t.Fatalf("unexpected reading directory: %s", err)
	}

	if diff := cmp.Diff(expectedCloneTarSizes, sizes); diff != "" {
		t.Errorf("unexpected commits (-want +got):\n%s", diff)
	}
}
