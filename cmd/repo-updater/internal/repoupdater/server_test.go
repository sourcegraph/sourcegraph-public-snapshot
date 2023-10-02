package repoupdater

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repos/scheduler"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServer_handleRepoLookup(t *testing.T) {
	logger := logtest.Scoped(t)
	s := &Server{Logger: logger}

	h := ObservedHandler(
		logger,
		NewHandlerMetrics(),
		trace.NewNoopTracerProvider(),
	)(s.Handler())

	repoLookup := func(t *testing.T, repo api.RepoName) (resp *protocol.RepoLookupResult, statusCode int) {
		t.Helper()
		rr := httptest.NewRecorder()
		body, err := json.Marshal(protocol.RepoLookupArgs{Repo: repo})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/repo-lookup", bytes.NewReader(body))
		fmt.Printf("h: %v rr: %v req: %v\n", h, rr, req)
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
		mockRepoLookup = func(repoName api.RepoName) (*protocol.RepoLookupResult, error) {
			called = true
			if want := api.RepoName("github.com/a/b"); repoName != want {
				t.Errorf("got owner %q, want %q", repoName, want)
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
		mockRepoLookup = func(api.RepoName) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{Repo: nil}, nil
		}
		defer func() { mockRepoLookup = nil }()

		if got, want := repoLookupResult(t, "github.com/a/b"), (protocol.RepoLookupResult{}); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		mockRepoLookup = func(api.RepoName) (*protocol.RepoLookupResult, error) {
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
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "a",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Name:        "github.com/c/d",
				Description: "b",
				Fork:        true,
			},
		}
		mockRepoLookup = func(api.RepoName) (*protocol.RepoLookupResult, error) {
			return &want, nil
		}
		defer func() { mockRepoLookup = nil }()
		if got := repoLookupResult(t, "github.com/c/d"); !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}

func TestServer_EnqueueRepoUpdate(t *testing.T) {
	ctx := context.Background()

	svc := types.ExternalService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(`{
"url": "https://github.com",
"token": "secret-token",
"repos": ["owner/name"]
}`),
	}

	repo := types.Repo{
		ID:   1,
		Name: "github.com/foo/bar",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Metadata: new(github.Repository),
	}

	initStore := func(db database.DB) repos.Store {
		store := repos.NewStore(logtest.Scoped(t), db)
		if err := store.ExternalServiceStore().Upsert(ctx, &svc); err != nil {
			t.Fatal(err)
		}
		if err := store.RepoStore().Create(ctx, &repo); err != nil {
			t.Fatal(err)
		}
		return store
	}

	type testCase struct {
		name string
		repo api.RepoName
		res  *protocol.RepoUpdateResponse
		err  string
		init func(database.DB) repos.Store
	}

	testCases := []testCase{{
		name: "returns an error on store failure",
		init: func(realDB database.DB) repos.Store {
			mockRepos := dbmocks.NewMockRepoStore()
			mockRepos.ListFunc.SetDefaultReturn(nil, errors.New("boom"))
			realStore := initStore(realDB)
			mockStore := repos.NewMockStoreFrom(realStore)
			mockStore.RepoStoreFunc.SetDefaultReturn(mockRepos)
			return mockStore
		},
		err: `store.list-repos: boom`,
	}, {
		name: "missing repo",
		init: initStore,
		repo: "foo",
		err:  `repo foo not found with response: repo "foo" not found in store`,
	}, {
		name: "existing repo",
		repo: repo.Name,
		init: initStore,
		res: &protocol.RepoUpdateResponse{
			ID:   repo.ID,
			Name: string(repo.Name),
		},
	}}

	logger := logtest.Scoped(t)
	for _, tc := range testCases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			sqlDB := dbtest.NewDB(logger, t)
			store := tc.init(database.NewDB(logger, sqlDB))

			s := &Server{Logger: logger, Store: store, Scheduler: &fakeScheduler{}}
			gs := grpc.NewServer(defaults.ServerOptions(logger)...)
			proto.RegisterRepoUpdaterServiceServer(gs, &RepoUpdaterServiceServer{Server: s})

			srv := httptest.NewServer(internalgrpc.MultiplexHandlers(gs, s.Handler()))
			defer srv.Close()

			cli := repoupdater.NewClient(srv.URL)
			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.EnqueueRepoUpdate(ctx, tc.repo)
			if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
				t.Errorf("have err: %q, want: %q", have, want)
			}

			if have, want := res, tc.res; !reflect.DeepEqual(have, want) {
				t.Errorf("response: %s", cmp.Diff(have, want))
			}
		})
	}
}

