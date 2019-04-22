package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func TestServer_handleRepoInfo(t *testing.T) {
	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()
	_, ok := s.locker.TryAcquire("/testroot/a", "test status")
	if !ok {
		t.Fatal("could not acquire lock")
	}

	getRepoInfo := func(t *testing.T, repos ...api.RepoName) (resp protocol.RepoInfoResponse) {
		rr := httptest.NewRecorder()
		body, err := json.Marshal(protocol.RepoInfoRequest{Repos: repos})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/repos", bytes.NewReader(body))
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
		origRepoCloned := repoCloned
		repoCloned = func(dir string) bool { return false }
		defer func() { repoCloned = origRepoCloned }()

		want := protocol.RepoInfoResponse{
			Results: map[api.RepoName]*protocol.RepoInfo{
				"x": {},
			},
		}
		if got := getRepoInfo(t, "x"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("cloning", func(t *testing.T) {
		origRepoCloned := repoCloned
		repoCloned = func(dir string) bool { return false }
		defer func() { repoCloned = origRepoCloned }()

		want := protocol.RepoInfoResponse{
			Results: map[api.RepoName]*protocol.RepoInfo{
				"a": {
					CloneInProgress: true,
					CloneProgress:   "test status",
				},
			},
		}
		if got := getRepoInfo(t, "a"); !reflect.DeepEqual(got, want) {
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

		lastChanged := time.Date(1987, 1, 2, 3, 4, 5, 6, time.UTC)
		origRepoLastChanged := repoLastChanged
		repoLastChanged = func(dir string) (time.Time, error) { return lastChanged, nil }
		defer func() { repoLastChanged = origRepoLastChanged }()

		origRepoRemoteURL := repoRemoteURL
		repoRemoteURL = func(context.Context, string) (string, error) { return "u", nil }
		defer func() { repoRemoteURL = origRepoRemoteURL }()

		want := protocol.RepoInfoResponse{
			Results: map[api.RepoName]*protocol.RepoInfo{
				"x": {
					Cloned:      true,
					LastFetched: &lastFetched,
					LastChanged: &lastChanged,
					URL:         "u",
				},
			},
		}
		if got := getRepoInfo(t, "x"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("mutliple", func(t *testing.T) {
		origRepoCloned := repoCloned
		repoCloned = func(dir string) bool { return false }
		defer func() { repoCloned = origRepoCloned }()

		want := protocol.RepoInfoResponse{
			Results: map[api.RepoName]*protocol.RepoInfo{
				"a": {
					CloneInProgress: true,
					CloneProgress:   "test status",
				},
				"b": {}, // not cloned
			},
		}
		if got := getRepoInfo(t, "a", "b"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v want %+v", got, want)
		}
	})
}
