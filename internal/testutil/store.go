package testutil

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

func NewStore(files map[string]string) (*store.Store, func(), error) {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := w.WriteHeader(hdr); err != nil {
			return nil, nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, nil, err
		}
	}
	// git-archive usually includes a pax header we should ignore.
	// use a body which matches a test case. Ensures we don't return this
	// false entry as a result.
	if err := addpaxheader(w, "Hello world\n"); err != nil {
		return nil, nil, err
	}

	err := w.Close()
	if err != nil {
		return nil, nil, err
	}
	d, err := ioutil.TempDir("", "search_test")
	if err != nil {
		return nil, nil, err
	}
	return &store.Store{
		FetchTar: func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
		},
		Path: d,
	}, func() { os.RemoveAll(d) }, nil
}

func addpaxheader(w *tar.Writer, body string) error {
	hdr := &tar.Header{
		Name:       "pax_global_header",
		Typeflag:   tar.TypeXGlobalHeader,
		PAXRecords: map[string]string{"somekey": body},
	}
	return w.WriteHeader(hdr)
}