func TestServer_RepoLookup(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := repos.NewStore(logger, database.NewDB(logger, db))
	ctx := context.Background()
	clock := timeutil.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	githubSource := types.ExternalService{
		Kind:         extsvc.KindGitHub,
		CloudDefault: true,
		Config: extsvc.NewUnencryptedConfig(`{
"url": "https://github.com",
"token": "secret-token",
"repos": ["owner/name"]
}`),
	}
	awsSource := types.ExternalService{
		Kind: extsvc.KindAWSCodeCommit,
		Config: extsvc.NewUnencryptedConfig(`
{
  "region": "us-east-1",
  "accessKeyID": "abc",
  "secretAccessKey": "abc",
  "gitCredentials": {
    "username": "user",
    "password": "pass"
  }
}
`),
	}
	gitlabSource := types.ExternalService{
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config: extsvc.NewUnencryptedConfig(`
{
  "url": "https://gitlab.com",
  "token": "abc",
  "projectQuery": ["none"]
}
`),
	}
	npmSource := types.ExternalService{
		Kind: extsvc.KindNpmPackages,
		Config: extsvc.NewUnencryptedConfig(`
{
  "registry": "npm.org"
}
`),
	}

	if err := store.ExternalServiceStore().Upsert(ctx, &githubSource, &awsSource, &gitlabSource, &npmSource); err != nil {
		t.Fatal(err)
	}

	githubRepository := &types.Repo{
		Name:        "github.com/foo/bar",
		Description: "The description",
		Archived:    false,
		Fork:        false,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			githubSource.URN(): {
				ID:       githubSource.URN(),
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
	}

	awsCodeCommitRepository := &types.Repo{
		Name:        "git-codecommit.us-west-1.amazonaws.com/stripe-go",
		Description: "The stripe-go lib",
		Archived:    false,
		Fork:        false,
		CreatedAt:   now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceType: extsvc.TypeAWSCodeCommit,
			ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
		},
		Sources: map[string]*types.SourceInfo{
			awsSource.URN(): {
				ID:       awsSource.URN(),
				CloneURL: "git@git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go",
			},
		},
		Metadata: &awscodecommit.Repository{
			ARN:          "arn:aws:codecommit:us-west-1:999999999999:stripe-go",
			AccountID:    "999999999999",
			ID:           "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
			Name:         "stripe-go",
			Description:  "The stripe-go lib",
			HTTPCloneURL: "https://git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go",
			LastModified: &now,
		},
	}

	gitlabRepository := &types.Repo{
		Name:        "gitlab.com/gitlab-org/gitaly",
		Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
		URI:         "gitlab.com/gitlab-org/gitaly",
		CreatedAt:   now,
		UpdatedAt:   now,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
		Sources: map[string]*types.SourceInfo{
			gitlabSource.URN(): {
				ID:       gitlabSource.URN(),
				CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
			},
		},
		Metadata: &gitlab.Project{
			ProjectCommon: gitlab.ProjectCommon{
				ID:                2009901,
				PathWithNamespace: "gitlab-org/gitaly",
				Description:       "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
				WebURL:            "https://gitlab.com/gitlab-org/gitaly",
				HTTPURLToRepo:     "https://gitlab.com/gitlab-org/gitaly.git",
				SSHURLToRepo:      "git@gitlab.com:gitlab-org/gitaly.git",
			},
			Visibility: "",
			Archived:   false,
		},
	}

	npmRepository := &types.Repo{
		Name: "npm/package",
		URI:  "npm/package",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "npm/package",
			ServiceType: extsvc.TypeNpmPackages,
			ServiceID:   extsvc.TypeNpmPackages,
		},
		Sources: map[string]*types.SourceInfo{
			npmSource.URN(): {
				ID:       npmSource.URN(),
				CloneURL: "npm/package",
			},
		},
		Metadata: &reposource.NpmMetadata{Package: func() *reposource.NpmPackageName {
			p, _ := reposource.NewNpmPackageName("", "package")
			return p
		}()},
	}

	testCases := []struct {
		name        string
		args        protocol.RepoLookupArgs
		stored      types.Repos
		result      *protocol.RepoLookupResult
		src         repos.Source
		assert      typestest.ReposAssertion
		assertDelay time.Duration
		err         string
	}{
		{
			name: "found - aws code commit",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("git-codecommit.us-west-1.amazonaws.com/stripe-go"),
			},
			stored: []*types.Repo{awsCodeCommitRepository},
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "f001337a-3450-46fd-b7d2-650c0EXAMPLE",
					ServiceType: extsvc.TypeAWSCodeCommit,
					ServiceID:   "arn:aws:codecommit:us-west-1:999999999999:",
				},
				Name:        "git-codecommit.us-west-1.amazonaws.com/stripe-go",
				Description: "The stripe-go lib",
				VCS:         protocol.VCSInfo{URL: "git@git-codecommit.us-west-1.amazonaws.com/v1/repos/stripe-go"},
				Links: &protocol.RepoLinks{
					Root:   "https://us-west-1.console.aws.amazon.com/codesuite/codecommit/repositories/stripe-go/browse",
					Tree:   "https://us-west-1.console.aws.amazon.com/codesuite/codecommit/repositories/stripe-go/browse/{rev}/--/{path}",
					Blob:   "https://us-west-1.console.aws.amazon.com/codesuite/codecommit/repositories/stripe-go/browse/{rev}/--/{path}",
					Commit: "https://us-west-1.console.aws.amazon.com/codesuite/codecommit/repositories/stripe-go/commit/{commit}",
				},
			}},
		},
		{
			name: "not synced from non public codehost",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.private.corp/a/b"),
			},
			src:    repos.NewFakeSource(&githubSource, nil),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (name=%s notfound=%v)", api.RepoName("github.private.corp/a/b"), true),
		},
		{
			name: "synced - npm package host",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("npm/package"),
			},
			stored: []*types.Repo{},
			src:    repos.NewFakeSource(&npmSource, nil, npmRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternalRepo: npmRepository.ExternalRepo,
				Name:         npmRepository.Name,
				VCS:          protocol.VCSInfo{URL: string(npmRepository.Name)},
			}},
			assert: typestest.Assert.ReposEqual(npmRepository),
		},
		{
			name: "synced - github.com cloud default",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.com/foo/bar"),
			},
			stored: []*types.Repo{},
			src:    repos.NewFakeSource(&githubSource, nil, githubRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
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
			}},
			assert: typestest.Assert.ReposEqual(githubRepository),
		},
		{
			name: "found - github.com already exists",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.com/foo/bar"),
			},
			stored: []*types.Repo{githubRepository},
			src:    repos.NewFakeSource(&githubSource, nil, githubRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
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
			}},
		},
		{
			name: "not found - github.com",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.com/foo/bar"),
			},
			src:    repos.NewFakeSource(&githubSource, github.ErrRepoNotFound),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (name=%s notfound=%v)", api.RepoName("github.com/foo/bar"), true),
			assert: typestest.Assert.ReposEqual(),
		},
		{
			name: "unauthorized - github.com",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.com/foo/bar"),
			},
			src:    repos.NewFakeSource(&githubSource, &github.APIError{Code: http.StatusUnauthorized}),
			result: &protocol.RepoLookupResult{ErrorUnauthorized: true},
			err:    fmt.Sprintf("not authorized (name=%s noauthz=%v)", api.RepoName("github.com/foo/bar"), true),
			assert: typestest.Assert.ReposEqual(),
		},
		{
			name: "temporarily unavailable - github.com",
			args: protocol.RepoLookupArgs{
				Repo: api.RepoName("github.com/foo/bar"),
			},
			src:    repos.NewFakeSource(&githubSource, &github.APIError{Message: "API rate limit exceeded"}),
			result: &protocol.RepoLookupResult{ErrorTemporarilyUnavailable: true},
			err: fmt.Sprintf(
				"repository temporarily unavailable (name=%s istemporary=%v)",
				api.RepoName("github.com/foo/bar"),
				true,
			),
			assert: typestest.Assert.ReposEqual(),
		},
		{
			name:   "synced - gitlab.com",
			args:   protocol.RepoLookupArgs{Repo: gitlabRepository.Name},
			stored: []*types.Repo{},
			src:    repos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				Name:        "gitlab.com/gitlab-org/gitaly",
				Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
				Fork:        false,
				Archived:    false,
				VCS: protocol.VCSInfo{
					URL: "https://gitlab.com/gitlab-org/gitaly.git",
				},
				Links: &protocol.RepoLinks{
					Root:   "https://gitlab.com/gitlab-org/gitaly",
					Tree:   "https://gitlab.com/gitlab-org/gitaly/tree/{rev}/{path}",
					Blob:   "https://gitlab.com/gitlab-org/gitaly/blob/{rev}/{path}",
					Commit: "https://gitlab.com/gitlab-org/gitaly/commit/{commit}",
				},
				ExternalRepo: gitlabRepository.ExternalRepo,
			}},
			assert: typestest.Assert.ReposEqual(gitlabRepository),
		},
		{
			name:   "found - gitlab.com",
			args:   protocol.RepoLookupArgs{Repo: gitlabRepository.Name},
			stored: []*types.Repo{gitlabRepository},
			src:    repos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				Name:        "gitlab.com/gitlab-org/gitaly",
				Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
				Fork:        false,
				Archived:    false,
				VCS: protocol.VCSInfo{
					URL: "https://gitlab.com/gitlab-org/gitaly.git",
				},
				Links: &protocol.RepoLinks{
					Root:   "https://gitlab.com/gitlab-org/gitaly",
					Tree:   "https://gitlab.com/gitlab-org/gitaly/tree/{rev}/{path}",
					Blob:   "https://gitlab.com/gitlab-org/gitaly/blob/{rev}/{path}",
					Commit: "https://gitlab.com/gitlab-org/gitaly/commit/{commit}",
				},
				ExternalRepo: gitlabRepository.ExternalRepo,
			}},
		},
		{
			name: "Private repos are not supported on sourcegraph.com",
			args: protocol.RepoLookupArgs{
				Repo: githubRepository.Name,
			},
			src: repos.NewFakeSource(&githubSource, nil, githubRepository.With(func(r *types.Repo) {
				r.Private = true
			})),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (name=%s notfound=%v)", githubRepository.Name, true),
		},
		{
			name: "Private repos that used to be public should be removed asynchronously",
			args: protocol.RepoLookupArgs{
				Repo: githubRepository.Name,
			},
			src: repos.NewFakeSource(&githubSource, github.ErrRepoNotFound),
			stored: []*types.Repo{githubRepository.With(func(r *types.Repo) {
				r.UpdatedAt = r.UpdatedAt.Add(-time.Hour)
			})},
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
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
			}},
			assertDelay: time.Second,
			assert:      typestest.Assert.ReposEqual(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			_, err := db.ExecContext(ctx, "DELETE FROM repo")
			if err != nil {
				t.Fatal(err)
			}

			rs := tc.stored.Clone()
			err = store.RepoStore().Create(ctx, rs...)
			if err != nil {
				t.Fatal(err)
			}

			clock := clock
			logger := logtest.Scoped(t)
			syncer := &repos.Syncer{
				Now:     clock.Now,
				Store:   store,
				Sourcer: repos.NewFakeSourcer(nil, tc.src),
				ObsvCtx: observation.TestContextTB(t),
			}

			scheduler := scheduler.NewUpdateScheduler(logtest.Scoped(t), dbmocks.NewMockDB(), gitserver.NewMockClient())

			s := &Server{
				Logger:    logger,
				Syncer:    syncer,
				Store:     store,
				Scheduler: scheduler,
			}

			gs := grpc.NewServer(defaults.ServerOptions(logger)...)
			proto.RegisterRepoUpdaterServiceServer(gs, &RepoUpdaterServiceServer{Server: s})

			srv := httptest.NewServer(internalgrpc.MultiplexHandlers(gs, s.Handler()))
			defer srv.Close()

			cli := repoupdater.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.RepoLookup(ctx, tc.args)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("have err: %q, want: %q", have, want)
			}

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fatalf("response mismatch(-have, +want): %s", diff)
			}

			if tc.assert != nil {
				if tc.assertDelay != 0 {
					time.Sleep(tc.assertDelay)
				}
				rs, err := store.RepoStore().List(ctx, database.ReposListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				tc.assert(t, rs)
			}
		})
	}
}

