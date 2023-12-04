package search

import (
	"archive/zip"
	"bytes"
	"os"
	"testing"
)

func createZip(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, body := range files {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mockZipFile(data []byte) (*zipFile, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	zf := new(zipFile)
	if err := zf.PopulateFiles(r); err != nil {
		return nil, err
	}
	// Make a copy of data to avoid accidental alias/re-use bugs.
	// This method is only for testing, so don't sweat the performance.
	zf.Data = make([]byte, len(data))
	copy(zf.Data, data)
	// zf.f is intentionally left nil;
	// this is an indicator that this is a mock ZipFile.
	return zf, nil
}

func tempZipFileOnDisk(t *testing.T, data []byte) string {
	t.Helper()
	z, err := mockZipFile(data)
	if err != nil {
		t.Fatal(err)
	}
	d := t.TempDir()
	f, err := os.CreateTemp(d, "temp_zip")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(z.Data)
	if err != nil {
		t.Fatal(err)
	}
	return f.Name()
}
