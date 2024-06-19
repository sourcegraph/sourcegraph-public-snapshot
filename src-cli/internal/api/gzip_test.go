package api

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGzipReader(t *testing.T) {
	var uncompressed []byte
	for i := 0; i < 20000; i++ {
		uncompressed = append(uncompressed, byte(i))
	}

	contents, err := io.ReadAll(gzipReader(bytes.NewReader(uncompressed)))
	if err != nil {
		t.Fatalf("unexpected error reading from gzip reader: %s", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(contents))
	if err != nil {
		t.Fatalf("unexpected error creating gzip.Reader: %s", err)
	}
	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		t.Fatalf("unexpected error reading from gzip.Reader: %s", err)
	}
	if diff := cmp.Diff(decompressed, uncompressed); diff != "" {
		t.Errorf("unexpected gzipped contents (-want +got):\n%s", diff)
	}
}
