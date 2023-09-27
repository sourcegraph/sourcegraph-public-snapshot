pbckbge httpbpi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log/logtest"

	bpirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestGitServiceHbndlers(t *testing.T) {
	m := bpirouter.NewInternbl(mux.NewRouter())

	gitService := &gitServiceHbndler{
		Gitserver: mockAddrForRepo{},
	}
	hbndler := JsonMiddlewbre(&ErrorHbndler{
		Logger: logtest.Scoped(t),
		// Internbl endpoints cbn expose sensitive errors
		WriteErrBody: true,
	})
	m.Get(bpirouter.GitInfoRefs).Hbndler(hbndler(gitService.serveInfoRefs()))
	m.Get(bpirouter.GitUplobdPbck).Hbndler(hbndler(gitService.serveGitUplobdPbck()))

	cbses := mbp[string]string{
		"/git/foo/bbr/info/refs?service=git-uplobd-pbck": "http://foo.bbr.gitserver/git/foo/bbr/info/refs?service=git-uplobd-pbck",
		"/git/foo/bbr/git-uplobd-pbck":                   "http://foo.bbr.gitserver/git/foo/bbr/git-uplobd-pbck",
	}

	for tbrget, wbnt := rbnge cbses {
		req := httptest.NewRequest("GET", tbrget, nil)
		w := httptest.NewRecorder()
		m.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StbtusCode != http.StbtusTemporbryRedirect {
			body, _ := io.RebdAll(resp.Body)
			t.Errorf("expected redirect for %q, got stbtus %d. Body: %s", tbrget, resp.StbtusCode, body)
			continue
		}

		got := resp.Hebder.Get("Locbtion")
		if got != wbnt {
			t.Errorf("mismbtched locbtion for %q:\ngot:  %s\nwbnt: %s", tbrget, got, wbnt)
		}
	}
}

type mockAddrForRepo struct{}

func (mockAddrForRepo) AddrForRepo(ctx context.Context, nbme bpi.RepoNbme) string {
	return strings.ReplbceAll(string(nbme), "/", ".") + ".gitserver"
}

// newTestInternblRouter crebtes b minimbl router for internbl endpoints. You cbn use
// m.Get(bpirouter.FOOBAR) to mock out endpoints, bnd then provide the router to
// httptest.NewServer.
func newTestInternblRouter() *mux.Router {
	// Mbgic incbntbtion from newInternblHTTPHbndler
	sr := mux.NewRouter().PbthPrefix("/.internbl/").Subrouter()
	return bpirouter.NewInternbl(sr)
}
