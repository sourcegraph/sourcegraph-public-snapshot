package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

func TestGitServiceHandlers(t *testing.T) {
	db := dbmocks.NewMockDB()
	grpcServer := grpc.NewServer()
	dummyCodeIntelHandler := func(_ bool) http.Handler { return noopHandler }
	dummyComputeStreamHandler := func() http.Handler { return noopHandler }

	m := mux.NewRouter()

	RegisterInternalServices(m, grpcServer, db, nil, dummyCodeIntelHandler, nil, dummyComputeStreamHandler, nil)

	gitService := &gitServiceHandler{
		Gitserver: mockAddrForRepo{},
	}
	handler := JsonMiddleware(&ErrorHandler{
		Logger: logtest.Scoped(t),
		// Internal endpoints can expose sensitive errors
		WriteErrBody: true,
	})
	m.Get(gitInfoRefs).Handler(handler(gitService.serveInfoRefs()))
	m.Get(gitUploadPack).Handler(handler(gitService.serveGitUploadPack()))

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

func (mockAddrForRepo) AddrForRepo(ctx context.Context, name api.RepoName) string {
	return strings.ReplaceAll(string(name), "/", ".") + ".gitserver"
}
