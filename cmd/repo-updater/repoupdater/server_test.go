package repoupdater

import (
	"bytes"
	"context"
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

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
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
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
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
"URL": "https://github.com",
"Token": "secret-token"
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
			mockRepos := database.NewMockRepoStore()
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
			srv := httptest.NewServer(s.Handler())
			defer srv.Close()
			cli := repoupdater.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.EnqueueRepoUpdate(ctx, tc.repo)
			if have, want := fmt.Sprint(err), tc.err; have != want {
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
		Config:       extsvc.NewEmptyConfig(),
	}
	awsSource := types.ExternalService{
		Kind:   extsvc.KindAWSCodeCommit,
		Config: extsvc.NewEmptyConfig(),
	}
	gitlabSource := types.ExternalService{
		Kind:         extsvc.KindGitLab,
		CloudDefault: true,
		Config:       extsvc.NewEmptyConfig(),
	}

	npmSource := types.ExternalService{
		Kind:   extsvc.KindNpmPackages,
		Config: extsvc.NewEmptyConfig(),
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
				// In order for new versions of package repos to be synced quickly, it's necessary to enqueue
				// a high priority git update.
				Update: true,
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
				Logger:  logger,
				Now:     clock.Now,
				Store:   store,
				Sourcer: repos.NewFakeSourcer(nil, tc.src),
			}

			scheduler := repos.NewUpdateScheduler(logtest.Scoped(t), database.NewMockDB())

			s := &Server{
				Logger:    logger,
				Syncer:    syncer,
				Store:     store,
				Scheduler: scheduler,
			}

			srv := httptest.NewServer(s.Handler())
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

			if tc.args.Update {
				scheduleInfo := scheduler.ScheduleInfo(res.Repo.ID)
				if have, want := scheduleInfo.Queue.Priority, 1; have != want { // highPriority
					t.Fatalf("scheduler update priority mismatch: have %d, want %d", have, want)
				}
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
func (s *fakeScheduler) ScheduleInfo(id api.RepoID) *protocol.RepoUpdateSchedulerInfoResult {
	return &protocol.RepoUpdateSchedulerInfoResult{}
}

type fakePermsSyncer struct{}

func (*fakePermsSyncer) ScheduleUsers(ctx context.Context, opts authz.FetchPermsOptions, userIDs ...int32) {
}

func (*fakePermsSyncer) ScheduleRepos(ctx context.Context, repoIDs ...api.RepoID) {
}

func TestServer_handleSchedulePermsSync(t *testing.T) {
	tests := []struct {
		name           string
		permsSyncer    *fakePermsSyncer
		body           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "PermsSyncer not available",
			wantStatusCode: http.StatusForbidden,
			wantBody:       "null",
		},
		{
			name:           "bad JSON",
			permsSyncer:    &fakePermsSyncer{},
			body:           "{",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "unexpected EOF",
		},
		{
			name:           "missing ids",
			permsSyncer:    &fakePermsSyncer{},
			body:           "{}",
			wantStatusCode: http.StatusBadRequest,
			wantBody:       "neither user IDs nor repo IDs was provided in request (must provide at least one)",
		},

		{
			name:           "successful call with user IDs",
			permsSyncer:    &fakePermsSyncer{},
			body:           `{"user_ids": [1]}`,
			wantStatusCode: http.StatusOK,
			wantBody:       "null",
		},
		{
			name:           "successful call with repo IDs",
			permsSyncer:    &fakePermsSyncer{},
			body:           `{"repo_ids":[1]}`,
			wantStatusCode: http.StatusOK,
			wantBody:       "null",
		},
		{
			name:           "successful call with both IDs",
			permsSyncer:    &fakePermsSyncer{},
			body:           `{"user_ids": [1], "repo_ids":[1]}`,
			wantStatusCode: http.StatusOK,
			wantBody:       "null",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/schedule-perms-sync", strings.NewReader(test.body))
			w := httptest.NewRecorder()

			s := &Server{Logger: logtest.Scoped(t)}
			// NOTE: An interface has nil value is not a nil interface,
			// so should only assign to the interface when the value is not nil.
			if test.permsSyncer != nil {
				s.PermsSyncer = test.permsSyncer
			}
			s.handleSchedulePermsSync(w, r)

			if w.Code != test.wantStatusCode {
				t.Fatalf("Code: want %v but got %v", test.wantStatusCode, w.Code)
			} else if diff := cmp.Diff(test.wantBody, w.Body.String()); diff != "" {
				t.Fatalf("Body mismatch (-want +got):\n%s", diff)
			}
		})
	}
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

type testSource struct {
	fn func() error
}

var (
	_ repos.Source     = &testSource{}
	_ repos.UserSource = &testSource{}
)

func (t testSource) ListRepos(ctx context.Context, results chan repos.SourceResult) {
}

func (t testSource) ExternalServices() types.ExternalServices {
	return nil
}

func (t testSource) WithAuthenticator(a auth.Authenticator) (repos.Source, error) {
	return t, nil
}

func (t testSource) ValidateAuthenticator(ctx context.Context) error {
	return t.fn()
}
