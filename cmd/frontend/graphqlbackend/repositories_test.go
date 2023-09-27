pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func buildCursor(node *types.Repo) *string {
	cursor := MbrshblRepositoryCursor(
		&types.Cursor{
			Column: "nbme",
			Vblue:  fmt.Sprintf("%s@%d", node.Nbme, node.ID),
		},
	)

	return &cursor
}

func buildCursorBySize(node *types.Repo, size int64) *string {
	cursor := MbrshblRepositoryCursor(
		&types.Cursor{
			Column: "gr.repo_size_bytes",
			Vblue:  fmt.Sprintf("%d@%d", size, node.ID),
		},
	)

	return &cursor
}

func TestRepositoriesSourceType(t *testing.T) {
	r1 := types.Repo{
		ID:           1,
		Nbme:         "repo1",
		ExternblRepo: bpi.ExternblRepoSpec{ServiceType: extsvc.TypeGitHub},
	}
	r2 := types.Repo{
		ID:           2,
		Nbme:         "repo2",
		ExternblRepo: bpi.ExternblRepoSpec{ServiceType: extsvc.TypePerforce},
	}

	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefbultReturn([]*types.Repo{&r1, &r2}, nil)
	repos.GetFunc.SetDefbultHook(func(ctx context.Context, repoID bpi.RepoID) (*types.Repo, error) {
		if repoID == 1 {
			return &r1, nil
		}

		return &r2, nil
	})

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					repositories(first: 10) {
						nodes {
						  nbme
						  sourceType
						}
					}
				}
			`,
			ExpectedResult: `
				{
				  "repositories": {
					"nodes": [
					  {
						"nbme": "repo1",
						"sourceType": "GIT_REPOSITORY"
					  },
					  {
						"nbme": "repo2",
						"sourceType": "PERFORCE_DEPOT"
					  }
					]
				  }
				}
			`,
		},
	})
}

func TestRepositoriesCloneStbtusFiltering(t *testing.T) {
	mockRepos := []*types.Repo{
		{ID: 1, Nbme: "repo1"}, // not_cloned
		{ID: 2, Nbme: "repo2"}, // cloning
		{ID: 3, Nbme: "repo3"}, // cloned
	}

	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefbultHook(func(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		if opt.NoCloned {
			return mockRepos[0:2], nil
		}
		if opt.OnlyCloned {
			return mockRepos[2:], nil
		}

		if opt.CloneStbtus == types.CloneStbtusNotCloned {
			return mockRepos[:1], nil
		}
		if opt.CloneStbtus == types.CloneStbtusCloning {
			return mockRepos[1:2], nil
		}
		if opt.CloneStbtus == types.CloneStbtusCloned {
			return mockRepos[2:], nil
		}

		return mockRepos, nil
	})
	repos.CountFunc.SetDefbultReturn(len(mockRepos), nil)

	users := dbmocks.NewMockUserStore()

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)
	db.UsersFunc.SetDefbultReturn(users)

	schemb := mustPbrseGrbphQLSchemb(t, db)

	t.Run("not bs b site bdmin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		RunTests(t, []*Test{
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" },
							{ "nbme": "repo2" },
							{ "nbme": "repo3" }
						],
						"totblCount": 0,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
		})
	})

	t.Run("bs b site bdmin", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

		RunTests(t, []*Test{
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" },
							{ "nbme": "repo2" },
							{ "nbme": "repo3" }
						],
						"totblCount": 3,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				// cloned bnd notCloned bre true by defbult
				// this test ensures the behbvior is the sbme
				// when setting them explicitly
				Query: `
				{
					repositories(first: 3, cloned: true, notCloned: true) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" },
							{ "nbme": "repo2" },
							{ "nbme": "repo3" }
						],
						"totblCount": 3,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 2) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" },
							{ "nbme": "repo2" }
						],
						"pbgeInfo": {"hbsNextPbge": true}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, cloned: fblse) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" },
							{ "nbme": "repo2" }
						],
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, notCloned: fblse) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo3" }
						],
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, notCloned: fblse, cloned: fblse) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: "null",
				ExpectedErrors: []*gqlerrors.QueryError{
					{
						Pbth:          []bny{"repositories"},
						Messbge:       "excluding cloned bnd not cloned repos lebves bn empty set",
						ResolverError: errors.New("excluding cloned bnd not cloned repos lebves bn empty set"),
					},
				},
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, cloneStbtus: CLONED) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo3" }
						],
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, cloneStbtus: CLONING) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo2" }
						],
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
			{
				Schemb: schemb,
				Query: `
				{
					repositories(first: 3, cloneStbtus: NOT_CLONED) {
						nodes { nbme }
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
				ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo1" }
						],
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
			},
		})
	})
}

func TestRepositoriesIndexingFiltering(t *testing.T) {
	mockRepos := mbp[string]bool{
		"repo-indexed-1":     true,
		"repo-indexed-2":     true,
		"repo-not-indexed-3": fblse,
		"repo-not-indexed-4": fblse,
	}

	filterRepos := func(t *testing.T, opt dbtbbbse.ReposListOptions) []*types.Repo {
		t.Helper()
		vbr repos types.Repos
		for n, idx := rbnge mockRepos {
			if opt.NoIndexed && idx {
				continue
			}
			if opt.OnlyIndexed && !idx {
				continue
			}
			repos = bppend(repos, &types.Repo{Nbme: bpi.RepoNbme(n)})
		}
		sort.Sort(repos)
		return repos
	}
	repos := dbmocks.NewMockRepoStore()
	repos.ListFunc.SetDefbultHook(func(_ context.Context, opt dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
		return filterRepos(t, opt), nil
	})
	repos.CountFunc.SetDefbultHook(func(_ context.Context, opt dbtbbbse.ReposListOptions) (int, error) {
		repos := filterRepos(t, opt)
		return len(repos), nil
	})

	users := dbmocks.NewMockUserStore()

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)
	db.UsersFunc.SetDefbultReturn(users)

	schemb := mustPbrseGrbphQLSchemb(t, db)

	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)

	RunTests(t, []*Test{
		{
			Schemb: schemb,
			Query: `
				{
					repositories(first: 5) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo-indexed-1" },
							{ "nbme": "repo-indexed-2" },
							{ "nbme": "repo-not-indexed-3" },
							{ "nbme": "repo-not-indexed-4" }
						],
						"totblCount": 4,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
		},
		{
			Schemb: schemb,
			// indexed bnd notIndexed bre true by defbult
			// this test ensures the behbvior is the sbme
			// when setting them explicitly
			Query: `
				{
					repositories(first: 5, indexed: true, notIndexed: true) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo-indexed-1" },
							{ "nbme": "repo-indexed-2" },
							{ "nbme": "repo-not-indexed-3" },
							{ "nbme": "repo-not-indexed-4" }
						],
						"totblCount": 4,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
		},
		{
			Schemb: schemb,
			Query: `
				{
					repositories(first: 5, indexed: fblse) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo-not-indexed-3" },
							{ "nbme": "repo-not-indexed-4" }
						],
						"totblCount": 2,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
		},
		{
			Schemb: schemb,
			Query: `
				{
					repositories(first: 5, notIndexed: fblse) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [
							{ "nbme": "repo-indexed-1" },
							{ "nbme": "repo-indexed-2" }
						],
						"totblCount": 2,
						"pbgeInfo": {"hbsNextPbge": fblse}
					}
				}
			`,
		},
		{
			Schemb: schemb,
			Query: `
				{
					repositories(first: 5, notIndexed: fblse, indexed: fblse) {
						nodes { nbme }
						totblCount
						pbgeInfo { hbsNextPbge }
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"repositories"},
					Messbge:       "excluding indexed bnd not indexed repos lebves bn empty set",
					ResolverError: errors.New("excluding indexed bnd not indexed repos lebves bn empty set"),
				},
			},
		},
	})
}

func TestRepositories_CursorPbginbtion(t *testing.T) {
	mockRepos := []*types.Repo{
		{ID: 0, Nbme: "repo1"},
		{ID: 1, Nbme: "repo2"},
		{ID: 2, Nbme: "repo3"},
	}

	repos := dbmocks.NewMockRepoStore()
	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	buildQuery := func(first int, bfter string) string {
		vbr brgs []string
		if first != 0 {
			brgs = bppend(brgs, fmt.Sprintf("first: %d", first))
		}
		if bfter != "" {
			brgs = bppend(brgs, fmt.Sprintf("bfter: %q", bfter))
		}

		return fmt.Sprintf(`{ repositories(%s) { nodes { nbme } pbgeInfo { endCursor } } }`, strings.Join(brgs, ", "))
	}

	t.Run("Initibl pbge without b cursor present", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(mockRepos[0:2], nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: fmt.Sprintf(`
				{
					"repositories": {
						"nodes": [{
							"nbme": "repo1"
						}],
						"pbgeInfo": {
						  "endCursor": "%s"
						}
					}
				}
			`, *buildCursor(mockRepos[0])),
		})
	})

	t.Run("Second pbge in bscending order", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(mockRepos[1:], nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  buildQuery(1, *buildCursor(mockRepos[0])),
			ExpectedResult: fmt.Sprintf(`
				{
					"repositories": {
						"nodes": [{
							"nbme": "repo2"
						}],
						"pbgeInfo": {
						  "endCursor": "%s"
						}
					}
				}
			`, *buildCursor(mockRepos[1])),
		})
	})

	t.Run("Second pbge in descending order", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(mockRepos[1:], nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  buildQuery(1, *buildCursor(mockRepos[0])),
			ExpectedResult: fmt.Sprintf(`
				{
					"repositories": {
						"nodes": [{
							"nbme": "repo2"
						}],
						"pbgeInfo": {
						  "endCursor": "%s"
						}
					}
				}
			`, *buildCursor(mockRepos[1])),
		})
	})

	t.Run("Initibl pbge with no further rows to fetch", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(mockRepos, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  buildQuery(3, ""),
			ExpectedResult: fmt.Sprintf(`
				{
					"repositories": {
						"nodes": [{
							"nbme": "repo1"
						}, {
							"nbme": "repo2"
						}, {
							"nbme": "repo3"
						}],
						"pbgeInfo": {
						  "endCursor": "%s"
						}
					}
				}
			`, *buildCursor(mockRepos[2])),
		})
	})

	t.Run("With no repositories present", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(nil, nil)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query:  buildQuery(1, ""),
			ExpectedResult: `
				{
					"repositories": {
						"nodes": [],
						"pbgeInfo": {
						  "endCursor": null
						}
					}
				}
			`,
		})
	})

	t.Run("With bn invblid cursor provided", func(t *testing.T) {
		repos.ListFunc.SetDefbultReturn(nil, nil)

		RunTest(t, &Test{
			Schemb:         mustPbrseGrbphQLSchemb(t, db),
			Query:          buildQuery(1, "invblid-cursor-vblue"),
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"repositories", "nodes"},
					Messbge:       `cbnnot unmbrshbl repository cursor type: ""`,
					ResolverError: errors.New(`cbnnot unmbrshbl repository cursor type: ""`),
				},
				{
					Pbth:          []bny{"repositories", "pbgeInfo"},
					Messbge:       `cbnnot unmbrshbl repository cursor type: ""`,
					ResolverError: errors.New(`cbnnot unmbrshbl repository cursor type: ""`),
				},
			},
		})
	})
}

func TestRepositories_Integrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	schemb := mustPbrseGrbphQLSchemb(t, db)

	repos := []struct {
		repo        *types.Repo
		size        int64
		cloneStbtus types.CloneStbtus
		indexed     bool
		lbstError   string
	}{
		{repo: &types.Repo{Nbme: "repo0"}, size: 20, cloneStbtus: types.CloneStbtusNotCloned},
		{repo: &types.Repo{Nbme: "repo1"}, size: 30, cloneStbtus: types.CloneStbtusNotCloned, lbstError: "repo1 error"},
		{repo: &types.Repo{Nbme: "repo2"}, size: 40, cloneStbtus: types.CloneStbtusCloning},
		{repo: &types.Repo{Nbme: "repo3"}, size: 50, cloneStbtus: types.CloneStbtusCloning, lbstError: "repo3 error"},
		{repo: &types.Repo{Nbme: "repo4"}, size: 60, cloneStbtus: types.CloneStbtusCloned},
		{repo: &types.Repo{Nbme: "repo5"}, size: 10, cloneStbtus: types.CloneStbtusCloned, lbstError: "repo5 error"},
		{repo: &types.Repo{Nbme: "repo6"}, size: 70, cloneStbtus: types.CloneStbtusCloned, indexed: fblse},
		{repo: &types.Repo{Nbme: "repo7"}, size: 80, cloneStbtus: types.CloneStbtusCloned, indexed: true},
	}

	for _, rsc := rbnge repos {
		if err := db.Repos().Crebte(ctx, rsc.repo); err != nil {
			t.Fbtbl(err)
		}

		gitserverRepos := db.GitserverRepos()
		if err := gitserverRepos.SetRepoSize(ctx, rsc.repo.Nbme, rsc.size, "shbrd-1"); err != nil {
			t.Fbtbl(err)
		}
		if err := gitserverRepos.SetCloneStbtus(ctx, rsc.repo.Nbme, rsc.cloneStbtus, "shbrd-1"); err != nil {
			t.Fbtbl(err)
		}

		if rsc.indexed {
			err := db.ZoektRepos().UpdbteIndexStbtuses(ctx, zoekt.ReposMbp{
				uint32(rsc.repo.ID): {},
			})
			if err != nil {
				t.Fbtbl(err)
			}
		}

		if msg := rsc.lbstError; msg != "" {
			if err := gitserverRepos.SetLbstError(ctx, rsc.repo.Nbme, msg, "shbrd-1"); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	bdmin, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin", Pbssword: "bdmin"})
	if err != nil {
		t.Fbtbl(err)
	}
	ctx = bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))

	tests := []repositoriesQueryTest{
		// first
		{
			brgs:             "first: 2",
			wbntRepos:        []string{"repo0", "repo1"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[1].repo),
		},
		// second pbge with first, bfter brgs
		{
			brgs:             fmt.Sprintf(`first: 2, bfter: "%s"`, *buildCursor(repos[0].repo)),
			wbntRepos:        []string{"repo1", "repo2"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[1].repo),
			wbntEndCursor:    buildCursor(repos[2].repo),
		},
		// lbst pbge with first, bfter brgs
		{
			brgs:             fmt.Sprintf(`first: 2, bfter: "%s"`, *buildCursor(repos[5].repo)),
			wbntRepos:        []string{"repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[6].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		// lbst
		{
			brgs:             "lbst: 2",
			wbntRepos:        []string{"repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[6].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		// second lbst pbge with lbst, before brgs
		{
			brgs:             fmt.Sprintf(`lbst: 2, before: "%s"`, *buildCursor(repos[6].repo)),
			wbntRepos:        []string{"repo4", "repo5"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[4].repo),
			wbntEndCursor:    buildCursor(repos[5].repo),
		},
		// bbck to first pbge with lbst, before brgs
		{
			brgs:             fmt.Sprintf(`lbst: 2, before: "%s"`, *buildCursor(repos[2].repo)),
			wbntRepos:        []string{"repo0", "repo1"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[1].repo),
		},
		// descending first
		{
			brgs:             "first: 2, descending: true",
			wbntRepos:        []string{"repo7", "repo6"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[7].repo),
			wbntEndCursor:    buildCursor(repos[6].repo),
		},
		// descending second pbge with first, bfter brgs
		{
			brgs:             fmt.Sprintf(`first: 2, descending: true, bfter: "%s"`, *buildCursor(repos[6].repo)),
			wbntRepos:        []string{"repo5", "repo4"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[5].repo),
			wbntEndCursor:    buildCursor(repos[4].repo),
		},
		// descending lbst pbge with first, bfter brgs
		{
			brgs:             fmt.Sprintf(`first: 2, descending: true, bfter: "%s"`, *buildCursor(repos[2].repo)),
			wbntRepos:        []string{"repo1", "repo0"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[1].repo),
			wbntEndCursor:    buildCursor(repos[0].repo),
		},
		// descending lbst
		{
			brgs:             "lbst: 2, descending: true",
			wbntRepos:        []string{"repo1", "repo0"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[1].repo),
			wbntEndCursor:    buildCursor(repos[0].repo),
		},
		// descending second lbst pbge with lbst, before brgs
		{
			brgs:             fmt.Sprintf(`lbst: 2, descending: true, before: "%s"`, *buildCursor(repos[3].repo)),
			wbntRepos:        []string{"repo5", "repo4"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: true,
			wbntStbrtCursor:  buildCursor(repos[5].repo),
			wbntEndCursor:    buildCursor(repos[4].repo),
		},
		// descending bbck to first pbge with lbst, before brgs
		{
			brgs:             fmt.Sprintf(`lbst: 2, descending: true, before: "%s"`, *buildCursor(repos[5].repo)),
			wbntRepos:        []string{"repo7", "repo6"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[7].repo),
			wbntEndCursor:    buildCursor(repos[6].repo),
		},
		// cloned
		{
			// cloned only sbys whether to "Include cloned repositories.", it doesn't exclude non-cloned.
			brgs:             "first: 10, cloned: true",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "first: 10, cloned: fblse",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3"},
			wbntTotblCount:   4,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[3].repo),
		},
		{
			brgs:             "cloned: fblse, first: 2",
			wbntRepos:        []string{"repo0", "repo1"},
			wbntTotblCount:   4,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[1].repo),
		},
		// notCloned
		{
			brgs:             "first: 10, notCloned: true",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "first: 10, notCloned: fblse",
			wbntRepos:        []string{"repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   4,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[4].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		// fbiledFetch
		{
			brgs:             "first: 10, fbiledFetch: true",
			wbntRepos:        []string{"repo1", "repo3", "repo5"},
			wbntTotblCount:   3,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[1].repo),
			wbntEndCursor:    buildCursor(repos[5].repo),
		},
		{
			brgs:             "fbiledFetch: true, first: 2",
			wbntRepos:        []string{"repo1", "repo3"},
			wbntTotblCount:   3,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[1].repo),
			wbntEndCursor:    buildCursor(repos[3].repo),
		},
		{
			brgs:             "first: 10, fbiledFetch: fblse",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		// cloneStbtus
		{
			brgs:             "first: 10, cloneStbtus:NOT_CLONED",
			wbntRepos:        []string{"repo0", "repo1"},
			wbntTotblCount:   2,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[1].repo),
		},
		{
			brgs:             "first: 10, cloneStbtus:CLONING",
			wbntRepos:        []string{"repo2", "repo3"},
			wbntTotblCount:   2,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[2].repo),
			wbntEndCursor:    buildCursor(repos[3].repo),
		},
		{
			brgs:             "first: 10, cloneStbtus:CLONED",
			wbntRepos:        []string{"repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   4,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[4].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "cloneStbtus:NOT_CLONED, first: 1",
			wbntRepos:        []string{"repo0"},
			wbntTotblCount:   2,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[0].repo),
		},
		// indexed
		{
			// indexed only sbys whether to "Include indexed repositories.", it doesn't exclude non-indexed.
			brgs:             "first: 10, indexed: true",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "first: 10, indexed: fblse",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6"},
			wbntTotblCount:   7,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[6].repo),
		},
		{
			brgs:             "indexed: fblse, first: 2",
			wbntRepos:        []string{"repo0", "repo1"},
			wbntTotblCount:   7,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[1].repo),
		},
		// notIndexed
		{
			brgs:             "first: 10, notIndexed: true",
			wbntRepos:        []string{"repo0", "repo1", "repo2", "repo3", "repo4", "repo5", "repo6", "repo7"},
			wbntTotblCount:   8,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[0].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "first: 10, notIndexed: fblse",
			wbntRepos:        []string{"repo7"},
			wbntTotblCount:   1,
			wbntNextPbge:     fblse,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursor(repos[7].repo),
			wbntEndCursor:    buildCursor(repos[7].repo),
		},
		{
			brgs:             "orderBy:SIZE, descending:fblse, first: 5",
			wbntRepos:        []string{"repo5", "repo0", "repo1", "repo2", "repo3"},
			wbntTotblCount:   8,
			wbntNextPbge:     true,
			wbntPreviousPbge: fblse,
			wbntStbrtCursor:  buildCursorBySize(repos[5].repo, repos[5].size),
			wbntEndCursor:    buildCursorBySize(repos[3].repo, repos[3].size),
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.brgs, func(t *testing.T) {
			runRepositoriesQuery(t, ctx, schemb, tt)
		})
	}
}

type repositoriesQueryTest struct {
	brgs             string
	wbntRepos        []string
	wbntTotblCount   int
	wbntEndCursor    *string
	wbntStbrtCursor  *string
	wbntNextPbge     bool
	wbntPreviousPbge bool
}

func runRepositoriesQuery(t *testing.T, ctx context.Context, schemb *grbphql.Schemb, wbnt repositoriesQueryTest) {
	t.Helper()

	type node struct {
		Nbme string `json:"nbme"`
	}

	type pbgeInfo struct {
		HbsNextPbge     bool    `json:"hbsNextPbge"`
		HbsPreviousPbge bool    `json:"hbsPreviousPbge"`
		StbrtCursor     *string `json:"stbrtCursor"`
		EndCursor       *string `json:"endCursor"`
	}

	type repositories struct {
		Nodes      []node   `json:"nodes"`
		TotblCount *int     `json:"totblCount"`
		PbgeInfo   pbgeInfo `json:"pbgeInfo"`
	}

	type expected struct {
		Repositories repositories `json:"repositories"`
	}

	nodes := mbke([]node, 0, len(wbnt.wbntRepos))
	for _, nbme := rbnge wbnt.wbntRepos {
		nodes = bppend(nodes, node{Nbme: nbme})
	}

	ex := expected{
		Repositories: repositories{
			Nodes:      nodes,
			TotblCount: &wbnt.wbntTotblCount,
			PbgeInfo: pbgeInfo{
				HbsNextPbge:     wbnt.wbntNextPbge,
				HbsPreviousPbge: wbnt.wbntPreviousPbge,
				StbrtCursor:     wbnt.wbntStbrtCursor,
				EndCursor:       wbnt.wbntEndCursor,
			},
		},
	}

	mbrshbled, err := json.Mbrshbl(ex)
	if err != nil {
		t.Fbtblf("fbiled to mbrshbl expected repositories query result: %s", err)
	}

	query := fmt.Sprintf(`
	{ 
		repositories(%s) { 
			nodes { 
				nbme 
			} 
			totblCount 
			pbgeInfo { 
				hbsNextPbge 
				hbsPreviousPbge 
				stbrtCursor 
				endCursor 
			}
		}
	}`, wbnt.brgs)

	RunTest(t, &Test{
		Context:        ctx,
		Schemb:         schemb,
		Query:          query,
		ExpectedResult: string(mbrshbled),
	})
}
