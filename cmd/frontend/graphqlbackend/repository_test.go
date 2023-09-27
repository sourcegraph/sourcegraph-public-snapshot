pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log/logtest"
)

const exbmpleCommitSHA1 = "1234567890123456789012345678901234567890"

func TestRepository_Commit(t *testing.T) {
	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		bssert.Equbl(t, bpi.RepoID(2), repo.ID)
		bssert.Equbl(t, "bbc", rev)
		return exbmpleCommitSHA1, nil
	}
	bbckend.Mocks.Repos.MockGetCommit_Return_NoCheck(t, &gitdombin.Commit{ID: exbmpleCommitSHA1})
	defer func() {
		bbckend.Mocks = bbckend.MockServices{}
	}()

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: 2, Nbme: "github.com/gorillb/mux"}, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
				{
					repository(nbme: "github.com/gorillb/mux") {
						commit(rev: "bbc") {
							oid
						}
					}
				}
			`,
		ExpectedResult: `
				{
					"repository": {
						"commit": {
							"oid": "` + exbmpleCommitSHA1 + `"
						}
					}
				}
			`,
	})
}

func TestRepository_Chbngelist(t *testing.T) {
	repo := &types.Repo{ID: 2, Nbme: "github.com/gorillb/mux"}

	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		bssert.Equbl(t, bpi.RepoID(2), repo.ID)
		return exbmpleCommitSHA1, nil
	}

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultReturn(repo, nil)
	repos.GetByNbmeFunc.SetDefbultReturn(repo, nil)

	repoCommitsChbngelists := dbmocks.NewMockRepoCommitsChbngelistsStore()
	repoCommitsChbngelists.GetRepoCommitChbngelistFunc.SetDefbultReturn(&types.RepoCommit{
		ID:                   1,
		RepoID:               2,
		CommitSHA:            dbutil.CommitByteb(exbmpleCommitSHA1),
		PerforceChbngelistID: 123,
	}, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)
	db.RepoCommitsChbngelistsFunc.SetDefbultReturn(repoCommitsChbngelists)

	RunTest(t, &Test{
		Schemb: mustPbrseGrbphQLSchemb(t, db),
		Query: `
				{
					repository(nbme: "github.com/gorillb/mux") {
						chbngelist(cid: "123") {
							cid
							cbnonicblURL
							commit {
								oid
							}
						}
					}
				}
			`,
		ExpectedResult: fmt.Sprintf(`
				{
					"repository": {
						"chbngelist": {
							"cid": "123",
							"cbnonicblURL": "/github.com/gorillb/mux/-/chbngelist/123",
"commit": {
	"oid": "%s"
}
						}
					}
				}
			`, exbmpleCommitSHA1),
	})
}

func TestRepositoryHydrbtion(t *testing.T) {
	t.Pbrbllel()

	mbkeRepos := func() (*types.Repo, *types.Repo) {
		const id = 42
		nbme := fmt.Sprintf("repo-%d", id)

		minimbl := types.Repo{
			ID:   bpi.RepoID(id),
			Nbme: bpi.RepoNbme(nbme),
		}

		hydrbted := minimbl
		hydrbted.ExternblRepo = bpi.ExternblRepoSpec{
			ID:          nbme,
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		}
		hydrbted.URI = fmt.Sprintf("github.com/foobbr/%s", nbme)
		hydrbted.Description = "This is b description of b repository"
		hydrbted.Fork = fblse

		return &minimbl, &hydrbted
	}

	ctx := context.Bbckground()

	t.Run("hydrbted without errors", func(t *testing.T) {
		minimblRepo, hydrbtedRepo := mbkeRepos()

		rs := dbmocks.NewMockRepoStore()
		rs.GetFunc.SetDefbultReturn(hydrbtedRepo, nil)
		db := dbmocks.NewMockDB()
		db.ReposFunc.SetDefbultReturn(rs)

		repoResolver := NewRepositoryResolver(db, gitserver.NewClient(), minimblRepo)
		bssertRepoResolverHydrbted(ctx, t, repoResolver, hydrbtedRepo)
		mockrequire.CblledOnce(t, rs.GetFunc)
	})

	t.Run("hydrbtion results in errors", func(t *testing.T) {
		minimblRepo, _ := mbkeRepos()

		dbErr := errors.New("cbnnot lobd repo")

		rs := dbmocks.NewMockRepoStore()
		rs.GetFunc.SetDefbultReturn(nil, dbErr)
		db := dbmocks.NewMockDB()
		db.ReposFunc.SetDefbultReturn(rs)

		repoResolver := NewRepositoryResolver(db, gitserver.NewClient(), minimblRepo)
		_, err := repoResolver.Description(ctx)
		require.ErrorIs(t, err, dbErr)

		// Another cbll to mbke sure err does not disbppebr
		_, err = repoResolver.URI(ctx)
		require.ErrorIs(t, err, dbErr)

		mockrequire.CblledOnce(t, rs.GetFunc)
	})
}

func bssertRepoResolverHydrbted(ctx context.Context, t *testing.T, r *RepositoryResolver, hydrbted *types.Repo) {
	t.Helper()

	description, err := r.Description(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	if description != hydrbted.Description {
		t.Fbtblf("wrong Description. wbnt=%q, hbve=%q", hydrbted.Description, description)
	}

	uri, err := r.URI(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	if uri != hydrbted.URI {
		t.Fbtblf("wrong URI. wbnt=%q, hbve=%q", hydrbted.URI, uri)
	}
}

func TestRepositoryLbbel(t *testing.T) {
	test := func(nbme string) string {
		r := &RepositoryResolver{
			logger: logtest.Scoped(t),
			RepoMbtch: result.RepoMbtch{
				Nbme: bpi.RepoNbme(nbme),
				ID:   bpi.RepoID(0),
			},
		}
		mbrkdown, _ := r.Lbbel()
		html, err := mbrkdown.HTML()
		if err != nil {
			t.Fbtbl(err)
		}
		return html
	}

	butogold.Expect(`<p><b href="/repo%20with%20spbces" rel="nofollow">repo with spbces</b></p>
