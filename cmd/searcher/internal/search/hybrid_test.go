package search_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/web"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/internal/search"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestHybridSearch(t *testing.T) {
	// TODO maybe we should create a real git repo and then have FetchTar/etc
	// all work against it. That would make me feel more confident in
	// implementation.

	files := map[string]struct {
		body string
		typ  fileType
	}{
		"added.md": {`hello world I am added`, typeFile},

		"changed.go": {`package main

import "fmt"

func main() {
	fmt.Println("Hello world")
}
`, typeFile},

		"unchanged.md": {`# Hello World

Hello world example in go`, typeFile},
	}

	filesIndexed := map[string]struct {
		body string
		typ  fileType
	}{
		"changed.go": {`
This result should not appear even though it contains "world" since the file has changed.
`, typeFile},

		"removed.md": {`
This result should not appear even though it contains "world" since the file has been removed.
`, typeFile},

		"unchanged.md": {`# Hello World

Hello world example in go`, typeFile},
	}

	// We explicitly remove "unchanged.md" from files so the test has to rely
	// on the results from Zoekt.
	if unchanged := "unchanged.md"; files[unchanged] != filesIndexed[unchanged] {
		t.Fatal()
	} else {
		delete(files, unchanged)
	}

	gitDiffOutput := strings.Join([]string{
		"M", "changed.go",
		"A", "added.md",
		"D", "removed.md",
		"", // trailing null
	}, "\x00")

	pattern := protocol.PatternInfo{Pattern: "world"}
	wantRaw := `
added.md:1:1:
hello world I am added
changed.go:6:6:
	fmt.Println("Hello world")
unchanged.md:1:1:
# Hello World
unchanged.md:3:3:
Hello world example in go
`

	s := newStore(t, files)

	// explictly remove FetchTar since we should only be using FetchTarByPath
	s.FetchTar = nil

	// Ensure we don't ask for unchanged
	fetchTarPaths := s.FetchTarPaths
	s.FetchTarPaths = func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
		for _, p := range paths {
			if strings.Contains(p, "unchanged") {
				return nil, errors.Errorf("should not ask for unchanged path: %s", p)
			}
		}
		return fetchTarPaths(ctx, repo, commit, paths)
	}

	// we expect one command against git, lets just fake it.
	ts := httptest.NewServer(&search.Service{
		GitOutput: func(ctx context.Context, repo api.RepoName, args ...string) ([]byte, error) {
			want := []string{"diff", "-z", "--name-status", "--no-renames", "indexedfdeadbeefdeadbeefdeadbeefdeadbeef", "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}
			if d := cmp.Diff(want, args); d != "" {
				return nil, errors.Errorf("git diff mismatch (-want, +got):\n%s", d)
			}
			return []byte(gitDiffOutput), nil
		},
		Store: s,
		Log:   logtest.Scoped(t),
	})
	defer ts.Close()

	zoektURL := newZoekt(t, &zoekt.Repository{
		Name: "foo",
		ID:   123,
		Branches: []zoekt.RepositoryBranch{{
			Name:    "HEAD",
			Version: "indexedfdeadbeefdeadbeefdeadbeefdeadbeef",
		}},
	}, filesIndexed)

	req := protocol.Request{
		Repo:             "foo",
		RepoID:           123,
		URL:              "u",
		Commit:           "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		PatternInfo:      pattern,
		FetchTimeout:     fetchTimeoutForCI(t),
		IndexerEndpoints: []string{zoektURL},
		FeatHybrid:       true,
	}
	m, err := doSearch(ts.URL, &req)
	if err != nil {
		t.Fatal(err)
	}

	sort.Sort(sortByPath(m))
	got := strings.TrimSpace(toString(m))
	want := strings.TrimSpace(wantRaw)
	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("mismatch (-want, +got):\n%s", d)
	}
}

func newZoekt(t *testing.T, repo *zoekt.Repository, files map[string]struct {
	body string
	typ  fileType
}) string {
	var docs []zoekt.Document
	for name, file := range files {
		docs = append(docs, zoekt.Document{
			Name:     name,
			Content:  []byte(file.body),
			Branches: []string{"HEAD"},
		})
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	b, err := zoekt.NewIndexBuilder(repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if err := b.Add(d); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	if err := b.Write(&buf); err != nil {
		t.Fatal(err)
	}
	f := &memSeeker{data: buf.Bytes()}

	searcher, err := zoekt.NewSearcher(f)
	if err != nil {
		t.Fatal(err)
	}

	h, err := web.NewMux(&web.Server{
		Searcher: &streamer{Searcher: searcher},
		RPC:      true,
		Top:      web.Top,
	})
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	return ts.Listener.Addr().String()
}

type streamer struct {
	zoekt.Searcher
}

func (s *streamer) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, sender zoekt.Sender) (err error) {
	res, err := s.Searcher.Search(ctx, q, opts)
	if err != nil {
		return err
	}
	sender.Send(res)
	return nil
}

type memSeeker struct {
	data []byte
}

func (s *memSeeker) Name() string {
	return "memseeker"
}

func (s *memSeeker) Close() {}
func (s *memSeeker) Read(off, sz uint32) ([]byte, error) {
	return s.data[off : off+sz], nil
}

func (s *memSeeker) Size() (uint32, error) {
	return uint32(len(s.data)), nil
}
