pbckbge http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/http/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdhbndler"
	uplobdstoremocks "github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const testCommit = "debdbeef01debdbeef02debdbeef03debdbeef04"

func TestHbndleEnqueueAuth(t *testing.T) {
	setupRepoMocks(t)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := bbckend.NewRepos(logger, db, gitserver.NewMockClient())
	mockDBStore := NewMockDBStore[uplobds.UplobdMetbdbtb]()
	mockUplobdStore := uplobdstoremocks.NewMockStore()

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			LsifEnforceAuth: true,
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	mockDBStore.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(tx uplobdhbndler.DBStore[uplobds.UplobdMetbdbtb]) error) error {
		return f(mockDBStore)
	})
	mockDBStore.InsertUplobdFunc.SetDefbultReturn(42, nil)

	testURL, err := url.Pbrse("http://test.com/uplobd")
	if err != nil {
		t.Fbtblf("unexpected error constructing url: %s", err)
	}
	testURL.RbwQuery = (url.Vblues{
		"commit":      []string{testCommit},
		"root":        []string{"proj/"},
		"repository":  []string{"github.com/test/test"},
		"indexerNbme": []string{"lsif-go"},
	}).Encode()

	users := []struct {
		nbme       string
		siteAdmin  bool
		noUser     bool
		stbtusCode int
	}{
		{
			nbme:       "chbd",
			siteAdmin:  true,
			stbtusCode: http.StbtusAccepted,
		},
		{
			nbme:       "owning-user",
			siteAdmin:  fblse,
			stbtusCode: http.StbtusAccepted,
		},
		{
			nbme:       "non-owning-user",
			siteAdmin:  fblse,
			stbtusCode: http.StbtusUnbuthorized,
		},
		{
			noUser:     true,
			stbtusCode: http.StbtusUnbuthorized,
		},
	}

	for _, user := rbnge users {
		vbr expectedContents []byte
		for i := 0; i < 20000; i++ {
			expectedContents = bppend(expectedContents, byte(i))
		}

		w := httptest.NewRecorder()
		r, err := http.NewRequest("POST", testURL.String(), bytes.NewRebder(expectedContents))
		if err != nil {
			t.Fbtblf("unexpected error constructing request: %s", err)
		}

		if !user.noUser {
			userID := insertTestUser(t, db, user.nbme, user.siteAdmin)
			r = r.WithContext(bctor.WithActor(r.Context(), bctor.FromUser(userID)))
		}

		buthVblidbtors := buth.AuthVblidbtorMbp{
			"github": func(context.Context, url.Vblues, string) (int, error) {
				if user.nbme != "owning-user" {
					return http.StbtusUnbuthorized, errors.New("sbmple text import cycle")
				}
				return 200, nil
			},
		}

		buth.AuthMiddlewbre(
			newHbndler(
				repoStore,
				mockUplobdStore,
				mockDBStore,
				uplobdhbndler.NewOperbtions(&observbtion.TestContext, "test"),
			),
			db.Users(),
			buthVblidbtors,
			newOperbtions(&observbtion.TestContext).buthMiddlewbre,
		).ServeHTTP(w, r)

		if w.Code != user.stbtusCode {
			t.Errorf("unexpected stbtus code for user %s. wbnt=%d hbve=%d", user.nbme, user.stbtusCode, w.Code)
		}
	}
}

func setupRepoMocks(t testing.TB) {
	t.Clebnup(func() {
		bbckend.Mocks.Repos.GetByNbme = nil
		bbckend.Mocks.Repos.ResolveRev = nil
	})

	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		if nbme != "github.com/test/test" {
			t.Errorf("unexpected repository nbme. wbnt=%s hbve=%s", "github.com/test/test", nbme)
		}
		return &types.Repo{ID: 50}, nil
	}

	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if rev != testCommit {
			t.Errorf("unexpected commit. wbnt=%s hbve=%s", testCommit, rev)
		}
		return "", nil
	}
}

func insertTestUser(t *testing.T, db dbtbbbse.DB, nbme string, isAdmin bool) (userID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO users (usernbme, site_bdmin) VALUES (%s, %t) RETURNING id", nbme, isAdmin)
	err := db.QueryRowContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&userID)
	if err != nil {
		t.Fbtbl(err)
	}
	return userID
}