type fakeScheduler struct{}

func (s *fakeScheduler) UpdateOnce(_ api.RepoID, _ api.RepoName) {}
func (s *fakeScheduler) ScheduleInfo(_ api.RepoID) *protocol.RepoUpdateSchedulerInfoResult {
	return &protocol.RepoUpdateSchedulerInfoResult{}
}

func TestServer_handleExternalServiceValidate(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantErrCode int
	}{
		{
			name:        "unauthorized",
			err:         &repoupdater.ErrUnauthorized{NoAuthz: true},
			wantErrCode: 401,
		},
		{
			name:        "forbidden",
			err:         repos.ErrForbidden{},
			wantErrCode: 403,
		},
		{
			name:        "other",
			err:         errors.Errorf("Any error"),
			wantErrCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			src := testSource{
				fn: func() error {
					return test.err
				},
			}

			es := &types.ExternalService{ID: 1, Kind: extsvc.KindGitHub, Config: extsvc.NewEmptyConfig()}
			statusCode, _ := handleExternalServiceValidate(context.Background(), logtest.Scoped(t), es, src)
			if statusCode != test.wantErrCode {
				t.Errorf("Code: want %v but got %v", test.wantErrCode, statusCode)
			}
		})
	}
}

func TestExternalServiceValidate_ValidatesToken(t *testing.T) {
	var (
		src    repos.Source
		called bool
		ctx    = context.Background()
	)
	src = testSource{
		fn: func() error {
			called = true
			return nil
		},
	}
	err := externalServiceValidate(ctx, &types.ExternalService{}, src)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if !called {
		t.Errorf("expected called, got not called")
	}
}

