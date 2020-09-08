package store

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/zoekt/ignore"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestPrepareZip(t *testing.T) {
	s, cleanup := tmpStore(t)
	defer cleanup()

	wantRepo := gitserver.Repo{Name: "foo"}
	wantCommit := api.CommitID("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	returnFetch := make(chan struct{})
	var gotRepo gitserver.Repo
	var gotCommit api.CommitID
	var fetchZipCalled int64
	s.FetchTar = func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
		<-returnFetch
		atomic.AddInt64(&fetchZipCalled, 1)
		gotRepo = repo
		gotCommit = commit
		return emptyTar(t), nil
	}

	// Fetch same commit in parallel to ensure single-flighting works
	startPrepareZip := make(chan struct{})
	prepareZipErr := make(chan error)
	for i := 0; i < 10; i++ {
		go func() {
			<-startPrepareZip
			_, err := s.PrepareZip(context.Background(), wantRepo, wantCommit)
			prepareZipErr <- err
		}()
	}
	close(startPrepareZip)
	close(returnFetch)
	for i := 0; i < 10; i++ {
		err := <-prepareZipErr
		if err != nil {
			t.Fatal("expected PrepareZip to succeed:", err)
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
	_, err := s.PrepareZip(context.Background(), wantRepo, wantCommit)
	if err != nil {
		t.Fatal("expected PrepareZip to succeed:", err)
	}
}

func TestPrepareZip_fetchTarFail(t *testing.T) {
	fetchErr := errors.New("test")
	s, cleanup := tmpStore(t)
	defer cleanup()
	s.FetchTar = func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
		return nil, fetchErr
	}
	_, err := s.PrepareZip(context.Background(), gitserver.Repo{Name: "foo"}, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if errors.Cause(err) != fetchErr {
		t.Fatalf("expected PrepareZip to fail with %v, failed with %v", fetchErr, err)
	}
}

func TestPrepareZip_errHeader(t *testing.T) {
	s, cleanup := tmpStore(t)
	defer cleanup()
	s.FetchTar = func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
		buf := new(bytes.Buffer)
		w := tar.NewWriter(buf)
		w.Flush()
		buf.WriteString("oh yeah")
		err := w.Close()
		if err != nil {
			t.Fatal(err)
		}
		return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
	}
	_, err := s.PrepareZip(context.Background(), gitserver.Repo{Name: "foo"}, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if got, want := errors.Cause(err).Error(), tar.ErrHeader.Error(); got != want {
		t.Fatalf("expected PrepareZip to fail with tar.ErrHeader, failed with %v", got)
	}
	if !errors.Cause(err).(interface{ Temporary() bool }).Temporary() {
		t.Fatalf("expected PrepareZip to fail with a temporary error, failed with %v", err)
	}
}

func TestIngoreSizeMax(t *testing.T) {
	patterns := []string{
		"foo",
		"foo.*",
		"foo_*",
		"*.foo",
		"bar.baz",
	}
	tests := []struct {
		name    string
		ignored bool
	}{
		// Pass
		{"foo", true},
		{"foo.bar", true},
		{"foo_bar", true},
		{"bar.baz", true},
		{"bar.foo", true},
		// Fail
		{"baz.foo.bar", false},
		{"bar_baz", false},
		{"baz.baz", false},
	}

	for _, test := range tests {
		if got, want := ignoreSizeMax(test.name, patterns), test.ignored; got != want {
			t.Errorf("case %s got %v want %v", test.name, got, want)
		}
	}
}

func tmpStore(t *testing.T) (*Store, func()) {
	d, err := ioutil.TempDir("", "store_test")
	if err != nil {
		t.Fatal(err)
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
	}
	return ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
}

func writeArchive(w io.Writer, files map[string]string) (err error) {
	tw := tar.NewWriter(w)
	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}

func TestNewIgnoreMatcher(t *testing.T) {
	// create a tar with 4 files + ignore-file
	n := 4
	files := map[string]string{}
	for i := 0; i < n; i++ {
		s := fmt.Sprintf("%d", i)
		files["F"+s] = strings.Repeat("a", 10)
	}
	files[ignore.IgnoreFile] = "dir/"
	archive := bytes.Buffer{}
	err := writeArchive(&archive, files)
	if err != nil {
		t.Fatal(err)
	}

	// we use an io.TeeReader just like in the prod code to
	// make sure that newIgnoreMatcher really exhausts tee
	var buf bytes.Buffer
	tee := io.TeeReader(&archive, &buf)
	ig, err := newIgnoreMatcher(tar.NewReader(tee))
	if err != nil {
		t.Fatal(err)
	}

	if !ig.Match("dir/foo.txt") {
		t.Fatal("ig should haved matched dir/foo.txt")
	}

	// buf should contain the entire archive
	i := 0
	tr := tar.NewReader(&buf)
	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}
		i++
	}
	if i != n+1 { // +1 for the ignore-file
		t.Fatal("newIgnoreMatcher did no exhaust tee")
	}
}
