package batches

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/mock"
)

func TestRepoFetcher_Fetch(t *testing.T) {
	workspaceTmpDir := func(t *testing.T) string {
		testTempDir, err := ioutil.TempDir("", "executor-integration-test-*")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(testTempDir) })

		return testTempDir
	}

	repo := &graphql.Repository{
		ID:            "src-cli",
		Name:          "github.com/sourcegraph/src-cli",
		DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	}

	archive := mock.RepoArchive{
		Repo: repo,
		Files: map[string]string{
			"README.md": "# Welcome to the README\n",
		},
	}

	t.Run("success", func(t *testing.T) {
		requestsReceived := 0
		callback := func(_ http.ResponseWriter, _ *http.Request) {
			requestsReceived++
		}

		ts := httptest.NewServer(mock.NewZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}

		zip := rf.Checkout(repo, "")
		err := zip.Fetch(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantZipFile := repo.Slug() + ".zip"
		ok, err := dirContains(rf.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}

		if have, want := zip.Path(), filepath.Join(path.Clean(rf.dir), wantZipFile); want != have {
			t.Errorf("unexpected path: have=%q want=%q", have, want)
		}
		zip.Close()

		// Create it a second time and make sure that the server wasn't called
		err = zip.Fetch(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		zip.Close()

		if requestsReceived != 1 {
			t.Fatalf("wrong number of requests received: %d", requestsReceived)
		}
	})

	t.Run("delete on close", func(t *testing.T) {
		ts := httptest.NewServer(mock.NewZipArchivesMux(t, nil, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: true,
		}

		zip := rf.Checkout(repo, "")

		err := zip.Fetch(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		wantZipFile := repo.Slug() + ".zip"
		ok, err := dirContains(rf.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}

		// Should be deleted after closing
		zip.Close()

		ok, err = dirContains(rf.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("temp dir contains zip file but should not")
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		// We create a context that is canceled after the server sent down the
		// first file to simulate a slow download that's aborted by the user hitting Ctrl-C.

		firstFileWritten := make(chan struct{})
		callback := func(w http.ResponseWriter, r *http.Request) {
			// We flush the headers and the first file
			w.(http.Flusher).Flush()

			// Wait a bit for the client to start writing the file
			time.Sleep(50 * time.Millisecond)

			// Cancel the context to simulate the Ctrl-C
			firstFileWritten <- struct{}{}

			<-r.Context().Done()
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-firstFileWritten
			cancel()
		}()

		ts := httptest.NewServer(mock.NewZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}

		zip := rf.Checkout(repo, "")
		if err := zip.Fetch(ctx); err == nil {
			t.Error("error is nil")
		}

		zipFile := repo.Slug() + ".zip"
		ok, err := dirContains(rf.dir, zipFile)
		if err != nil {
			t.Error(err)
		}
		if ok {
			t.Errorf("zip file in temp dir was not cleaned up")
		}
	})

	t.Run("non-default branch", func(t *testing.T) {
		otherBranchOID := "f00b4r"
		repo := &graphql.Repository{
			ID:            "src-cli-with-non-main-branch",
			Name:          "github.com/sourcegraph/src-cli",
			DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},

			Commit: graphql.Target{OID: otherBranchOID},
			Branch: graphql.Branch{Name: "other-branch", Target: graphql.Target{OID: otherBranchOID}},
		}

		archive := mock.RepoArchive{Repo: repo, Files: map[string]string{}}

		ts := httptest.NewServer(mock.NewZipArchivesMux(t, nil, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}
		zip := rf.Checkout(repo, "")

		err := zip.Fetch(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		wantZipFile := repo.Slug() + ".zip"
		ok, err := dirContains(rf.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}
	})

	t.Run("path in repository", func(t *testing.T) {
		additionalFiles := mock.MockRepoAdditionalFiles{
			Repo: repo,
			AdditionalFiles: map[string]string{
				".gitignore":     "node_modules",
				".gitattributes": "* -text",
				"a/.gitignore":   "node_modules-in-a",
			},
		}

		path := "a/b"
		archive := mock.RepoArchive{
			Repo: repo,
			Path: path,
			Files: map[string]string{
				"a/b/1.txt": "this is 1",
				"a/b/2.txt": "this is 1",
			},
		}

		var requestedArchivePath string
		callback := func(w http.ResponseWriter, r *http.Request) {
			s := strings.SplitN(r.URL.Path, "/raw/", 2)
			requestedArchivePath = s[1]
		}

		var requestedFiles []string
		middle := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				s := strings.SplitN(r.URL.Path, "/raw/", 2)
				requestedFiles = append(requestedFiles, s[1])

				next.ServeHTTP(w, r)
			})
		}

		mux := mock.NewZipArchivesMux(t, callback, archive)
		mock.HandleAdditionalFiles(mux, additionalFiles, middle)

		ts := httptest.NewServer(mux)
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}
		zip := rf.Checkout(repo, path)

		err := zip.Fetch(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !cmp.Equal(path, requestedArchivePath) {
			t.Errorf("wrong paths requested (-want +got):\n%s", cmp.Diff(path, requestedArchivePath))
		}

		wantRequestedFiles := []string{".gitignore", ".gitattributes", "a/.gitignore"}
		if !cmp.Equal(wantRequestedFiles, requestedFiles, cmpopts.SortSlices(sortStrings)) {
			t.Errorf("wrong paths requested (-want +got):\n%s", cmp.Diff(wantRequestedFiles, requestedFiles))
		}

		wantZipFile := repo.SlugForPath(path) + ".zip"
		ok, err := dirContains(rf.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}
	})
}

func sortStrings(a, b string) bool { return a < b }

func dirContains(dir, filename string) (bool, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.Name() == filename {
			return true, nil
		}
	}

	return false, nil
}
