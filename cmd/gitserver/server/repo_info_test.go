package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func TestServer_handleRepoInfo(t *testing.T) {
	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()
	s.setCloneLock("/testroot/a")

	getRepoInfo := func(t *testing.T, repo string) (resp protocol.RepoInfoResponse) {
		rr := httptest.NewRecorder()
		body, err := json.Marshal(protocol.RepoInfoRequest{Repo: repo})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/repo", bytes.NewReader(body))
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("http non-200 status %d", rr.Code)
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		return resp
	}

	t.Run("not cloned", func(t *testing.T) {
		orig := repoCloned
		repoCloned = func(dir string) bool { return false }
		defer func() { repoCloned = orig }()

		if got, want := getRepoInfo(t, "x"), (protocol.RepoInfoResponse{}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("cloning", func(t *testing.T) {
		orig := repoCloned
		repoCloned = func(dir string) bool { return false }
		defer func() { repoCloned = orig }()

		if got, want := getRepoInfo(t, "a"), (protocol.RepoInfoResponse{CloneInProgress: true}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("cloned", func(t *testing.T) {
		origRepoCloned := repoCloned
		repoCloned = func(dir string) bool { return true }
		defer func() { repoCloned = origRepoCloned }()

		lastFetched := time.Date(1988, 1, 2, 3, 4, 5, 6, time.UTC)
		origRepoLastFetched := repoLastFetched
		repoLastFetched = func(dir string) (time.Time, error) { return lastFetched, nil }
		defer func() { repoLastFetched = origRepoLastFetched }()

		if got, want := getRepoInfo(t, "x"), (protocol.RepoInfoResponse{Cloned: true, LastFetched: &lastFetched}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}
