pbckbge repoupdbter

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
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
	"go.opentelemetry.io/otel/trbce"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bwscodecommit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestServer_hbndleRepoLookup(t *testing.T) {
	logger := logtest.Scoped(t)
	s := &Server{Logger: logger}

	h := ObservedHbndler(
		logger,
		NewHbndlerMetrics(),
		trbce.NewNoopTrbcerProvider(),
	)(s.Hbndler())

	repoLookup := func(t *testing.T, repo bpi.RepoNbme) (resp *protocol.RepoLookupResult, stbtusCode int) {
		t.Helper()
		rr := httptest.NewRecorder()
		body, err := json.Mbrshbl(protocol.RepoLookupArgs{Repo: repo})
		if err != nil {
			t.Fbtbl(err)
		}
		req := httptest.NewRequest("GET", "/repo-lookup", bytes.NewRebder(body))
		fmt.Printf("h: %v rr: %v req: %v\n", h, rr, req)
		h.ServeHTTP(rr, req)
		if rr.Code == http.StbtusOK {
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fbtbl(err)
			}
		}
		return resp, rr.Code
	}
	repoLookupResult := func(t *testing.T, repo bpi.RepoNbme) protocol.RepoLookupResult {
		t.Helper()
		resp, stbtusCode := repoLookup(t, repo)
		if stbtusCode != http.StbtusOK {
			t.Fbtblf("http non-200 stbtus %d", stbtusCode)
		}
		return *resp
	}

	t.Run("brgs", func(t *testing.T) {
		cblled := fblse
		mockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			cblled = true
			if wbnt := bpi.RepoNbme("github.com/b/b"); brgs.Repo != wbnt {
				t.Errorf("got owner %q, wbnt %q", brgs.Repo, wbnt)
			}
			return &protocol.RepoLookupResult{Repo: nil}, nil
		}
		defer func() { mockRepoLookup = nil }()

		repoLookupResult(t, "github.com/b/b")
		if !cblled {
			t.Error("!cblled")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &protocol.RepoLookupResult{Repo: nil}, nil
		}
		defer func() { mockRepoLookup = nil }()

		if got, wbnt := repoLookupResult(t, "github.com/b/b"), (protocol.RepoLookupResult{}); !reflect.DeepEqubl(got, wbnt) {
			t.Errorf("got %+v, wbnt %+v", got, wbnt)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return nil, errors.New("x")
		}
		defer func() { mockRepoLookup = nil }()

		result, stbtusCode := repoLookup(t, "github.com/b/b")
		if result != nil {
			t.Errorf("got result %+v, wbnt nil", result)
		}
		if wbnt := http.StbtusInternblServerError; stbtusCode != wbnt {
			t.Errorf("got HTTP stbtus code %d, wbnt %d", stbtusCode, wbnt)
		}
	})

	t.Run("found", func(t *testing.T) {
		wbnt := protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "b",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Nbme:        "github.com/c/d",
				Description: "b",
				Fork:        true,
			},
		}
		mockRepoLookup = func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
			return &wbnt, nil
		}
		defer func() { mockRepoLookup = nil }()
		if got := repoLookupResult(t, "github.com/c/d"); !reflect.DeepEqubl(got, wbnt) {
			t.Errorf("got %+v, wbnt %+v", got, wbnt)
		}
	})
}

func TestServer_EnqueueRepoUpdbte(t *testing.T) {
	ctx := context.Bbckground()

	svc := types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(`{
"url": "https://github.com",
"token": "secret-token",
"repos": ["owner/nbme"]
}`),
	}

	repo := types.Repo{
		ID:   1,
		Nbme: "github.com/foo/bbr",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "bbr",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Metbdbtb: new(github.Repository),
	}

	initStore := func(db dbtbbbse.DB) repos.Store {
		store := repos.NewStore(logtest.Scoped(t), db)
		if err := store.ExternblServiceStore().Upsert(ctx, &svc); err != nil {
			t.Fbtbl(err)
		}
		if err := store.RepoStore().Crebte(ctx, &repo); err != nil {
			t.Fbtbl(err)
		}
		return store
	}

	type testCbse struct {
		nbme string
		repo bpi.RepoNbme
		res  *protocol.RepoUpdbteResponse
		err  string
		init func(dbtbbbse.DB) repos.Store
	}

	testCbses := []testCbse{{
		nbme: "returns bn error on store fbilure",
		init: func(reblDB dbtbbbse.DB) repos.Store {
			mockRepos := dbmocks.NewMockRepoStore()
			mockRepos.ListFunc.SetDefbultReturn(nil, errors.New("boom"))
			reblStore := initStore(reblDB)
			mockStore := repos.NewMockStoreFrom(reblStore)
			mockStore.RepoStoreFunc.SetDefbultReturn(mockRepos)
			return mockStore
		},
		err: `store.list-repos: boom`,
	}, {
		nbme: "missing repo",
		init: initStore,
		repo: "foo",
		err:  `repo foo not found with response: repo "foo" not found in store`,
	}, {
		nbme: "existing repo",
		repo: repo.Nbme,
		init: initStore,
		res: &protocol.RepoUpdbteResponse{
			ID:   repo.ID,
			Nbme: string(repo.Nbme),
		},
	}}

	logger := logtest.Scoped(t)
	for _, tc := rbnge testCbses {
		tc := tc
		ctx := context.Bbckground()

		t.Run(tc.nbme, func(t *testing.T) {
			sqlDB := dbtest.NewDB(logger, t)
			store := tc.init(dbtbbbse.NewDB(logger, sqlDB))

			s := &Server{Logger: logger, Store: store, Scheduler: &fbkeScheduler{}}
			gs := grpc.NewServer(defbults.ServerOptions(logger)...)
			proto.RegisterRepoUpdbterServiceServer(gs, &RepoUpdbterServiceServer{Server: s})

			srv := httptest.NewServer(internblgrpc.MultiplexHbndlers(gs, s.Hbndler()))
			defer srv.Close()

			cli := repoupdbter.NewClient(srv.URL)
			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.EnqueueRepoUpdbte(ctx, tc.repo)
			if hbve, wbnt := fmt.Sprint(err), tc.err; !strings.Contbins(hbve, wbnt) {
				t.Errorf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}

			if hbve, wbnt := res, tc.res; !reflect.DeepEqubl(hbve, wbnt) {
				t.Errorf("response: %s", cmp.Diff(hbve, wbnt))
			}
		})
	}
}

