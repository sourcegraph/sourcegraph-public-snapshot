package campaigns

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
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

	archive := mockRepoArchive{
		repo: repo,
		files: map[string]string{
			"README.md": "# Welcome to the README\n",
		},
	}

	t.Run("success", func(t *testing.T) {
		requestsReceived := 0
		callback := func(_ http.ResponseWriter, _ *http.Request) {
			requestsReceived++
		}

		ts := httptest.NewServer(newZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}

		rz, err := rf.Fetch(context.Background(), repo)
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

		if have, want := rz.Path(), filepath.Join(path.Clean(rf.dir), wantZipFile); want != have {
			t.Errorf("unexpected path: have=%q want=%q", have, want)
		}
		rz.Close()

		// Create it a second time and make sure that the server wasn't called
		rz, err = rf.Fetch(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		rz.Close()

		if requestsReceived != 1 {
			t.Fatalf("wrong number of requests received: %d", requestsReceived)
		}

		// Third time, but this time with cleanup, _after_ unzipping
		rf.deleteZips = true
		_, err = rf.Fetch(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if requestsReceived != 1 {
			t.Fatalf("wrong number of requests received: %d", requestsReceived)
		}
		rz.Close()

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

		ts := httptest.NewServer(newZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}

		if _, err := rf.Fetch(ctx, repo); err == nil {
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

		archive := mockRepoArchive{repo: repo, files: map[string]string{}}

		ts := httptest.NewServer(newZipArchivesMux(t, nil, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		rf := &repoFetcher{
			client:     client,
			dir:        workspaceTmpDir(t),
			deleteZips: false,
		}

		_, err := rf.Fetch(context.Background(), repo)
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
}
