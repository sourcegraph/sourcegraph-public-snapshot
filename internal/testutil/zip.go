package testutil

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

func CreateZip(files map[string]string) ([]byte, error) {
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

func MockZipFile(data []byte) (*store.ZipFile, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	zf := new(store.ZipFile)
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

func TempZipFromFiles(files map[string]string) (path string, cleanup func(), err error) {
	s, cleanup, err := NewStore(files)
	if err != nil {
		return "", cleanup, err
	}

	ctx := context.Background()
	repo := gitserver.Repo{Name: "foo", URL: "u"}
	var commit api.CommitID = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	path, err = s.PrepareZip(ctx, repo, commit)
	if err != nil {
		return "", cleanup, err
	}
	return path, cleanup, nil
}

func TempZipFileOnDisk(data []byte) (string, func(), error) {
	z, err := MockZipFile(data)
	if err != nil {
		return "", nil, err
	}
	d, err := ioutil.TempDir("", "temp_zip_dir")
	if err != nil {
		return "", nil, err
	}
	f, err := ioutil.TempFile(d, "temp_zip")
	if err != nil {
		return "", nil, err
	}
	_, err = f.Write(z.Data)
	if err != nil {
		return "", nil, err
	}
	return f.Name(), func() { os.RemoveAll(d) }, nil
}
