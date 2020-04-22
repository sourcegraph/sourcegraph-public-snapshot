package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	uirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

func TestRedirects(t *testing.T) {
	check := func(t *testing.T, path string, wantStatusCode int, wantRedirectLocation string) {
		t.Helper()

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", path, nil)
		uirouter.Router.ServeHTTP(rw, req)
		if rw.Code != wantStatusCode {
			t.Errorf("got HTTP response code %d, want %d", rw.Code, wantStatusCode)
		}
		if got := rw.Header().Get("Location"); got != wantRedirectLocation {
			t.Errorf("got redirect location %q, want %q", got, wantRedirectLocation)
		}
	}

	t.Run("on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset
		t.Run("root", func(t *testing.T) {
			check(t, "/", http.StatusTemporaryRedirect, "https://about.sourcegraph.com")
		})
	})
	t.Run("non-Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(false)
		defer envvar.MockSourcegraphDotComMode(orig) // reset
		t.Run("root", func(t *testing.T) {
			check(t, "/", http.StatusTemporaryRedirect, "/search")
		})
	})
}

func TestRepoShortName(t *testing.T) {
	tests := []struct {
		input api.RepoName
		want  string
	}{
		{input: "repo", want: "repo"},
		{input: "github.com/foo/bar", want: "foo/bar"},
		{input: "mycompany.com/foo", want: "foo"},
	}
	for _, tst := range tests {
		t.Run(string(tst.input), func(t *testing.T) {
			got := repoShortName(tst.input)
			if got != tst.want {
				t.Fatalf("input %q got %q want %q", tst.input, got, tst.want)
			}
		})
	}
}

func TestNewCommon_repo_error(t *testing.T) {
	cases := []struct {
		name string
		rev  string
		err  error

		want string
		code int
	}{{
		name: "cloning",
		err:  &vcs.RepoNotExistError{CloneInProgress: true},
		code: 200,
	}, {
		name: "repo-404",
		err:  &vcs.RepoNotExistError{Repo: "repo-404"},
		want: "repository does not exist: repo-404",
		code: 404,
	}, {
		name: "rev-404",
		rev:  "@marco",
		err:  &gitserver.RevisionNotFoundError{Repo: "rev-404", Spec: "marco"},
		want: "revision not found: rev-404@marco",
		code: 404,
	}, {
		name: "repoupdater-not-found",
		err:  repoupdater.ErrNotFound,
		want: repoupdater.ErrNotFound.Error(),
		code: 404,
	}, {
		name: "repoupdater-unauthorized",
		err:  repoupdater.ErrUnauthorized,
		want: repoupdater.ErrUnauthorized.Error(),
		code: 401,
	}, {
		name: "github.com/sourcegraphtest/Always500Test",
		want: "error caused by Always500Test repo name",
		code: 500,
	}}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backend.Mocks.Repos.MockGetByName(t, api.RepoName(tt.name), 1)
			backend.Mocks.Repos.MockGet(t, 1)
			backend.Mocks.Repos.ResolveRev = func(context.Context, *types.Repo, string) (api.CommitID, error) {
				if tt.err != nil {
					return "", tt.err
				}
				return api.CommitID("deadbeef"), nil
			}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req = mux.SetURLVars(req, map[string]string{
				"Repo": tt.name,
				"Rev":  tt.rev,
			})

			code := 200
			got := ""
			serveError := func(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
				got = err.Error()
				code = statusCode
			}

			_, err = newCommon(httptest.NewRecorder(), req, "test", serveError)
			if err != nil {
				if got != "" || code != 200 {
					t.Fatal("serveError called and error returned from newCommon")
				}
				code = 500
				got = err.Error()
			}

			if tt.want != got {
				t.Errorf("unexpected error.\ngot:  %s\nwant: %s", got, tt.want)
			}
			if tt.code != code {
				t.Errorf("unexpected status code: got=%d want=%d", code, tt.code)
			}
		})
	}
}

func init() {
	globals.ConfigurationServerFrontendOnly = &conf.Server{}
	globalstatedb.Mock.Get = func(ctx context.Context) (*globalstatedb.State, error) {
		return &globalstatedb.State{SiteID: "a"}, nil
	}
	siteid.Init()
}
