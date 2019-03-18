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

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
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

	t.Run("args", func(t *testing.T) {
		called := false
		mockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			called = true
			if want := api.RepoURI("github.com/a/b"); args.Repo != want {
				t.Errorf("got owner %q, want %q", args.Repo, want)
			}
			return &protocol.RepoLookupResult{Repo: nil}, nil
		}
		defer func() { mockRepoLookup = nil }()

		repoLookupResult(t, "github.com/a/b")
		if !called {
			t.Error("!called")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{Repo: nil}, nil
		}
		defer func() { mockRepoLookup = nil }()

		if got, want := repoLookupResult(t, "github.com/a/b"), (protocol.RepoLookupResult{}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return nil, errors.New("x")
		}
		defer func() { mockRepoLookup = nil }()

		result, statusCode := repoLookup(t, "github.com/a/b")
		if result != nil {
			t.Errorf("got result %+v, want nil", result)
		}
		if want := http.StatusInternalServerError; statusCode != want {
			t.Errorf("got HTTP status code %d, want %d", statusCode, want)
		}
	})

	t.Run("found", func(t *testing.T) {
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
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &want, nil
		}
		defer func() { mockRepoLookup = nil }()
		if got := repoLookupResult(t, "github.com/c/d"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}

func TestRepoLookup(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		if _, err := repoLookup(context.Background(), protocol.RepoLookupArgs{}); err == nil {
			t.Error()
		}
	})

	t.Run("github", func(t *testing.T) {
		t.Run("not authoritative", func(t *testing.T) {
			orig := repos.GetGitHubRepositoryMock
			repos.GetGitHubRepositoryMock = func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
				return nil, false, errors.New("x")
			}
			defer func() { repos.GetGitHubRepositoryMock = orig }()

			result, err := repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "example.com/a/b"})
			if err != nil {
				t.Fatal(err)
			}
			if want := (&protocol.RepoLookupResult{ErrorNotFound: true}); !reflect.DeepEqual(result, want) {
				t.Errorf("got result %+v, want nil", result)
			}
		})

		t.Run("not found", func(t *testing.T) {
			orig := repos.GetGitHubRepositoryMock
			repos.GetGitHubRepositoryMock = func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
				return nil, true, github.ErrNotFound
			}
			defer func() { repos.GetGitHubRepositoryMock = orig }()

			result, err := repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/a/b"})
			if err != nil {
				t.Fatal(err)
			}
			if want := (&protocol.RepoLookupResult{ErrorNotFound: true}); !reflect.DeepEqual(result, want) {
				t.Errorf("got result %+v, want nil", result)
			}
		})

		t.Run("unexpected error", func(t *testing.T) {
			wantErr := errors.New("x")

			orig := repos.GetGitHubRepositoryMock
			repos.GetGitHubRepositoryMock = func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
				return nil, true, wantErr
			}
			defer func() { repos.GetGitHubRepositoryMock = orig }()

			result, err := repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/a/b"})
			if err != wantErr {
				t.Fatal(err)
			}
			if result != nil {
				t.Errorf("got result %+v, want nil", result)
			}
		})

		t.Run("found", func(t *testing.T) {
			want := &protocol.RepoLookupResult{
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

			orig := repos.GetGitHubRepositoryMock
			repos.GetGitHubRepositoryMock = func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
				return want.Repo, true, nil
			}
			defer func() { repos.GetGitHubRepositoryMock = orig }()

			result, err := repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/c/d"})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result, want) {
				t.Errorf("got %+v, want %+v", result, want)
			}
		})
	})
}
