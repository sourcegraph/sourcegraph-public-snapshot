package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestGitServiceHandlers(t *testing.T) {
	m := apirouter.NewInternal(mux.NewRouter())

	gitService := &gitServiceHandler{
		Gitserver: mockAddrForRepo{},
	}
	handler := jsonMiddleware(&errorHandler{
		// Internal endpoints can expose sensitive errors
		WriteErrBody: true,
	})
	m.Get(apirouter.GitInfoRefs).Handler(handler(gitService.serveInfoRefs()))
	m.Get(apirouter.GitUploadPack).Handler(handler(gitService.serveGitUploadPack()))

	cases := map[string]string{
		"/git/foo/bar/info/refs?service=git-upload-pack": "http://foo.bar.gitserver/git/foo/bar/info/refs?service=git-upload-pack",
		"/git/foo/bar/git-upload-pack":                   "http://foo.bar.gitserver/git/foo/bar/git-upload-pack",
	}

	for target, want := range cases {
		req := httptest.NewRequest("GET", target, nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusTemporaryRedirect {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected redirect for %q, got status %d. Body: %s", target, resp.StatusCode, body)
			continue
		}

		got := resp.Header.Get("Location")
		if got != want {
			t.Errorf("mismatched location for %q:\ngot:  %s\nwant: %s", target, got, want)
		}
	}
}

type mockAddrForRepo struct{}

func (mockAddrForRepo) AddrForRepo(_ context.Context, name api.RepoName) (string, error) {
	return strings.ReplaceAll(string(name), "/", ".") + ".gitserver", nil
}