func TestServer_ExternalServiceNamespaces(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternalService{
		Kind:         extsvc.KindGitHub,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlabConnection := `
	{
	   "url": "https://gitlab.com",
	   "token": "abc",
	}`

	gitlabSource := types.ExternalService{
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(gitlabConnection),
	}

	githubOrg := &types.ExternalServiceNamespace{
		ID:         1,
		Name:       "sourcegraph",
		ExternalID: "aaaaa",
	}

	githubExternalServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternalService := types.ExternalService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternalServiceConfig),
	}

	gitlabExternalServiceConfig := `
	{
		"url": "https://gitlab.com",
		"token": "abc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlabExternalService := types.ExternalService{
		ID:           2,
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(gitlabExternalServiceConfig),
	}

	gitlabRepository := &types.Repo{
		Name:        "gitlab.com/gitlab-org/gitaly",
		Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
		URI:         "gitlab.com/gitlab-org/gitaly",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
		Sources: map[string]*types.SourceInfo{
			gitlabSource.URN(): {
				ID:       gitlabSource.URN(),
				CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
			},
		},
	}

	var idDoesNotExist int64 = 99

	testCases := []struct {
		name              string
		externalService   *types.ExternalService
		externalServiceID *int64
		kind              string
		config            string
		result            *protocol.ExternalServiceNamespacesResult
		src               repos.Source
		err               string
	}{
		{
			name:   "discoverable source - github",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			result: &protocol.ExternalServiceNamespacesResult{Namespaces: []*types.ExternalServiceNamespace{githubOrg}, Error: ""},
		},
		{
			name:   "unavailable - github.com",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, &types.Repo{}), true, githubOrg),
			result: &protocol.ExternalServiceNamespacesResult{Error: "fake source unavailable"},
			err:    "fake source unavailable",
		},
		{
			name:   "discoverable source - github - empty namespaces result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, &types.Repo{}), false),
			result: &protocol.ExternalServiceNamespacesResult{Namespaces: []*types.ExternalServiceNamespace{}, Error: ""},
		},
		{
			name:   "source does not implement discoverable source",
			kind:   extsvc.KindGitLab,
			config: gitlabConnection,
			src:    repos.NewFakeSource(&gitlabSource, nil, &types.Repo{}),
			result: &protocol.ExternalServiceNamespacesResult{Error: repos.UnimplementedDiscoverySource},
			err:    repos.UnimplementedDiscoverySource,
		},
		{
			name:              "discoverable source - github - use existing external service",
			externalService:   &githubExternalService,
			externalServiceID: &githubExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			result:            &protocol.ExternalServiceNamespacesResult{Namespaces: []*types.ExternalServiceNamespace{githubOrg}, Error: ""},
		},
		{
			name:              "external service for ID does not exist and other config parameters are not attempted",
			externalService:   &githubExternalService,
			externalServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, &types.Repo{}), false, githubOrg),
			result:            &protocol.ExternalServiceNamespacesResult{Error: fmt.Sprintf("external service not found: %d", idDoesNotExist)},
			err:               fmt.Sprintf("external service not found: %d", idDoesNotExist),
		},
		{
			name:              "source does not implement discoverable source - use existing external service",
			externalService:   &gitlabExternalService,
			externalServiceID: &gitlabExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			src:               repos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			result:            &protocol.ExternalServiceNamespacesResult{Error: repos.UnimplementedDiscoverySource},
			err:               repos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			logger := logtest.Scoped(t)
			var (
				sqlDB *sql.DB
				store repos.Store
			)

			if tc.externalService != nil {
				sqlDB = dbtest.NewDB(logger, t)
				store = repos.NewStore(logtest.Scoped(t), database.NewDB(logger, sqlDB))
				if err := store.ExternalServiceStore().Upsert(ctx, tc.externalService); err != nil {
					t.Fatal(err)
				}
			}

			s := &Server{
				Store:  store,
				Logger: logger,
			}

			mockNewGenericSourcer = func() repos.Sourcer {
				return repos.NewFakeSourcer(nil, tc.src)
			}
			t.Cleanup(func() { mockNewGenericSourcer = nil })

			grpcServer := defaults.NewServer(logger)
			proto.RegisterRepoUpdaterServiceServer(grpcServer, &RepoUpdaterServiceServer{Server: s})
			handler := internalgrpc.MultiplexHandlers(grpcServer, s.Handler())

			srv := httptest.NewServer(handler)
			defer srv.Close()

			cli := repoupdater.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			args := protocol.ExternalServiceNamespacesArgs{
				ExternalServiceID: tc.externalServiceID,
				Kind:              tc.kind,
				Config:            tc.config,
			}

			res, err := cli.ExternalServiceNamespaces(ctx, args)
			if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}
			if err != nil {
				return
			}

			if have, want := res.Error, tc.result.Error; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fatalf("response mismatch(-have, +want): %s", diff)
			}
		})
	}
}

func TestServer_ExternalServiceRepositories(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternalService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlabConnection := `
	{
	   "url": "https://gitlab.com",
	   "token": "abc",
	}`

	gitlabSource := types.ExternalService{
		ID:           2,
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(gitlabConnection),
	}

	githubRepository := &types.Repo{
		Name:        "github.com/foo/bar",
		Description: "The description",
		Archived:    false,
		Fork:        false,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			githubSource.URN(): {
				ID:       githubSource.URN(),
				CloneURL: "git@github.com:foo/bar.git",
			},
		},
	}

	gitlabRepository := &types.Repo{
		Name:        "gitlab.com/gitlab-org/gitaly",
		Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
		URI:         "gitlab.com/gitlab-org/gitaly",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
		Sources: map[string]*types.SourceInfo{
			gitlabSource.URN(): {
				ID:       gitlabSource.URN(),
				CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
			},
		},
	}

	githubExternalServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternalService := types.ExternalService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternalServiceConfig),
	}

	gitlabExternalServiceConfig := `
	{
		"url": "https://gitlab.com",
		"token": "abc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlabExternalService := types.ExternalService{
		ID:           2,
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config:       extsvc.NewUnencryptedConfig(gitlabExternalServiceConfig),
	}

	var idDoesNotExist int64 = 99

	testCases := []struct {
		name              string
		externalService   *types.ExternalService
		externalServiceID *int64
		kind              string
		config            string
		query             string
		first             int32
		excludeRepos      []string
		result            *protocol.ExternalServiceRepositoriesResult
		src               repos.Source
		err               string
	}{
		{
			name:         "discoverable source - github",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{},
			src:          repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       &protocol.ExternalServiceRepositoriesResult{Repos: []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()}, Error: ""},
		},
		{
			name:         "discoverable source - github - non empty query string",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "myquerystring",
			first:        5,
			excludeRepos: []string{},
			src:          repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       &protocol.ExternalServiceRepositoriesResult{Repos: []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()}, Error: ""},
		},
		{
			name:         "discoverable source - github - non empty excludeRepos",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{"org1/repo1", "owner2/repo2"},
			src:          repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, githubRepository), false),
			result:       &protocol.ExternalServiceRepositoriesResult{Repos: []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()}, Error: ""},
		},
		{
			name:   "unavailable - github.com",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil, githubRepository), true),
			result: &protocol.ExternalServiceRepositoriesResult{Error: "fake source unavailable"},
			err:    "fake source unavailable",
		},
		{
			name:   "discoverable source - github - empty repositories result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubSource, nil), false),
			result: &protocol.ExternalServiceRepositoriesResult{Repos: []*types.ExternalServiceRepository{}, Error: ""},
		},
		{
			name:   "source does not implement discoverable source",
			kind:   extsvc.KindGitLab,
			config: gitlabConnection,
			src:    repos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			result: &protocol.ExternalServiceRepositoriesResult{Error: repos.UnimplementedDiscoverySource},
			err:    repos.UnimplementedDiscoverySource,
		},
		{
			name:              "discoverable source - github - use existing external service",
			externalService:   &githubExternalService,
			externalServiceID: &githubExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubExternalService, nil, githubRepository), false),
			result:            &protocol.ExternalServiceRepositoriesResult{Repos: []*types.ExternalServiceRepository{githubRepository.ToExternalServiceRepository()}, Error: ""},
		},
		{
			name:              "external service for ID does not exist and other config parameters are not attempted",
			externalService:   &githubExternalService,
			externalServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubExternalServiceConfig,
			query:             "myquerystring",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFakeDiscoverableSource(repos.NewFakeSource(&githubExternalService, nil, githubRepository), false),
			result:            &protocol.ExternalServiceRepositoriesResult{Error: fmt.Sprintf("external service not found: %d", idDoesNotExist)},
			err:               fmt.Sprintf("external service not found: %d", idDoesNotExist),
		},
		{
			name:              "source does not implement discoverable source - use existing external service",
			externalService:   &gitlabExternalService,
			externalServiceID: &gitlabExternalService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFakeSource(&gitlabSource, nil, gitlabRepository),
			result:            &protocol.ExternalServiceRepositoriesResult{Error: repos.UnimplementedDiscoverySource},
			err:               repos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			logger := logtest.Scoped(t)
			var (
				sqlDB *sql.DB
				store repos.Store
			)

			if tc.externalService != nil {
				sqlDB = dbtest.NewDB(logger, t)
				store = repos.NewStore(logtest.Scoped(t), database.NewDB(logger, sqlDB))
				if err := store.ExternalServiceStore().Upsert(ctx, tc.externalService); err != nil {
					t.Fatal(err)
				}
			}

			s := &Server{
				Store:  store,
				Logger: logger,
			}

			mockNewGenericSourcer = func() repos.Sourcer {
				return repos.NewFakeSourcer(nil, tc.src)
			}
			t.Cleanup(func() { mockNewGenericSourcer = nil })

			grpcServer := defaults.NewServer(logger)
			proto.RegisterRepoUpdaterServiceServer(grpcServer, &RepoUpdaterServiceServer{Server: s})
			handler := internalgrpc.MultiplexHandlers(grpcServer, s.Handler())

			srv := httptest.NewServer(handler)
			defer srv.Close()

			cli := repoupdater.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			args := protocol.ExternalServiceRepositoriesArgs{
				ExternalServiceID: tc.externalServiceID,
				Kind:              tc.kind,
				Config:            tc.config,
				Query:             tc.query,
				First:             tc.first,
				ExcludeRepos:      tc.excludeRepos,
			}

			res, err := cli.ExternalServiceRepositories(ctx, args)
			if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}
			if err != nil {
				return
			}

			if have, want := res.Error, tc.result.Error; !strings.Contains(have, want) {
				t.Fatalf("have err: %q, want: %q", have, want)
			}
			res.Error = ""
			tc.result.Error = ""

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fatalf("response mismatch(-have, +want): %s", diff)
			}
		})
	}
}