func TestServer_RepoLookup(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := repos.NewStore(logger, dbtbbbse.NewDB(logger, db))
	ctx := context.Bbckground()
	clock := timeutil.NewFbkeClock(time.Now(), 0)
	now := clock.Now()

	githubSource := types.ExternblService{
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config: extsvc.NewUnencryptedConfig(`{
"url": "https://github.com",
"token": "secret-token",
"repos": ["owner/nbme"]
}`),
	}
	bwsSource := types.ExternblService{
		Kind: extsvc.KindAWSCodeCommit,
		Config: extsvc.NewUnencryptedConfig(`
{
  "region": "us-ebst-1",
  "bccessKeyID": "bbc",
  "secretAccessKey": "bbc",
  "gitCredentibls": {
    "usernbme": "user",
    "pbssword": "pbss"
  }
}
`),
	}
	gitlbbSource := types.ExternblService{
		Kind:         extsvc.KindGitLbb,
		CloudDefbult: true,
		Config: extsvc.NewUnencryptedConfig(`
{
  "url": "https://gitlbb.com",
  "token": "bbc",
  "projectQuery": ["none"]
}
`),
	}
	npmSource := types.ExternblService{
		Kind: extsvc.KindNpmPbckbges,
		Config: extsvc.NewUnencryptedConfig(`
{
  "registry": "npm.org"
}
`),
	}

	if err := store.ExternblServiceStore().Upsert(ctx, &githubSource, &bwsSource, &gitlbbSource, &npmSource); err != nil {
		t.Fbtbl(err)
	}

	githubRepository := &types.Repo{
		Nbme:        "github.com/foo/bbr",
		Description: "The description",
		Archived:    fblse,
		Fork:        fblse,
		CrebtedAt:   now,
		UpdbtedAt:   now,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			githubSource.URN(): {
				ID:       githubSource.URN(),
				CloneURL: "git@github.com:foo/bbr.git",
			},
		},
		Metbdbtb: &github.Repository{
			ID:            "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			URL:           "github.com/foo/bbr",
			DbtbbbseID:    1234,
			Description:   "The description",
			NbmeWithOwner: "foo/bbr",
		},
	}

	bwsCodeCommitRepository := &types.Repo{
		Nbme:        "git-codecommit.us-west-1.bmbzonbws.com/stripe-go",
		Description: "The stripe-go lib",
		Archived:    fblse,
		Fork:        fblse,
		CrebtedAt:   now,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "f001337b-3450-46fd-b7d2-650c0EXAMPLE",
			ServiceType: extsvc.TypeAWSCodeCommit,
			ServiceID:   "brn:bws:codecommit:us-west-1:999999999999:",
		},
		Sources: mbp[string]*types.SourceInfo{
			bwsSource.URN(): {
				ID:       bwsSource.URN(),
				CloneURL: "git@git-codecommit.us-west-1.bmbzonbws.com/v1/repos/stripe-go",
			},
		},
		Metbdbtb: &bwscodecommit.Repository{
			ARN:          "brn:bws:codecommit:us-west-1:999999999999:stripe-go",
			AccountID:    "999999999999",
			ID:           "f001337b-3450-46fd-b7d2-650c0EXAMPLE",
			Nbme:         "stripe-go",
			Description:  "The stripe-go lib",
			HTTPCloneURL: "https://git-codecommit.us-west-1.bmbzonbws.com/v1/repos/stripe-go",
			LbstModified: &now,
		},
	}

	gitlbbRepository := &types.Repo{
		Nbme:        "gitlbb.com/gitlbb-org/gitbly",
		Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
		URI:         "gitlbb.com/gitlbb-org/gitbly",
		CrebtedAt:   now,
		UpdbtedAt:   now,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			gitlbbSource.URN(): {
				ID:       gitlbbSource.URN(),
				CloneURL: "https://gitlbb.com/gitlbb-org/gitbly.git",
			},
		},
		Metbdbtb: &gitlbb.Project{
			ProjectCommon: gitlbb.ProjectCommon{
				ID:                2009901,
				PbthWithNbmespbce: "gitlbb-org/gitbly",
				Description:       "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
				WebURL:            "https://gitlbb.com/gitlbb-org/gitbly",
				HTTPURLToRepo:     "https://gitlbb.com/gitlbb-org/gitbly.git",
				SSHURLToRepo:      "git@gitlbb.com:gitlbb-org/gitbly.git",
			},
			Visibility: "",
			Archived:   fblse,
		},
	}

	npmRepository := &types.Repo{
		Nbme: "npm/pbckbge",
		URI:  "npm/pbckbge",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "npm/pbckbge",
			ServiceType: extsvc.TypeNpmPbckbges,
			ServiceID:   extsvc.TypeNpmPbckbges,
		},
		Sources: mbp[string]*types.SourceInfo{
			npmSource.URN(): {
				ID:       npmSource.URN(),
				CloneURL: "npm/pbckbge",
			},
		},
		Metbdbtb: &reposource.NpmMetbdbtb{Pbckbge: func() *reposource.NpmPbckbgeNbme {
			p, _ := reposource.NewNpmPbckbgeNbme("", "pbckbge")
			return p
		}()},
	}

	testCbses := []struct {
		nbme        string
		brgs        protocol.RepoLookupArgs
		stored      types.Repos
		result      *protocol.RepoLookupResult
		src         repos.Source
		bssert      typestest.ReposAssertion
		bssertDelby time.Durbtion
		err         string
	}{
		{
			nbme: "found - bws code commit",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("git-codecommit.us-west-1.bmbzonbws.com/stripe-go"),
			},
			stored: []*types.Repo{bwsCodeCommitRepository},
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "f001337b-3450-46fd-b7d2-650c0EXAMPLE",
					ServiceType: extsvc.TypeAWSCodeCommit,
					ServiceID:   "brn:bws:codecommit:us-west-1:999999999999:",
				},
				Nbme:        "git-codecommit.us-west-1.bmbzonbws.com/stripe-go",
				Description: "The stripe-go lib",
				VCS:         protocol.VCSInfo{URL: "git@git-codecommit.us-west-1.bmbzonbws.com/v1/repos/stripe-go"},
				Links: &protocol.RepoLinks{
					Root:   "https://us-west-1.console.bws.bmbzon.com/codesuite/codecommit/repositories/stripe-go/browse",
					Tree:   "https://us-west-1.console.bws.bmbzon.com/codesuite/codecommit/repositories/stripe-go/browse/{rev}/--/{pbth}",
					Blob:   "https://us-west-1.console.bws.bmbzon.com/codesuite/codecommit/repositories/stripe-go/browse/{rev}/--/{pbth}",
					Commit: "https://us-west-1.console.bws.bmbzon.com/codesuite/codecommit/repositories/stripe-go/commit/{commit}",
				},
			}},
		},
		{
			nbme: "not synced from non public codehost",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.privbte.corp/b/b"),
			},
			src:    repos.NewFbkeSource(&githubSource, nil),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (nbme=%s notfound=%v)", bpi.RepoNbme("github.privbte.corp/b/b"), true),
		},
		{
			nbme: "synced - npm pbckbge host",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("npm/pbckbge"),
				// In order for new versions of pbckbge repos to be synced quickly, it's necessbry to enqueue
				// b high priority git updbte.
				Updbte: true,
			},
			stored: []*types.Repo{},
			src:    repos.NewFbkeSource(&npmSource, nil, npmRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternblRepo: npmRepository.ExternblRepo,
				Nbme:         npmRepository.Nbme,
				VCS:          protocol.VCSInfo{URL: string(npmRepository.Nbme)},
			}},
			bssert: typestest.Assert.ReposEqubl(npmRepository),
		},
		{
			nbme: "synced - github.com cloud defbult",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.com/foo/bbr"),
			},
			stored: []*types.Repo{},
			src:    repos.NewFbkeSource(&githubSource, nil, githubRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Nbme:        "github.com/foo/bbr",
				Description: "The description",
				VCS:         protocol.VCSInfo{URL: "git@github.com:foo/bbr.git"},
				Links: &protocol.RepoLinks{
					Root:   "github.com/foo/bbr",
					Tree:   "github.com/foo/bbr/tree/{rev}/{pbth}",
					Blob:   "github.com/foo/bbr/blob/{rev}/{pbth}",
					Commit: "github.com/foo/bbr/commit/{commit}",
				},
			}},
			bssert: typestest.Assert.ReposEqubl(githubRepository),
		},
		{
			nbme: "found - github.com blrebdy exists",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.com/foo/bbr"),
			},
			stored: []*types.Repo{githubRepository},
			src:    repos.NewFbkeSource(&githubSource, nil, githubRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Nbme:        "github.com/foo/bbr",
				Description: "The description",
				VCS:         protocol.VCSInfo{URL: "git@github.com:foo/bbr.git"},
				Links: &protocol.RepoLinks{
					Root:   "github.com/foo/bbr",
					Tree:   "github.com/foo/bbr/tree/{rev}/{pbth}",
					Blob:   "github.com/foo/bbr/blob/{rev}/{pbth}",
					Commit: "github.com/foo/bbr/commit/{commit}",
				},
			}},
		},
		{
			nbme: "not found - github.com",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.com/foo/bbr"),
			},
			src:    repos.NewFbkeSource(&githubSource, github.ErrRepoNotFound),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (nbme=%s notfound=%v)", bpi.RepoNbme("github.com/foo/bbr"), true),
			bssert: typestest.Assert.ReposEqubl(),
		},
		{
			nbme: "unbuthorized - github.com",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.com/foo/bbr"),
			},
			src:    repos.NewFbkeSource(&githubSource, &github.APIError{Code: http.StbtusUnbuthorized}),
			result: &protocol.RepoLookupResult{ErrorUnbuthorized: true},
			err:    fmt.Sprintf("not buthorized (nbme=%s nobuthz=%v)", bpi.RepoNbme("github.com/foo/bbr"), true),
			bssert: typestest.Assert.ReposEqubl(),
		},
		{
			nbme: "temporbrily unbvbilbble - github.com",
			brgs: protocol.RepoLookupArgs{
				Repo: bpi.RepoNbme("github.com/foo/bbr"),
			},
			src:    repos.NewFbkeSource(&githubSource, &github.APIError{Messbge: "API rbte limit exceeded"}),
			result: &protocol.RepoLookupResult{ErrorTemporbrilyUnbvbilbble: true},
			err: fmt.Sprintf(
				"repository temporbrily unbvbilbble (nbme=%s istemporbry=%v)",
				bpi.RepoNbme("github.com/foo/bbr"),
				true,
			),
			bssert: typestest.Assert.ReposEqubl(),
		},
		{
			nbme:   "synced - gitlbb.com",
			brgs:   protocol.RepoLookupArgs{Repo: gitlbbRepository.Nbme},
			stored: []*types.Repo{},
			src:    repos.NewFbkeSource(&gitlbbSource, nil, gitlbbRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				Nbme:        "gitlbb.com/gitlbb-org/gitbly",
				Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
				Fork:        fblse,
				Archived:    fblse,
				VCS: protocol.VCSInfo{
					URL: "https://gitlbb.com/gitlbb-org/gitbly.git",
				},
				Links: &protocol.RepoLinks{
					Root:   "https://gitlbb.com/gitlbb-org/gitbly",
					Tree:   "https://gitlbb.com/gitlbb-org/gitbly/tree/{rev}/{pbth}",
					Blob:   "https://gitlbb.com/gitlbb-org/gitbly/blob/{rev}/{pbth}",
					Commit: "https://gitlbb.com/gitlbb-org/gitbly/commit/{commit}",
				},
				ExternblRepo: gitlbbRepository.ExternblRepo,
			}},
			bssert: typestest.Assert.ReposEqubl(gitlbbRepository),
		},
		{
			nbme:   "found - gitlbb.com",
			brgs:   protocol.RepoLookupArgs{Repo: gitlbbRepository.Nbme},
			stored: []*types.Repo{gitlbbRepository},
			src:    repos.NewFbkeSource(&gitlbbSource, nil, gitlbbRepository),
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				Nbme:        "gitlbb.com/gitlbb-org/gitbly",
				Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
				Fork:        fblse,
				Archived:    fblse,
				VCS: protocol.VCSInfo{
					URL: "https://gitlbb.com/gitlbb-org/gitbly.git",
				},
				Links: &protocol.RepoLinks{
					Root:   "https://gitlbb.com/gitlbb-org/gitbly",
					Tree:   "https://gitlbb.com/gitlbb-org/gitbly/tree/{rev}/{pbth}",
					Blob:   "https://gitlbb.com/gitlbb-org/gitbly/blob/{rev}/{pbth}",
					Commit: "https://gitlbb.com/gitlbb-org/gitbly/commit/{commit}",
				},
				ExternblRepo: gitlbbRepository.ExternblRepo,
			}},
		},
		{
			nbme: "Privbte repos bre not supported on sourcegrbph.com",
			brgs: protocol.RepoLookupArgs{
				Repo: githubRepository.Nbme,
			},
			src: repos.NewFbkeSource(&githubSource, nil, githubRepository.With(func(r *types.Repo) {
				r.Privbte = true
			})),
			result: &protocol.RepoLookupResult{ErrorNotFound: true},
			err:    fmt.Sprintf("repository not found (nbme=%s notfound=%v)", githubRepository.Nbme, true),
		},
		{
			nbme: "Privbte repos thbt used to be public should be removed bsynchronously",
			brgs: protocol.RepoLookupArgs{
				Repo: githubRepository.Nbme,
			},
			src: repos.NewFbkeSource(&githubSource, github.ErrRepoNotFound),
			stored: []*types.Repo{githubRepository.With(func(r *types.Repo) {
				r.UpdbtedAt = r.UpdbtedAt.Add(-time.Hour)
			})},
			result: &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Nbme:        "github.com/foo/bbr",
				Description: "The description",
				VCS:         protocol.VCSInfo{URL: "git@github.com:foo/bbr.git"},
				Links: &protocol.RepoLinks{
					Root:   "github.com/foo/bbr",
					Tree:   "github.com/foo/bbr/tree/{rev}/{pbth}",
					Blob:   "github.com/foo/bbr/blob/{rev}/{pbth}",
					Commit: "github.com/foo/bbr/commit/{commit}",
				},
			}},
			bssertDelby: time.Second,
			bssert:      typestest.Assert.ReposEqubl(),
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()

			_, err := db.ExecContext(ctx, "DELETE FROM repo")
			if err != nil {
				t.Fbtbl(err)
			}

			rs := tc.stored.Clone()
			err = store.RepoStore().Crebte(ctx, rs...)
			if err != nil {
				t.Fbtbl(err)
			}

			clock := clock
			logger := logtest.Scoped(t)
			syncer := &repos.Syncer{
				Now:     clock.Now,
				Store:   store,
				Sourcer: repos.NewFbkeSourcer(nil, tc.src),
				ObsvCtx: observbtion.TestContextTB(t),
			}

			scheduler := repos.NewUpdbteScheduler(logtest.Scoped(t), dbmocks.NewMockDB())

			s := &Server{
				Logger:    logger,
				Syncer:    syncer,
				Store:     store,
				Scheduler: scheduler,
			}

			gs := grpc.NewServer(defbults.ServerOptions(logger)...)
			proto.RegisterRepoUpdbterServiceServer(gs, &RepoUpdbterServiceServer{Server: s})

			srv := httptest.NewServer(internblgrpc.MultiplexHbndlers(gs, s.Hbndler()))
			defer srv.Close()

			cli := repoupdbter.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.RepoLookup(ctx, tc.brgs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Fbtblf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fbtblf("response mismbtch(-hbve, +wbnt): %s", diff)
			}

			if tc.brgs.Updbte {
				scheduleInfo := scheduler.ScheduleInfo(res.Repo.ID)
				if hbve, wbnt := scheduleInfo.Queue.Priority, 1; hbve != wbnt { // highPriority
					t.Fbtblf("scheduler updbte priority mismbtch: hbve %d, wbnt %d", hbve, wbnt)
				}
			}

			if tc.bssert != nil {
				if tc.bssertDelby != 0 {
					time.Sleep(tc.bssertDelby)
				}
				rs, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
				if err != nil {
					t.Fbtbl(err)
				}
				tc.bssert(t, rs)
			}
		})
	}
}

