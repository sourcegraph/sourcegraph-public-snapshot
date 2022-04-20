package unpacktest

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func CreateZipArchive(t testing.TB, files map[string]io.Reader) (io.ReadCloser, error) {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	defer zw.Close()

	for name, f := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		_, err = io.Copy(w, f)
		if err != nil {
			t.Fatal(err)
		}
	}

	return io.NopCloser(&b), nil
}
