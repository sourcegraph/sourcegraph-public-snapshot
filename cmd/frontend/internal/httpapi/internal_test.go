package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

func TestGitServiceHandlers(t *testing.T) {
	m := apirouter.NewInternal(mux.NewRouter())

	gitService := &gitServiceHandler{
		Gitserver: mockAddrForRepo{},
	}
	handler := jsonMiddleware(&errorHandler{
		Logger: logtest.Scoped(t),
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

func (mockAddrForRepo) AddrForRepo(name api.RepoName) string {
	return strings.ReplaceAll(string(name), "/", ".") + ".gitserver"
}

// newTestInternalRouter creates a minimal router for internal endpoints. You can use
// m.Get(apirouter.FOOBAR) to mock out endpoints, and then provide the router to
// httptest.NewServer.
func newTestInternalRouter() *mux.Router {
	// Magic incantation from newInternalHTTPHandler
	sr := mux.NewRouter().PathPrefix("/.internal/").Subrouter()
	return apirouter.NewInternal(sr)
}

func TestDecodeSendEmail(t *testing.T) {
	m := newTestInternalRouter()
	m.Get(apirouter.SendEmail).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg, err := decodeSendEmail(r)

		assert.NoError(t, err)
		assert.Equal(t, msg.Source, "testdecode")
		assert.Contains(t, msg.To, "foobar@foobar.com")

		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(m)
	t.Cleanup(ts.Close)

	client := *internalapi.Client
	client.URL = ts.URL

	// Do not worry about error here, run assertions in the test handler
	err := client.SendEmail(context.Background(), "testdecode", txtypes.Message{
		To: []string{"foobar@foobar.com"},
	})
	assert.NoError(t, err)
}
