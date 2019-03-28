package repoupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func TestServer_handleExternalServiceSync(t *testing.T) {
	for _, tc := range []struct {
		name string
		svc  *api.ExternalService
		err  string
	}{
		{
			name: "bad kind",
			svc:  &api.ExternalService{},
			err:  "<nil>",
		},
		{
			name: "bad service config",
			svc: &api.ExternalService{
				DisplayName: "Other",
				Kind:        "OTHER",
				Config:      "{",
			},
			err: "external-service=0: config error: failed to parse JSON: [CloseBraceExpected]; \n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fa := repos.NewFakeInternalAPI([]*api.ExternalService{tc.svc}, nil)
			s := Server{OtherReposSyncer: repos.NewOtherReposSyncer(fa, nil)}
			ts := httptest.NewServer(s.Handler())
			defer ts.Close()

			cli := repoupdater.Client{URL: ts.URL, HTTPClient: http.DefaultClient}
			ctx := context.Background()

			_, err := cli.SyncExternalService(ctx, *tc.svc)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("\nhave: %s\nwant: %s", have, want)
			}
		})
	}
}

func TestServer_handleRepoLookup(t *testing.T) {
	s := &Server{
		OtherReposSyncer: repos.NewOtherReposSyncer(repos.NewFakeInternalAPI(nil, nil), nil),
		InternalAPI:      &internalAPIFake{},
	}
	h := s.Handler()

	repoLookup := func(t *testing.T, repo api.RepoName) (resp *protocol.RepoLookupResult, statusCode int) {
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
	repoLookupResult := func(t *testing.T, repo api.RepoName) protocol.RepoLookupResult {
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
			if want := api.RepoName("github.com/a/b"); args.Repo != want {
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
					ServiceType: github.ServiceType,
					ServiceID:   "https://github.com/",
				},
				Name:        "github.com/c/d",
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
	s := Server{
		Store:            new(repos.FakeStore),
		OtherReposSyncer: repos.NewOtherReposSyncer(repos.NewFakeInternalAPI(nil, nil), nil),
		InternalAPI:      &internalAPIFake{},
	}

	t.Run("no args", func(t *testing.T) {
		if _, err := s.repoLookup(context.Background(), protocol.RepoLookupArgs{}); err == nil {
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

			result, err := s.repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "example.com/a/b"})
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

			result, err := s.repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/a/b"})
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

			result, err := s.repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/a/b"})
			if err != wantErr {
				t.Fatal(err)
			}
			if result != nil {
				t.Errorf("got result %+v, want nil", result)
			}
		})
	})
}

func TestRepoLookup_found(t *testing.T) {
	fa := &internalAPIFake{
		metadataUpdate: make(chan *api.ReposUpdateMetadataRequest, 1),
	}
	s := Server{
		Store:            new(repos.FakeStore),
		OtherReposSyncer: repos.NewOtherReposSyncer(repos.NewFakeInternalAPI(nil, nil), nil),
		InternalAPI:      fa,
	}

	want := &protocol.RepoLookupResult{
		Repo: &protocol.RepoInfo{
			ExternalRepo: &api.ExternalRepoSpec{
				ID:          "a",
				ServiceType: github.ServiceType,
				ServiceID:   "https://github.com/",
			},
			Name:        "github.com/c/d",
			Description: "b",
			Fork:        true,
		},
	}

	orig := repos.GetGitHubRepositoryMock
	repos.GetGitHubRepositoryMock = func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
		return want.Repo, true, nil
	}
	defer func() { repos.GetGitHubRepositoryMock = orig }()

	result, err := s.repoLookup(context.Background(), protocol.RepoLookupArgs{Repo: "github.com/c/d"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("got %+v, want %+v", result, want)
	}

	select {
	case got := <-fa.metadataUpdate:
		want2 := &api.ReposUpdateMetadataRequest{
			RepoName:    want.Repo.Name,
			Description: want.Repo.Description,
			Fork:        want.Repo.Fork,
			Archived:    want.Repo.Archived,
		}
		if !reflect.DeepEqual(got, want2) {
			t.Errorf("got %+v, want %+v", got, want2)
		}
	case <-time.After(5 * time.Second):
		t.Error("ReposUpdateMetadata was not called")
	}
}

func TestRepoLookup_syncer(t *testing.T) {
	now := time.Now().UTC()
	ctx := context.Background()

	store := new(repos.FakeStore)
	_ = store.UpsertRepos(ctx, &repos.Repo{
		Name:        "github.com/foo/bar",
		Description: "The description",
		Language:    "barlang",
		Enabled:     true,
		Archived:    false,
		Fork:        false,
		CreatedAt:   now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: "github",
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*repos.SourceInfo{
			"extsvc:123": {
				ID:       "extsvc:123",
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
		Metadata: &github.Repository{
			ID:            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			URL:           "github.com/foo/bar",
			DatabaseID:    1234,
			Description:   "The description",
			NameWithOwner: "foo/bar",
		},
	})

	s := Server{
		OtherReposSyncer: repos.NewOtherReposSyncer(repos.NewFakeInternalAPI(nil, nil), nil),
		Syncer:           &repos.Syncer{},
		Store:            store,
		InternalAPI:      &internalAPIFake{},
	}

	t.Run("not found", func(t *testing.T) {
		result, err := s.repoLookup(ctx, protocol.RepoLookupArgs{Repo: "github.com/a/b"})
		if err != nil {
			t.Fatal(err)
		}
		if want := (&protocol.RepoLookupResult{ErrorNotFound: true}); !reflect.DeepEqual(result, want) {
			t.Errorf("got result %+v, want nil", result)
		}
	})

	t.Run("found", func(t *testing.T) {
		want := &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{
				ExternalRepo: &api.ExternalRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: github.ServiceType,
					ServiceID:   "https://github.com/",
				},
				Name:        "github.com/foo/bar",
				Description: "The description",
				VCS:         protocol.VCSInfo{URL: "git@github.com:foo/bar.git"},
				Links: &protocol.RepoLinks{
					Root:   "github.com/foo/bar",
					Tree:   "github.com/foo/bar/tree/{rev}/{path}",
					Blob:   "github.com/foo/bar/blob/{rev}/{path}",
					Commit: "github.com/foo/bar/commit/{commit}",
				},
			},
		}

		result, err := s.repoLookup(ctx, protocol.RepoLookupArgs{Repo: "github.com/foo/bar"})
		if err != nil {
			t.Fatal(err)
		}
		if diff := pretty.Compare(result, want); diff != "" {
			t.Errorf("ListRepos:\n%s", diff)
			return
		}
	})
}

type internalAPIFake struct {
	metadataUpdate chan *api.ReposUpdateMetadataRequest
}

func (a *internalAPIFake) ReposUpdateMetadata(ctx context.Context, repo api.RepoName, description string, fork, archived bool) error {
	if a.metadataUpdate != nil {
		a.metadataUpdate <- &api.ReposUpdateMetadataRequest{
			RepoName:    repo,
			Description: description,
			Fork:        fork,
			Archived:    archived,
		}
	}
	return nil
}

func init() {
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
}
