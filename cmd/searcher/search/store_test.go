package search

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpenReader(t *testing.T) {
	s, cleanup := tmpStore(t)
	defer cleanup()

	wantRepo := "foo"
	wantCommit := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	returnFetch := make(chan struct{})
	var gotRepo, gotCommit string
	var fetchZipCalled int64
	s.FetchTar = func(ctx context.Context, repo, commit string) (io.ReadCloser, error) {
		<-returnFetch
		atomic.AddInt64(&fetchZipCalled, 1)
		gotRepo = repo
		gotCommit = commit
		return emptyTar(t), nil
	}

	// Fetch same commit in parallel to ensure single-flighting works
	startOpenReader := make(chan struct{})
	openReaderErr := make(chan error)
	for i := 0; i < 10; i++ {
		go func() {
			<-startOpenReader
			ar, err := s.openReader(context.Background(), wantRepo, wantCommit)
			openReaderErr <- err
			if err == nil {
				ar.Close()
			}
		}()
	}
	close(startOpenReader)
	close(returnFetch)
	for i := 0; i < 10; i++ {
		err := <-openReaderErr
		if err != nil {
			t.Fatal("expected openReader to succeed:", err)
		}
	}

	if gotCommit != wantCommit {
		t.Errorf("fetched wrong commit. got=%v want=%v", gotCommit, wantCommit)
	}
	if gotRepo != wantRepo {
		t.Errorf("fetched wrong repo. got=%v want=%v", gotRepo, wantRepo)
	}

	// Wait for item to appear on disk cache, then test again to ensure we
	// use the disk cache.
	onDisk := false
	for i := 0; i < 500; i++ {
		files, _ := ioutil.ReadDir(s.Path)
		if len(files) != 0 {
			onDisk = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !onDisk {
		t.Fatal("timed out waiting for items to appear in cache at", s.Path)
	}
	ar, err := s.openReader(context.Background(), wantRepo, wantCommit)
	if err != nil {
		t.Fatal("expected openReader to succeed:", err)
		return
	}
	ar.Close()

	// The rest of this test is about how Store behaves when there are
	// existing archives in the cache (we use the ones placed here
	// previously).

	// Place some files that the store should delete on startup
	shouldDelete := []string{"unrelated", "corrupt.zip", "partial.zip.part"}
	for _, name := range shouldDelete {
		path := filepath.Join(s.Path, name)
		if err := ioutil.WriteFile(path, []byte("Hello World\n"), 0600); err != nil {
			t.Fatal("Failed to write bad cache item", name, err)
		}
	}

	var calledFetchTar bool
	s = &Store{
		Path: s.Path,
		FetchTar: func(ctx context.Context, repo, commit string) (io.ReadCloser, error) {
			calledFetchTar = true
			return nil, errors.New("should not be called")
		},
	}

	ar, err = s.openReader(context.Background(), wantRepo, wantCommit)
	if err != nil {
		t.Fatal(err)
	}
	ar.Close()

	if calledFetchTar {
		t.Fatal("Did not use on-disk cache")
	}

	for _, name := range shouldDelete {
		path := filepath.Join(s.Path, name)
		if _, err := os.Stat(path); err == nil {
			t.Fatal("did not delete bad cache item", name)
		}
	}
}

func TestOpenReader_fetchTarFail(t *testing.T) {
	fetchErr := errors.New("test")
	s, cleanup := tmpStore(t)
	defer cleanup()
	s.FetchTar = func(ctx context.Context, repo, commit string) (io.ReadCloser, error) {
		return nil, fetchErr
	}
	_, err := s.openReader(context.Background(), "foo", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != fetchErr {
		t.Fatalf("expected openReader to fail with %v, failed with %v", fetchErr, err)
	}
}

func tmpStore(t *testing.T) (*Store, func()) {
	d, err := ioutil.TempDir("", "search_test")
	if err != nil {
		t.Fatal(err)
		return nil, nil
	}
	return &Store{
		Path: d,
	}, func() { os.RemoveAll(d) }
}

func emptyTar(t *testing.T) io.ReadCloser {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	err := w.Close()
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
}
