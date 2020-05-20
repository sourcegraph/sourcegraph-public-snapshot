package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

func TestStitchMultipart(t *testing.T) {
	bundleDir := testRoot(t)

	var expectedContents []byte
	for i := 0; i < 50; i++ {
		partContents := make([]byte, 20000)
		for j := 0; j < 20000; j++ {
			partContents[j] = byte(i)
		}
		expectedContents = append(expectedContents, partContents...)

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)

		if _, err := io.Copy(gzipWriter, bytes.NewReader(partContents)); err != nil {
			t.Fatalf("unexpected error writing to buffer: %s", err)
		}
		if err := gzipWriter.Close(); err != nil {
			t.Fatalf("unexpected error closing gzip writer: %s", err)
		}

		filename := paths.UploadPartFilename(bundleDir, 42, int64(i))
		if err := ioutil.WriteFile(filename, buf.Bytes(), os.ModePerm); err != nil {
			t.Fatalf("unexpected error writing file: %s", err)
		}
	}

	if err := stitchMultipart(bundleDir, 42); err != nil {
		t.Fatalf("unexpected error stitching multipart files: %s", err)
	}

	f, err := os.Open(paths.UploadFilename(bundleDir, 42))
	if err != nil {
		t.Fatalf("unexpected error opening file: %s", err)
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("unexpected error opening gzip reader: %s", err)
	}

	contents, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}

	if diff := cmp.Diff(expectedContents, contents); diff != "" {
		t.Errorf("unexpected file contents (-want +got):\n%s", diff)
	}

	for i := 0; i < 50; i++ {
		if _, err := os.Stat(paths.UploadPartFilename(bundleDir, 42, int64(i))); !os.IsNotExist(err) {
			t.Errorf("unexpected error. want=%q have=%q", os.ErrNotExist, err)
		}
	}
}

func testRoot(t *testing.T) string {
	bundleDir, err := ioutil.TempDir("", "precise-code-intel-bundle-manager-")
	if err != nil {
		t.Fatalf("unexpected error creating test directory: %s", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(bundleDir)
	})

	for _, dir := range []string{"", "uploads", "dbs"} {
		path := filepath.Join(bundleDir, dir)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			t.Fatalf("unexpected error creating test directory: %s", err)
		}
	}

	return bundleDir
}
