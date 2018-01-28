package repoupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func TestServer_handleRepoLookup(t *testing.T) {
	s := &Server{}
	h := s.Handler()

	repoLookup := func(t *testing.T, repo api.RepoURI) (resp *protocol.RepoLookupResult, statusCode int) {
		t.Helper()
		rr := httptest.NewRecorder()
		body, err := json.Marshal(protocol.RepoLookupArgs{Repo: repo})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/repo-lookup", bytes.NewReader(body))
		h.ServeHTTP(rr, req)
		if rr.Code == http.StatusOK {
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatal(err)
			}
		}
		return resp, rr.Code
	}
	repoLookupResult := func(t *testing.T, repo api.RepoURI) protocol.RepoLookupResult {
		t.Helper()
		resp, statusCode := repoLookup(t, repo)
		if statusCode != http.StatusOK {
			t.Fatalf("http non-200 status %d", statusCode)
		}
		return *resp
	}

	t.Run("call args", func(t *testing.T) {
		called := false
		orig := github.GetRepositoryMock
		github.GetRepositoryMock = func(ctx context.Context, owner, name string) (*github.Repository, error) {
			called = true
			if want := "a"; owner != want {
				t.Errorf("got owner %q, want %q", owner, want)
			}
			if want := "b"; name != want {
				t.Errorf("got name %q, want %q", name, want)
			}
			return nil, github.ErrNotFound
		}
		defer func() { github.GetRepositoryMock = orig }()

		repoLookupResult(t, "github.com/a/b")
		if !called {
			t.Error("!called")
		}
	})

	t.Run("not found", func(t *testing.T) {
		orig := github.GetRepositoryMock
		github.GetRepositoryMock = func(ctx context.Context, owner, name string) (*github.Repository, error) {
			return nil, github.ErrNotFound
		}
		defer func() { github.GetRepositoryMock = orig }()

		if got, want := repoLookupResult(t, "github.com/a/b"), (protocol.RepoLookupResult{}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		orig := github.GetRepositoryMock
		github.GetRepositoryMock = func(ctx context.Context, owner, name string) (*github.Repository, error) {
			return nil, errors.New("x")
		}
		defer func() { github.GetRepositoryMock = orig }()

		result, statusCode := repoLookup(t, "github.com/a/b")
		if result != nil {
			t.Errorf("got result %+v, want nil", result)
		}
		if want := http.StatusInternalServerError; statusCode != want {
			t.Errorf("got HTTP status code %d, want %d", statusCode, want)
		}
	})

	t.Run("found", func(t *testing.T) {
		orig := github.GetRepositoryMock
		github.GetRepositoryMock = func(ctx context.Context, owner, name string) (*github.Repository, error) {
			return &github.Repository{
				ID:            "a",
				Description:   "b",
				NameWithOwner: "c/d",
				IsFork:        true,
			}, nil
		}
		defer func() { github.GetRepositoryMock = orig }()

		want := protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{
				ExternalRepo: &api.ExternalRepoSpec{
					ID:          "a",
					ServiceType: repos.GitHubServiceType,
					ServiceID:   "https://github.com/",
				},
				URI:         "github.com/c/d",
				Description: "b",
				Fork:        true,
			},
		}
		if got := repoLookupResult(t, "github.com/c/d"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}