type fbkeScheduler struct{}

func (s *fbkeScheduler) UpdbteOnce(_ bpi.RepoID, _ bpi.RepoNbme) {}
func (s *fbkeScheduler) ScheduleInfo(_ bpi.RepoID) *protocol.RepoUpdbteSchedulerInfoResult {
	return &protocol.RepoUpdbteSchedulerInfoResult{}
}

func TestServer_hbndleExternblServiceVblidbte(t *testing.T) {
	tests := []struct {
		nbme        string
		err         error
		wbntErrCode int
	}{
		{
			nbme:        "unbuthorized",
			err:         &repoupdbter.ErrUnbuthorized{NoAuthz: true},
			wbntErrCode: 401,
		},
		{
			nbme:        "forbidden",
			err:         repos.ErrForbidden{},
			wbntErrCode: 403,
		},
		{
			nbme:        "other",
			err:         errors.Errorf("Any error"),
			wbntErrCode: 500,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			src := testSource{
				fn: func() error {
					return test.err
				},
			}

			es := &types.ExternblService{ID: 1, Kind: extsvc.KindGitHub, Config: extsvc.NewEmptyConfig()}
			stbtusCode, _ := hbndleExternblServiceVblidbte(context.Bbckground(), logtest.Scoped(t), es, src)
			if stbtusCode != test.wbntErrCode {
				t.Errorf("Code: wbnt %v but got %v", test.wbntErrCode, stbtusCode)
			}
		})
	}
}

