package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/vcssyncer"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_Gitolite_listRepos(t *testing.T) {
	tests := []struct {
		name            string
		listRepos       map[string][]*gitolite.Repo
		configs         []*schema.GitoliteConnection
		gitoliteHost    string
		expResponseCode int
		expResponseBody []*gitolite.Repo
		wantedErr       string
	}{
		{
			name: "Simple case (git@sourcegraph.com)",
			listRepos: map[string][]*gitolite.Repo{
				"git@sourcegraph.com": {
					{Name: "myrepo", URL: "git@sourcegraph.com:myrepo"},
				},
			},
			configs: []*schema.GitoliteConnection{
				{
					Host:   "git@sourcegraph.com",
					Prefix: "sourcegraph.com/",
				},
			},
			gitoliteHost:    "git@sourcegraph.com",
			expResponseCode: 200,
			expResponseBody: []*gitolite.Repo{
				{Name: "myrepo", URL: "git@sourcegraph.com:myrepo"},
			},
		},
		{
			name: "Invalid gitoliteHost (--invalidhostnexample.com)",
			listRepos: map[string][]*gitolite.Repo{
				"git@sourcegraph.com": {
					{Name: "myrepo", URL: "git@sourcegraph.com:myrepo"},
				},
			},
			configs: []*schema.GitoliteConnection{
				{
					Host:   "git@sourcegraph.com",
					Prefix: "sourcegraph.com/",
				},
			},
			gitoliteHost:    "--invalidhostnexample.com",
			expResponseCode: 500,
			expResponseBody: nil,
			wantedErr:       "invalid gitolite host",
		},
		{
			name: "Empty (but valid) gitoliteHost",
			listRepos: map[string][]*gitolite.Repo{
				"git@gitolite.example.com": {
					{Name: "myrepo", URL: "git@gitolite.example.com:myrepo"},
				},
			},
			configs: []*schema.GitoliteConnection{
				{
					Host:   "git@gitolite.example.com",
					Prefix: "gitolite.example.com/",
				},
			},
			gitoliteHost:    "",
			expResponseCode: 200,
			expResponseBody: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := gitoliteFetcher{
				client: stubGitoliteClient{
					ListRepos_: func(ctx context.Context, host string) ([]*gitolite.Repo, error) {
						return test.listRepos[host], nil
					},
				},
			}
			resp, err := g.listRepos(context.Background(), test.gitoliteHost)
			if err != nil {
				if test.wantedErr != "" {
					if diff := cmp.Diff(test.wantedErr, err.Error()); diff != "" {
						t.Errorf("unexpected error diff:\n%s", diff)
					}
				} else {

					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(test.expResponseBody, resp); diff != "" {
				t.Errorf("unexpected response body diff:\n%s", diff)
			}
		})
	}
}

func TestCheckSSRFHeader(t *testing.T) {
	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefaultReturn(gr)
	s := &Server{
		Logger:            logtest.Scoped(t),
		ObservationCtx:    observation.TestContextTB(t),
		ReposDir:          "/testroot",
		skipCloneForTests: true,
		GetRemoteURLFunc: func(ctx context.Context, name api.RepoName) (string, error) {
			return "https://" + string(name) + ".git", nil
		},
		GetVCSSyncer: func(ctx context.Context, name api.RepoName) (vcssyncer.VCSSyncer, error) {
			return vcssyncer.NewGitRepoSyncer(logtest.Scoped(t), wrexec.NewNoOpRecordingCommandFactory()), nil
		},
		DB:         db,
		Locker:     NewRepositoryLocker(),
		RPSLimiter: ratelimit.NewInstrumentedLimiter("GitserverTest", rate.NewLimiter(rate.Inf, 10)),
	}
	h := s.Handler()

	oldFetcher := defaultGitolite
	t.Cleanup(func() {
		defaultGitolite = oldFetcher
	})
	defaultGitolite = gitoliteFetcher{
		client: stubGitoliteClient{
			ListRepos_: func(ctx context.Context, host string) ([]*gitolite.Repo, error) {
				return []*gitolite.Repo{}, nil
			},
		},
	}

	t.Run("header missing", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/list-gitolite?gitolite=127.0.0.1", nil)
		if err != nil {
			t.Fatal(err)
		}
		h.ServeHTTP(rw, r)

		assert.Equal(t, 400, rw.Code)
	})

	t.Run("header supplied", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/list-gitolite?gitolite=127.0.0.1", nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-Requested-With", "Sourcegraph")
		h.ServeHTTP(rw, r)

		assert.Equal(t, 200, rw.Code)
	})
}

type stubGitoliteClient struct {
	ListRepos_ func(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

func (c stubGitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return c.ListRepos_(ctx, host)
}