type testSource struct {
	fn func() error
}

var (
	_ repos.Source     = &testSource{}
	_ repos.UserSource = &testSource{}
)

func (t testSource) ListRepos(_ context.Context, _ chan repos.SourceResult) {
}

func (t testSource) ExternalServices() types.ExternalServices {
	return nil
}

func (t testSource) CheckConnection(_ context.Context) error {
	return nil
}

func (t testSource) WithAuthenticator(_ auth.Authenticator) (repos.Source, error) {
	return t, nil
}

func (t testSource) ValidateAuthenticator(_ context.Context) error {
	return t.fn()
}

func TestGrpcErrToStatus(t *testing.T) {
	testCases := []struct {
		description  string
		input        error
		expectedCode int
	}{
		{
			description:  "nil error",
			input:        nil,
			expectedCode: http.StatusOK,
		},
		{
			description:  "non-status error",
			input:        errors.New("non-status error"),
			expectedCode: http.StatusInternalServerError,
		},

		{
			description:  "status error context.Canceled",
			input:        context.Canceled,
			expectedCode: http.StatusInternalServerError,
		},
		{
			description:  "status error context.DeadlineExceeded",
			input:        context.DeadlineExceeded,
			expectedCode: http.StatusInternalServerError,
		},
		{
			description:  "status error codes.NotFound",
			input:        status.Errorf(codes.NotFound, "not found"),
			expectedCode: http.StatusNotFound,
		},
		{
			description:  "status error codes.Internal",
			input:        status.Errorf(codes.Internal, "internal error"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			description:  "status error codes.InvalidArgument",
			input:        status.Errorf(codes.InvalidArgument, "invalid argument"),
			expectedCode: http.StatusBadRequest,
		},

		{
			description:  "status error codes.PermissionDenied",
			input:        status.Errorf(codes.PermissionDenied, "permission denied"),
			expectedCode: http.StatusUnauthorized,
		},

		{
			description:  "status error codes.Unavailable",
			input:        status.Errorf(codes.Unavailable, "unavailable"),
			expectedCode: http.StatusServiceUnavailable,
		},

		{
			description:  "status error codes.unimplemented",
			input:        status.Errorf(codes.Unimplemented, "unimplemented"),
			expectedCode: http.StatusNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := grpcErrToStatus(tc.input)
			if result != tc.expectedCode {
				t.Errorf("Expected status code %d, but got %d", tc.expectedCode, result)
			}
		})
	}
}