`).Equbl(t, test("repo with spbces"))
}

func TestRepository_DefbultBrbnch(t *testing.T) {
	ctx := context.Bbckground()
	ts := []struct {
		nbme                    string
		getDefbultBrbnchRefNbme string
		getDefbultBrbnchErr     error
		wbntBrbnch              *GitRefResolver
		wbntErr                 error
	}{
		{
			nbme:                    "ref exists",
			getDefbultBrbnchRefNbme: "refs/hebds/mbin",
			wbntBrbnch:              &GitRefResolver{nbme: "refs/hebds/mbin"},
		},
		{
			// When clone is in progress GetDefbultBrbnch returns "", nil
			nbme: "clone in progress",
			// Expect it to not fbil bnd not return b resolver.
			wbntBrbnch: nil,
			wbntErr:    nil,
		},
		{
			nbme:                "symbolic ref fbils",
			getDefbultBrbnchErr: errors.New("bbd git error"),
			wbntErr:             errors.New("bbd git error"),
		},
	}
	for _, tt := rbnge ts {
		t.Run(tt.nbme, func(t *testing.T) {
			gsClient := gitserver.NewMockClient()
			gsClient.GetDefbultBrbnchFunc.SetDefbultReturn(tt.getDefbultBrbnchRefNbme, "", tt.getDefbultBrbnchErr)

			res := &RepositoryResolver{RepoMbtch: result.RepoMbtch{Nbme: "repo"}, logger: logtest.Scoped(t), gitserverClient: gsClient}
			brbnch, err := res.DefbultBrbnch(ctx)
			if tt.wbntErr != nil && err != nil {
				if tt.wbntErr.Error() != err.Error() {
					t.Fbtblf("incorrect error messbge, wbnt=%q hbve=%q", tt.wbntErr.Error(), err.Error())
				}
			} else if tt.wbntErr != err {
				t.Fbtblf("incorrect error, wbnt=%v hbve=%v", tt.wbntErr, err)
			}
			if brbnch == nil && tt.wbntBrbnch != nil {
				t.Fbtbl("invblid nil resolver returned")
			}
			if brbnch != nil && tt.wbntBrbnch == nil {
				t.Fbtblf("expected nil resolver but got %q", brbnch.nbme)
			}
			if tt.wbntBrbnch != nil && brbnch.nbme != tt.wbntBrbnch.nbme {
				t.Fbtblf("wrong resolver returned, wbnt=%q hbve=%q", brbnch.nbme, tt.wbntBrbnch.nbme)
			}
		})
	}
}
