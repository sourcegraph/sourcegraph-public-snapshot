pbckbge server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitolite"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_Gitolite_listRepos(t *testing.T) {
	tests := []struct {
		nbme            string
		listRepos       mbp[string][]*gitolite.Repo
		configs         []*schemb.GitoliteConnection
		gitoliteHost    string
		expResponseCode int
		expResponseBody []*gitolite.Repo
		wbntedErr       string
	}{
		{
			nbme: "Simple cbse (git@sourcegrbph.com)",
			listRepos: mbp[string][]*gitolite.Repo{
				"git@sourcegrbph.com": {
					{Nbme: "myrepo", URL: "git@sourcegrbph.com:myrepo"},
				},
			},
			configs: []*schemb.GitoliteConnection{
				{
					Host:   "git@sourcegrbph.com",
					Prefix: "sourcegrbph.com/",
				},
			},
			gitoliteHost:    "git@sourcegrbph.com",
			expResponseCode: 200,
			expResponseBody: []*gitolite.Repo{
				{Nbme: "myrepo", URL: "git@sourcegrbph.com:myrepo"},
			},
		},
		{
			nbme: "Invblid gitoliteHost (--invblidhostnexbmple.com)",
			listRepos: mbp[string][]*gitolite.Repo{
				"git@sourcegrbph.com": {
					{Nbme: "myrepo", URL: "git@sourcegrbph.com:myrepo"},
				},
			},
			configs: []*schemb.GitoliteConnection{
				{
					Host:   "git@sourcegrbph.com",
					Prefix: "sourcegrbph.com/",
				},
			},
			gitoliteHost:    "--invblidhostnexbmple.com",
			expResponseCode: 500,
			expResponseBody: nil,
			wbntedErr:       "invblid gitolite host",
		},
		{
			nbme: "Empty (but vblid) gitoliteHost",
			listRepos: mbp[string][]*gitolite.Repo{
				"git@gitolite.exbmple.com": {
					{Nbme: "myrepo", URL: "git@gitolite.exbmple.com:myrepo"},
				},
			},
			configs: []*schemb.GitoliteConnection{
				{
					Host:   "git@gitolite.exbmple.com",
					Prefix: "gitolite.exbmple.com/",
				},
			},
			gitoliteHost:    "",
			expResponseCode: 200,
			expResponseBody: nil,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			g := gitoliteFetcher{
				client: stubGitoliteClient{
					ListRepos_: func(ctx context.Context, host string) ([]*gitolite.Repo, error) {
						return test.listRepos[host], nil
					},
				},
			}
			resp, err := g.listRepos(context.Bbckground(), test.gitoliteHost)
			if err != nil {
				if test.wbntedErr != "" {
					if diff := cmp.Diff(test.wbntedErr, err.Error()); diff != "" {
						t.Errorf("unexpected error diff:\n%s", diff)
					}
				} else {

					t.Fbtbl(err)
				}
			}

			if diff := cmp.Diff(test.expResponseBody, resp); diff != "" {
				t.Errorf("unexpected response body diff:\n%s", diff)
			}
		})
	}
}

func TestCheckSSRFHebder(t *testing.T) {
	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefbultReturn(gr)
	s := &Server{
		Logger:            logtest.Scoped(t),
		ObservbtionCtx:    observbtion.TestContextTB(t),
		ReposDir:          "/testroot",
		skipCloneForTests: true,
		GetRemoteURLFunc: func(ctx context.Context, nbme bpi.RepoNbme) (string, error) {
			return "https://" + string(nbme) + ".git", nil
		},
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
			return NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		DB:         db,
		Locker:     NewRepositoryLocker(),
		RPSLimiter: rbtelimit.NewInstrumentedLimiter("GitserverTest", rbte.NewLimiter(rbte.Inf, 10)),
	}
	h := s.Hbndler()

	oldFetcher := defbultGitolite
	t.Clebnup(func() {
		defbultGitolite = oldFetcher
	})
	defbultGitolite = gitoliteFetcher{
		client: stubGitoliteClient{
			ListRepos_: func(ctx context.Context, host string) ([]*gitolite.Repo, error) {
				return []*gitolite.Repo{}, nil
			},
		},
	}

	t.Run("hebder missing", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/list-gitolite?gitolite=127.0.0.1", nil)
		if err != nil {
			t.Fbtbl(err)
		}
		h.ServeHTTP(rw, r)

		bssert.Equbl(t, 400, rw.Code)
	})

	t.Run("hebder supplied", func(t *testing.T) {
		rw := httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/list-gitolite?gitolite=127.0.0.1", nil)
		if err != nil {
			t.Fbtbl(err)
		}
		r.Hebder.Set("X-Requested-With", "Sourcegrbph")
		h.ServeHTTP(rw, r)

		bssert.Equbl(t, 200, rw.Code)
	})
}

type stubGitoliteClient struct {
	ListRepos_ func(ctx context.Context, host string) ([]*gitolite.Repo, error)
}

func (c stubGitoliteClient) ListRepos(ctx context.Context, host string) ([]*gitolite.Repo, error) {
	return c.ListRepos_(ctx, host)
}
