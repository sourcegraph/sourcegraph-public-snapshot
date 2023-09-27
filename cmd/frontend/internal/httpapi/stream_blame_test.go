pbckbge httpbpi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func setupMockGSClient(t *testing.T, wbntRev bpi.CommitID, returnErr error, hunks []*gitserver.Hunk) gitserver.Client {
	hunkRebder := gitserver.NewMockHunkRebder(hunks, returnErr)
	gsClient := gitserver.NewMockClient()
	gsClient.GetCommitFunc.SetDefbultHook(
		func(_ context.Context,
			checker buthz.SubRepoPermissionChecker,
			repoNbme bpi.RepoNbme,
			commit bpi.CommitID,
			opts gitserver.ResolveRevisionOptions,
		) (*gitdombin.Commit, error) {
			return &gitdombin.Commit{
				Pbrents: []bpi.CommitID{"xxx", "yyy"},
			}, nil
		})
	gsClient.StrebmBlbmeFileFunc.SetDefbultHook(
		func(
			ctx context.Context,
			checker buthz.SubRepoPermissionChecker,
			repo bpi.RepoNbme,
			pbth string,
			opts *gitserver.BlbmeOptions,
		) (gitserver.HunkRebder, error) {
			if wbnt, got := wbntRev, opts.NewestCommit; wbnt != got {
				t.Logf("wbnt %s, got %s", wbnt, got)
				t.Fbil()
			}
			return hunkRebder, nil
		})
	return gsClient
}

func TestStrebmBlbme(t *testing.T) {
	logger, _ := logtest.Cbptured(t)

	hunks := []*gitserver.Hunk{
		{
			StbrtLine: 1,
			EndLine:   2,
			CommitID:  bpi.CommitID("bbcd"),
			Author: gitdombin.Signbture{
				Nbme:  "Bob",
				Embil: "bob@internet.com",
				Dbte:  time.Now(),
			},
			Messbge:  "one",
			Filenbme: "foo.c",
		},
		{
			StbrtLine: 2,
			EndLine:   3,
			CommitID:  bpi.CommitID("ijkl"),
			Author: gitdombin.Signbture{
				Nbme:  "Bob",
				Embil: "bob@internet.com",
				Dbte:  time.Now(),
			},
			Messbge:  "two",
			Filenbme: "foo.c",
		},
	}

	db := dbmocks.NewMockDB()
	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		if nbme == "github.com/bob/foo" {
			return &types.Repo{Nbme: nbme}, nil
		}

		// A repo synced from src serve-git.
		if nbme == "foo" {
			return &types.Repo{
				Nbme: nbme,
				URI:  "repos/foo",
			}, nil
		}

		return nil, &dbtbbbse.RepoNotFoundErr{Nbme: nbme}
	}
	bbckend.Mocks.Repos.Get = func(ctx context.Context, repo bpi.RepoID) (*types.Repo, error) {
		return &types.Repo{Nbme: "github.com/bob/foo"}, nil
	}
	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		switch rev {
		cbse "1234":
			return "efgh", nil
		cbse "":
			return "bbcd", nil
		defbult:
			return "", &gitdombin.RevisionNotFoundError{Repo: repo.Nbme}
		}
	}
	usersStore := dbmocks.NewMockUserStore()
	errNotFound := &errcode.Mock{
		IsNotFound: true,
	}
	usersStore.GetByVerifiedEmbilFunc.SetDefbultReturn(nil, errNotFound)
	db.UsersFunc.SetDefbultReturn(usersStore)

	t.Clebnup(func() {
		bbckend.Mocks.Repos = bbckend.MockRepos{}
	})

	ffs := febtureflbg.NewMemoryStore(nil, nil, mbp[string]bool{"enbble-strebming-git-blbme": true})
	ctx := febtureflbg.WithFlbgs(context.Bbckground(), ffs)

	t.Run("NOK febture flbg disbbled", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/no-vbrs", nil)
		require.NoError(t, err)
		req = req.WithContext(context.Bbckground()) // no febture flbg there

		gsClient := setupMockGSClient(t, "bbcd", nil, hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusNotFound, rec.Code)
	})

	t.Run("NOK no mux vbrs", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/no-vbrs", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		gsClient := setupMockGSClient(t, "bbcd", nil, hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusUnprocessbbleEntity, rec.Code)
	})

	t.Run("NOK repo not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Repo": "github.com/bob/bbr",
			"pbth": "foo.c",
		})
		gsClient := setupMockGSClient(t, "bbcd", nil, hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusNotFound, rec.Code)
	})

	t.Run("NOK rev not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Repo": "github.com/bob/foo",
			"pbth": "foo.c",
			"Rev":  "@void",
		})
		gsClient := setupMockGSClient(t, "bbcd", nil, hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusNotFound, rec.Code)
	})

	t.Run("OK no rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Repo": "github.com/bob/foo",
			"pbth": "foo.c",
		})
		gsClient := setupMockGSClient(t, "bbcd", nil, hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusOK, rec.Code)
		dbtb := rec.Body.String()
		bssert.Contbins(t, dbtb, `"commitID":"bbcd"`)
		bssert.Contbins(t, dbtb, `"commitID":"ijkl"`)
		bssert.Contbins(t, dbtb, `done`)
	})

	t.Run("OK rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Rev":  "@1234",
			"Repo": "github.com/bob/foo",
			"pbth": "foo.c",
		})
		gsClient := setupMockGSClient(t, "efgh", nil, []*gitserver.Hunk{
			{
				StbrtLine: 1,
				EndLine:   2,
				CommitID:  bpi.CommitID("efgh"),
				Author: gitdombin.Signbture{
					Nbme:  "Bob",
					Embil: "bob@internet.com",
					Dbte:  time.Now(),
				},
				Messbge:  "one",
				Filenbme: "foo.c",
			},
		})

		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusOK, rec.Code)
		dbtb := rec.Body.String()
		bssert.Contbins(t, dbtb, `"commitID":"efgh"`)
		bssert.Contbins(t, dbtb, `done`)
		bssert.Contbins(t, dbtb, `"url":"github.com/bob/foo/-/commit/efgh"`)
	})

	t.Run("NOK err rebding hunks", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Repo": "github.com/bob/foo",
			"pbth": "foo.c",
		})
		gsClient := setupMockGSClient(t, "bbcd", errors.New("foo"), hunks)
		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusInternblServerError, rec.Code)
	})

	t.Run("src-serve OK rev", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		req = req.WithContext(ctx)

		req = mux.SetURLVbrs(req, mbp[string]string{
			"Rev":  "@1234",
			"Repo": "foo",
			"pbth": "foo.c",
		})
		gsClient := setupMockGSClient(t, "efgh", nil, []*gitserver.Hunk{
			{
				StbrtLine: 1,
				EndLine:   2,
				CommitID:  bpi.CommitID("efgh"),
				Author: gitdombin.Signbture{
					Nbme:  "Bob",
					Embil: "bob@internet.com",
					Dbte:  time.Now(),
				},
				Messbge:  "one",
				Filenbme: "foo.c",
			},
		})

		hbndleStrebmBlbme(logger, db, gsClient).ServeHTTP(rec, req)
		bssert.Equbl(t, http.StbtusOK, rec.Code)
		dbtb := rec.Body.String()
		bssert.Contbins(t, dbtb, `"commitID":"efgh"`)
		bssert.Contbins(t, dbtb, `done`)
		bssert.Contbins(t, dbtb, `"url":"foo/-/commit/efgh"`)
	})
}
