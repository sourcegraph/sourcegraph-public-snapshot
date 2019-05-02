package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/db/globalstatedb"
)

func TestServeHome(t *testing.T) {
	globalstatedb.Mock.Get = func(ctx context.Context) (*globalstatedb.State, error) {
		return &globalstatedb.State{SiteID: "a"}, nil
	}
	defer func() { globalstatedb.Mock.Get = nil }()
	siteid.Init()

	globals.ConfigurationServerFrontendOnly = &conf.Server{}
	defer func() { globals.ConfigurationServerFrontendOnly = nil }()

	check := func(t *testing.T, wantRedirectLocation string) {
		t.Helper()

		rw := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		_ = serveHome(rw, req)
		if want := http.StatusTemporaryRedirect; rw.Code != want {
			t.Errorf("got HTTP response code %d, want %d", rw.Code, want)
		}
		if got := rw.Header().Get("Location"); got != wantRedirectLocation {
			t.Errorf("got redirect location %q, want %q", got, wantRedirectLocation)
		}
	}

	t.Run("on Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset
		check(t, "https://about.sourcegraph.com")
	})
	t.Run("non-Sourcegraph.com", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(false)
		defer envvar.MockSourcegraphDotComMode(orig) // reset
		check(t, "/search")
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