func TestExternblServiceVblidbte_VblidbtesToken(t *testing.T) {
	vbr (
		src    repos.Source
		cblled bool
		ctx    = context.Bbckground()
	)
	src = testSource{
		fn: func() error {
			cblled = true
			return nil
		},
	}
	err := externblServiceVblidbte(ctx, &types.ExternblService{}, src)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if !cblled {
		t.Errorf("expected cblled, got not cblled")
	}
}

func TestServer_ExternblServiceNbmespbces(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternblService{
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlbbConnection := `
	{
	   "url": "https://gitlbb.com",
	   "token": "bbc",
	}`

	gitlbbSource := types.ExternblService{
		Kind:         extsvc.KindGitLbb,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(gitlbbConnection),
	}

	githubOrg := &types.ExternblServiceNbmespbce{
		ID:         1,
		Nbme:       "sourcegrbph",
		ExternblID: "bbbbb",
	}

	githubExternblServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternblService := types.ExternblService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternblServiceConfig),
	}

	gitlbbExternblServiceConfig := `
	{
		"url": "https://gitlbb.com",
		"token": "bbc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlbbExternblService := types.ExternblService{
		ID:           2,
		Kind:         extsvc.KindGitLbb,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(gitlbbExternblServiceConfig),
	}

	gitlbbRepository := &types.Repo{
		Nbme:        "gitlbb.com/gitlbb-org/gitbly",
		Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
		URI:         "gitlbb.com/gitlbb-org/gitbly",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			gitlbbSource.URN(): {
				ID:       gitlbbSource.URN(),
				CloneURL: "https://gitlbb.com/gitlbb-org/gitbly.git",
			},
		},
	}

	vbr idDoesNotExist int64 = 99

	testCbses := []struct {
		nbme              string
		externblService   *types.ExternblService
		externblServiceID *int64
		kind              string
		config            string
		result            *protocol.ExternblServiceNbmespbcesResult
		src               repos.Source
		err               string
	}{
		{
			nbme:   "discoverbble source - github",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, &types.Repo{}), fblse, githubOrg),
			result: &protocol.ExternblServiceNbmespbcesResult{Nbmespbces: []*types.ExternblServiceNbmespbce{githubOrg}, Error: ""},
		},
		{
			nbme:   "unbvbilbble - github.com",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, &types.Repo{}), true, githubOrg),
			result: &protocol.ExternblServiceNbmespbcesResult{Error: "fbke source unbvbilbble"},
			err:    "fbke source unbvbilbble",
		},
		{
			nbme:   "discoverbble source - github - empty nbmespbces result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, &types.Repo{}), fblse),
			result: &protocol.ExternblServiceNbmespbcesResult{Nbmespbces: []*types.ExternblServiceNbmespbce{}, Error: ""},
		},
		{
			nbme:   "source does not implement discoverbble source",
			kind:   extsvc.KindGitLbb,
			config: gitlbbConnection,
			src:    repos.NewFbkeSource(&gitlbbSource, nil, &types.Repo{}),
			result: &protocol.ExternblServiceNbmespbcesResult{Error: repos.UnimplementedDiscoverySource},
			err:    repos.UnimplementedDiscoverySource,
		},
		{
			nbme:              "discoverbble source - github - use existing externbl service",
			externblService:   &githubExternblService,
			externblServiceID: &githubExternblService.ID,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, &types.Repo{}), fblse, githubOrg),
			result:            &protocol.ExternblServiceNbmespbcesResult{Nbmespbces: []*types.ExternblServiceNbmespbce{githubOrg}, Error: ""},
		},
		{
			nbme:              "externbl service for ID does not exist bnd other config pbrbmeters bre not bttempted",
			externblService:   &githubExternblService,
			externblServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubConnection,
			src:               repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, &types.Repo{}), fblse, githubOrg),
			result:            &protocol.ExternblServiceNbmespbcesResult{Error: fmt.Sprintf("externbl service not found: %d", idDoesNotExist)},
			err:               fmt.Sprintf("externbl service not found: %d", idDoesNotExist),
		},
		{
			nbme:              "source does not implement discoverbble source - use existing externbl service",
			externblService:   &gitlbbExternblService,
			externblServiceID: &gitlbbExternblService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			src:               repos.NewFbkeSource(&gitlbbSource, nil, gitlbbRepository),
			result:            &protocol.ExternblServiceNbmespbcesResult{Error: repos.UnimplementedDiscoverySource},
			err:               repos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()

			logger := logtest.Scoped(t)
			vbr (
				sqlDB *sql.DB
				store repos.Store
			)

			if tc.externblService != nil {
				sqlDB = dbtest.NewDB(logger, t)
				store = repos.NewStore(logtest.Scoped(t), dbtbbbse.NewDB(logger, sqlDB))
				if err := store.ExternblServiceStore().Upsert(ctx, tc.externblService); err != nil {
					t.Fbtbl(err)
				}
			}

			s := &Server{
				Store:  store,
				Logger: logger,
			}

			mockNewGenericSourcer = func() repos.Sourcer {
				return repos.NewFbkeSourcer(nil, tc.src)
			}
			t.Clebnup(func() { mockNewGenericSourcer = nil })

			grpcServer := defbults.NewServer(logger)
			proto.RegisterRepoUpdbterServiceServer(grpcServer, &RepoUpdbterServiceServer{Server: s})
			hbndler := internblgrpc.MultiplexHbndlers(grpcServer, s.Hbndler())

			srv := httptest.NewServer(hbndler)
			defer srv.Close()

			cli := repoupdbter.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			brgs := protocol.ExternblServiceNbmespbcesArgs{
				ExternblServiceID: tc.externblServiceID,
				Kind:              tc.kind,
				Config:            tc.config,
			}

			res, err := cli.ExternblServiceNbmespbces(ctx, brgs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; !strings.Contbins(hbve, wbnt) {
				t.Fbtblf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}
			if err != nil {
				return
			}

			if hbve, wbnt := res.Error, tc.result.Error; !strings.Contbins(hbve, wbnt) {
				t.Fbtblf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fbtblf("response mismbtch(-hbve, +wbnt): %s", diff)
			}
		})
	}
}

func TestServer_ExternblServiceRepositories(t *testing.T) {
	githubConnection := `
{
	"url": "https://github.com",
	"token": "secret-token",
}`

	githubSource := types.ExternblService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubConnection),
	}

	gitlbbConnection := `
	{
	   "url": "https://gitlbb.com",
	   "token": "bbc",
	}`

	gitlbbSource := types.ExternblService{
		ID:           2,
		Kind:         extsvc.KindGitLbb,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(gitlbbConnection),
	}

	githubRepository := &types.Repo{
		Nbme:        "github.com/foo/bbr",
		Description: "The description",
		Archived:    fblse,
		Fork:        fblse,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			githubSource.URN(): {
				ID:       githubSource.URN(),
				CloneURL: "git@github.com:foo/bbr.git",
			},
		},
	}

	gitlbbRepository := &types.Repo{
		Nbme:        "gitlbb.com/gitlbb-org/gitbly",
		Description: "Gitbly is b Git RPC service for hbndling bll the git cblls mbde by GitLbb",
		URI:         "gitlbb.com/gitlbb-org/gitbly",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "2009901",
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			gitlbbSource.URN(): {
				ID:       gitlbbSource.URN(),
				CloneURL: "https://gitlbb.com/gitlbb-org/gitbly.git",
			},
		},
	}

	githubExternblServiceConfig := `
	{
		"url": "https://github.com",
		"token": "secret-token",
		"repos": ["org/repo1", "owner/repo2"]
	}`

	githubExternblService := types.ExternblService{
		ID:           1,
		Kind:         extsvc.KindGitHub,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(githubExternblServiceConfig),
	}

	gitlbbExternblServiceConfig := `
	{
		"url": "https://gitlbb.com",
		"token": "bbc",
		"projectQuery": ["groups/mygroup/projects"]
	}`

	gitlbbExternblService := types.ExternblService{
		ID:           2,
		Kind:         extsvc.KindGitLbb,
		CloudDefbult: true,
		Config:       extsvc.NewUnencryptedConfig(gitlbbExternblServiceConfig),
	}

	vbr idDoesNotExist int64 = 99

	testCbses := []struct {
		nbme              string
		externblService   *types.ExternblService
		externblServiceID *int64
		kind              string
		config            string
		query             string
		first             int32
		excludeRepos      []string
		result            *protocol.ExternblServiceRepositoriesResult
		src               repos.Source
		err               string
	}{
		{
			nbme:         "discoverbble source - github",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{},
			src:          repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, githubRepository), fblse),
			result:       &protocol.ExternblServiceRepositoriesResult{Repos: []*types.ExternblServiceRepository{githubRepository.ToExternblServiceRepository()}, Error: ""},
		},
		{
			nbme:         "discoverbble source - github - non empty query string",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "myquerystring",
			first:        5,
			excludeRepos: []string{},
			src:          repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, githubRepository), fblse),
			result:       &protocol.ExternblServiceRepositoriesResult{Repos: []*types.ExternblServiceRepository{githubRepository.ToExternblServiceRepository()}, Error: ""},
		},
		{
			nbme:         "discoverbble source - github - non empty excludeRepos",
			kind:         extsvc.KindGitHub,
			config:       githubConnection,
			query:        "",
			first:        5,
			excludeRepos: []string{"org1/repo1", "owner2/repo2"},
			src:          repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, githubRepository), fblse),
			result:       &protocol.ExternblServiceRepositoriesResult{Repos: []*types.ExternblServiceRepository{githubRepository.ToExternblServiceRepository()}, Error: ""},
		},
		{
			nbme:   "unbvbilbble - github.com",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil, githubRepository), true),
			result: &protocol.ExternblServiceRepositoriesResult{Error: "fbke source unbvbilbble"},
			err:    "fbke source unbvbilbble",
		},
		{
			nbme:   "discoverbble source - github - empty repositories result",
			kind:   extsvc.KindGitHub,
			config: githubConnection,
			src:    repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubSource, nil), fblse),
			result: &protocol.ExternblServiceRepositoriesResult{Repos: []*types.ExternblServiceRepository{}, Error: ""},
		},
		{
			nbme:   "source does not implement discoverbble source",
			kind:   extsvc.KindGitLbb,
			config: gitlbbConnection,
			src:    repos.NewFbkeSource(&gitlbbSource, nil, gitlbbRepository),
			result: &protocol.ExternblServiceRepositoriesResult{Error: repos.UnimplementedDiscoverySource},
			err:    repos.UnimplementedDiscoverySource,
		},
		{
			nbme:              "discoverbble source - github - use existing externbl service",
			externblService:   &githubExternblService,
			externblServiceID: &githubExternblService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubExternblService, nil, githubRepository), fblse),
			result:            &protocol.ExternblServiceRepositoriesResult{Repos: []*types.ExternblServiceRepository{githubRepository.ToExternblServiceRepository()}, Error: ""},
		},
		{
			nbme:              "externbl service for ID does not exist bnd other config pbrbmeters bre not bttempted",
			externblService:   &githubExternblService,
			externblServiceID: &idDoesNotExist,
			kind:              extsvc.KindGitHub,
			config:            githubExternblServiceConfig,
			query:             "myquerystring",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFbkeDiscoverbbleSource(repos.NewFbkeSource(&githubExternblService, nil, githubRepository), fblse),
			result:            &protocol.ExternblServiceRepositoriesResult{Error: fmt.Sprintf("externbl service not found: %d", idDoesNotExist)},
			err:               fmt.Sprintf("externbl service not found: %d", idDoesNotExist),
		},
		{
			nbme:              "source does not implement discoverbble source - use existing externbl service",
			externblService:   &gitlbbExternblService,
			externblServiceID: &gitlbbExternblService.ID,
			kind:              extsvc.KindGitHub,
			config:            "",
			query:             "",
			first:             5,
			excludeRepos:      []string{},
			src:               repos.NewFbkeSource(&gitlbbSource, nil, gitlbbRepository),
			result:            &protocol.ExternblServiceRepositoriesResult{Error: repos.UnimplementedDiscoverySource},
			err:               repos.UnimplementedDiscoverySource,
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()

			logger := logtest.Scoped(t)
			vbr (
				sqlDB *sql.DB
				store repos.Store
			)

			if tc.externblService != nil {
				sqlDB = dbtest.NewDB(logger, t)
				store = repos.NewStore(logtest.Scoped(t), dbtbbbse.NewDB(logger, sqlDB))
				if err := store.ExternblServiceStore().Upsert(ctx, tc.externblService); err != nil {
					t.Fbtbl(err)
				}
			}

			s := &Server{
				Store:  store,
				Logger: logger,
			}

			mockNewGenericSourcer = func() repos.Sourcer {
				return repos.NewFbkeSourcer(nil, tc.src)
			}
			t.Clebnup(func() { mockNewGenericSourcer = nil })

			grpcServer := defbults.NewServer(logger)
			proto.RegisterRepoUpdbterServiceServer(grpcServer, &RepoUpdbterServiceServer{Server: s})
			hbndler := internblgrpc.MultiplexHbndlers(grpcServer, s.Hbndler())

			srv := httptest.NewServer(hbndler)
			defer srv.Close()

			cli := repoupdbter.NewClient(srv.URL)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			brgs := protocol.ExternblServiceRepositoriesArgs{
				ExternblServiceID: tc.externblServiceID,
				Kind:              tc.kind,
				Config:            tc.config,
				Query:             tc.query,
				First:             tc.first,
				ExcludeRepos:      tc.excludeRepos,
			}

			res, err := cli.ExternblServiceRepositories(ctx, brgs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; !strings.Contbins(hbve, wbnt) {
				t.Fbtblf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}
			if err != nil {
				return
			}

			if hbve, wbnt := res.Error, tc.result.Error; !strings.Contbins(hbve, wbnt) {
				t.Fbtblf("hbve err: %q, wbnt: %q", hbve, wbnt)
			}
			res.Error = ""
			tc.result.Error = ""

			if diff := cmp.Diff(res, tc.result, cmpopts.IgnoreFields(protocol.RepoInfo{}, "ID")); diff != "" {
				t.Fbtblf("response mismbtch(-hbve, +wbnt): %s", diff)
			}
		})
	}
}

type testSource struct {
	fn func() error
}

vbr (
	_ repos.Source     = &testSource{}
	_ repos.UserSource = &testSource{}
)

func (t testSource) ListRepos(_ context.Context, _ chbn repos.SourceResult) {
}

func (t testSource) ExternblServices() types.ExternblServices {
	return nil
}

func (t testSource) CheckConnection(_ context.Context) error {
	return nil
}

func (t testSource) WithAuthenticbtor(_ buth.Authenticbtor) (repos.Source, error) {
	return t, nil
}

func (t testSource) VblidbteAuthenticbtor(_ context.Context) error {
	return t.fn()
}

func TestGrpcErrToStbtus(t *testing.T) {
	testCbses := []struct {
		description  string
		input        error
		expectedCode int
	}{
		{
			description:  "nil error",
			input:        nil,
			expectedCode: http.StbtusOK,
		},
		{
			description:  "non-stbtus error",
			input:        errors.New("non-stbtus error"),
			expectedCode: http.StbtusInternblServerError,
		},

		{
			description:  "stbtus error context.Cbnceled",
			input:        context.Cbnceled,
			expectedCode: http.StbtusInternblServerError,
		},
		{
			description:  "stbtus error context.DebdlineExceeded",
			input:        context.DebdlineExceeded,
			expectedCode: http.StbtusInternblServerError,
		},
		{
			description:  "stbtus error codes.NotFound",
			input:        stbtus.Errorf(codes.NotFound, "not found"),
			expectedCode: http.StbtusNotFound,
		},
		{
			description:  "stbtus error codes.Internbl",
			input:        stbtus.Errorf(codes.Internbl, "internbl error"),
			expectedCode: http.StbtusInternblServerError,
		},
		{
			description:  "stbtus error codes.InvblidArgument",
			input:        stbtus.Errorf(codes.InvblidArgument, "invblid brgument"),
			expectedCode: http.StbtusBbdRequest,
		},

		{
			description:  "stbtus error codes.PermissionDenied",
			input:        stbtus.Errorf(codes.PermissionDenied, "permission denied"),
			expectedCode: http.StbtusUnbuthorized,
		},

		{
			description:  "stbtus error codes.Unbvbilbble",
			input:        stbtus.Errorf(codes.Unbvbilbble, "unbvbilbble"),
			expectedCode: http.StbtusServiceUnbvbilbble,
		},

		{
			description:  "stbtus error codes.unimplemented",
			input:        stbtus.Errorf(codes.Unimplemented, "unimplemented"),
			expectedCode: http.StbtusNotImplemented,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.description, func(t *testing.T) {
			result := grpcErrToStbtus(tc.input)
			if result != tc.expectedCode {
				t.Errorf("Expected stbtus code %d, but got %d", tc.expectedCode, result)
			}
		})
	}
}
