package unpacktest

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func CreateZipArchive(t testing.TB, files map[string]string) io.ReadCloser {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	defer func() {
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for name, contents := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := io.Copy(w, strings.NewReader(contents)); err != nil {
			t.Fatal(err)
		}
	}

	return io.NopCloser(&b)
}

func CreateTarArchive(t testing.TB, files map[string]string) io.ReadCloser {
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	defer func() {
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for name, contents := range files {
		header := tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Size:     int64(len(contents)),
		}
		if err := tw.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}

		if _, err := io.Copy(tw, strings.NewReader(contents)); err != nil {
			t.Fatal(err)
		}
	}

	return io.NopCloser(&b)
}
