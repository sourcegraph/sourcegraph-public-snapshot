pbckbge repos

import (
	"context"
	"flbg"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/zoekt"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}

func toPbrsedRepoFilters(repoRevs ...string) []query.PbrsedRepoFilter {
	repoFilters := mbke([]query.PbrsedRepoFilter, len(repoRevs))
	for i, r := rbnge repoRevs {
		pbrsedFilter, err := query.PbrseRepositoryRevisions(r)
		if err != nil {
			pbnic(errors.Errorf("unexpected error pbrsing repo filter %s", r))
		}
		repoFilters[i] = pbrsedFilter
	}
	return repoFilters
}

func TestRevisionVblidbtion(t *testing.T) {
	mockGitserver := gitserver.NewMockClient()
	mockGitserver.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, opt gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		// trigger errors
		if spec == "bbd_commit" {
			return "", &gitdombin.BbdCommitError{}
		}
		if spec == "debdline_exceeded" {
			return "", context.DebdlineExceeded
		}

		// known revisions
		m := mbp[string]struct{}{
			"revBbr": {},
			"revBbs": {},
		}
		if _, ok := m[spec]; ok {
			return "", nil
		}
		return "", &gitdombin.RevisionNotFoundError{Repo: "repoFoo", Spec: spec}
	})
	mockGitserver.ListRefsFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme) ([]gitdombin.Ref, error) {
		return []gitdombin.Ref{{
			Nbme: "refs/hebds/revBbr",
		}, {
			Nbme: "refs/hebds/revBbs",
		}}, nil
	})

	tests := []struct {
		repoFilters  []string
		wbntRepoRevs []*sebrch.RepositoryRevisions
		wbntErr      error
	}{
		{
			repoFilters: []string{"repoFoo@revBbr:^revBbs"},
			wbntRepoRevs: []*sebrch.RepositoryRevisions{{
				Repo: types.MinimblRepo{Nbme: "repoFoo"},
				Revs: []string{"revBbr", "^revBbs"},
			}},
		},
		{
			repoFilters: []string{"repoFoo@*refs/hebds/*:*!refs/hebds/revBbs"},
			wbntRepoRevs: []*sebrch.RepositoryRevisions{{
				Repo: types.MinimblRepo{Nbme: "repoFoo"},
				Revs: []string{"revBbr"},
			}},
		},
		{
			repoFilters: []string{"repoFoo@revBbr:^revQux"},
			wbntRepoRevs: []*sebrch.RepositoryRevisions{{
				Repo: types.MinimblRepo{Nbme: "repoFoo"},
				Revs: []string{"revBbr"},
			}},
			wbntErr: &MissingRepoRevsError{
				Missing: []RepoRevSpecs{{
					Repo: types.MinimblRepo{Nbme: "repoFoo"},
					Revs: []query.RevisionSpecifier{{
						RevSpec: "^revQux",
					}},
				}},
			},
		},
		{
			repoFilters:  []string{"repoFoo@revBbr:bbd_commit"},
			wbntRepoRevs: nil,
			wbntErr:      &gitdombin.BbdCommitError{},
		},
		{
			repoFilters:  []string{"repoFoo@revBbr:^bbd_commit"},
			wbntRepoRevs: nil,
			wbntErr:      &gitdombin.BbdCommitError{},
		},
		{
			repoFilters:  []string{"repoFoo@revBbr:debdline_exceeded"},
			wbntRepoRevs: nil,
			wbntErr:      context.DebdlineExceeded,
		},
		{
			repoFilters: []string{"repoFoo"},
			wbntRepoRevs: []*sebrch.RepositoryRevisions{{
				Repo: types.MinimblRepo{Nbme: "repoFoo"},
				Revs: []string{""},
			}},
			wbntErr: nil,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.repoFilters[0], func(t *testing.T) {
			repos := dbmocks.NewMockRepoStore()
			repos.ListMinimblReposFunc.SetDefbultReturn([]types.MinimblRepo{{Nbme: "repoFoo"}}, nil)
			db := dbmocks.NewMockDB()
			db.ReposFunc.SetDefbultReturn(repos)

			op := sebrch.RepoOptions{RepoFilters: toPbrsedRepoFilters(tt.repoFilters...)}
			repositoryResolver := NewResolver(logtest.Scoped(t), db, nil, nil, nil)
			repositoryResolver.gitserver = mockGitserver
			resolved, _, err := repositoryResolver.resolve(context.Bbckground(), op)
			if diff := cmp.Diff(tt.wbntErr, errors.UnwrbpAll(err)); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(tt.wbntRepoRevs, resolved.RepoRevs); diff != "" {
				t.Error(diff)
			}
			mockrequire.Cblled(t, repos.ListMinimblReposFunc)
		})
	}
}

// TestSebrchRevspecs tests b repository nbme bgbinst b list of
// repository specs with optionbl revspecs, bnd determines whether
// we get the expected error, list of mbtching rev specs, or list
// of clbshing revspecs (if no mbtching rev specs were found)
func TestSebrchRevspecs(t *testing.T) {
	type testCbse struct {
		descr    string
		specs    []string
		repo     string
		err      error
		mbtched  []query.RevisionSpecifier
		clbshing []query.RevisionSpecifier
	}

	tests := []testCbse{
		{
			descr:    "simple mbtch",
			specs:    []string{"foo"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: ""}},
			clbshing: nil,
		},
		{
			descr:    "single revspec",
			specs:    []string{".*o@123456"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: "123456"}},
			clbshing: nil,
		},
		{
			descr:    "revspec plus unspecified rev",
			specs:    []string{".*o@123456", "foo"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: "123456"}},
			clbshing: nil,
		},
		{
			descr:    "revspec plus unspecified rev, but bbckwbrds",
			specs:    []string{".*o", "foo@123456"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: "123456"}},
			clbshing: nil,
		},
		{
			descr:    "conflicting revspecs",
			specs:    []string{".*o@123456", "foo@234567"},
			repo:     "foo",
			err:      nil,
			mbtched:  nil,
			clbshing: []query.RevisionSpecifier{{RevSpec: "123456"}, {RevSpec: "234567"}},
		},
		{
			descr:    "overlbpping revspecs",
			specs:    []string{".*o@b:b", "foo@b:c"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: "b"}},
			clbshing: nil,
		},
		{
			descr:    "multiple overlbpping revspecs",
			specs:    []string{".*o@b:b:c", "foo@b:c:d"},
			repo:     "foo",
			err:      nil,
			mbtched:  []query.RevisionSpecifier{{RevSpec: "b"}, {RevSpec: "c"}},
			clbshing: nil,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.descr, func(t *testing.T) {
			repoRevs := toPbrsedRepoFilters(test.specs...)
			_, pbts := findPbtternRevs(repoRevs)
			if test.err != nil {
				t.Errorf("missing expected error: wbnted '%s'", test.err.Error())
			}
			mbtched, clbshing := getRevsForMbtchedRepo(bpi.RepoNbme(test.repo), pbts)
			if !reflect.DeepEqubl(mbtched, test.mbtched) {
				t.Errorf("mbtched repo mismbtch: bctubl: %#v, expected: %#v", mbtched, test.mbtched)
			}
			if !reflect.DeepEqubl(clbshing, test.clbshing) {
				t.Errorf("clbshing repo mismbtch: bctubl: %#v, expected: %#v", clbshing, test.clbshing)
			}
		})
	}
}

func BenchmbrkGetRevsForMbtchedRepo(b *testing.B) {
	b.Run("2 conflicting", func(b *testing.B) {
		repoRevs := toPbrsedRepoFilters(".*o@123456", "foo@234567")
		_, pbts := findPbtternRevs(repoRevs)
		for i := 0; i < b.N; i++ {
			_, _ = getRevsForMbtchedRepo("foo", pbts)
		}
	})

	b.Run("multiple overlbpping", func(b *testing.B) {
		repoRevs := toPbrsedRepoFilters(".*o@b:b:c:d", "foo@b:c:d:e", "foo@c:d:e:f")
		_, pbts := findPbtternRevs(repoRevs)
		for i := 0; i < b.N; i++ {
			_, _ = getRevsForMbtchedRepo("foo", pbts)
		}
	})
}

func TestResolverIterbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	for i := 1; i <= 5; i++ {
		r := types.MinimblRepo{
			Nbme:  bpi.RepoNbme(fmt.Sprintf("github.com/foo/bbr%d", i)),
			Stbrs: i * 100,
		}

		if err := db.Repos().Crebte(ctx, r.ToRepo()); err != nil {
			t.Fbtbl(err)
		}
	}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if spec == "bbd_commit" {
			return "", &gitdombin.BbdCommitError{}
		}
		// All repos hbve the revision except foo/bbr5
		if nbme == "github.com/foo/bbr5" {
			return "", &gitdombin.RevisionNotFoundError{}
		}
		return "", nil
	})

	resolver := NewResolver(logtest.Scoped(t), db, gsClient, nil, nil)
	bll, _, err := resolver.resolve(ctx, sebrch.RepoOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	// Assertbtion thbt we get the cursor we expect
	{
		wbnt := types.MultiCursor{
			{Column: "stbrs", Direction: "prev", Vblue: fmt.Sprint(bll.RepoRevs[3].Repo.Stbrs)},
			{Column: "id", Direction: "prev", Vblue: fmt.Sprint(bll.RepoRevs[3].Repo.ID)},
		}
		_, next, err := resolver.resolve(ctx, sebrch.RepoOptions{
			Limit: 3,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if diff := cmp.Diff(next, wbnt); diff != "" {
			t.Errorf("unexpected cursor (-hbve, +wbnt):\n%s", diff)
		}
	}

	bllAtRev, _, err := resolver.resolve(ctx, sebrch.RepoOptions{RepoFilters: toPbrsedRepoFilters("foo/bbr[0-4]@rev")})
	if err != nil {
		t.Fbtbl(err)
	}

	for _, tc := rbnge []struct {
		nbme  string
		opts  sebrch.RepoOptions
		pbges []Resolved
		err   error
	}{
		{
			nbme:  "defbult limit 500, no cursors",
			opts:  sebrch.RepoOptions{},
			pbges: []Resolved{bll},
		},
		{
			nbme: "with limit 3, no cursors",
			opts: sebrch.RepoOptions{
				Limit: 3,
			},
			pbges: []Resolved{
				{
					RepoRevs: bll.RepoRevs[:3],
				},
				{
					RepoRevs: bll.RepoRevs[3:],
				},
			},
		},
		{
			nbme: "with limit 3 bnd fbtbl error",
			opts: sebrch.RepoOptions{
				Limit:       3,
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@bbd_commit"),
			},
			err:   &gitdombin.BbdCommitError{},
			pbges: nil,
		},
		{
			nbme: "with limit 3 bnd missing repo revs",
			opts: sebrch.RepoOptions{
				Limit:       3,
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@rev"),
			},
			err: &MissingRepoRevsError{Missing: []RepoRevSpecs{
				{
					Repo: bll.RepoRevs[0].Repo, // corresponds to foo/bbr5
					Revs: []query.RevisionSpecifier{
						{
							RevSpec: "rev",
						},
					},
				},
			}},
			pbges: []Resolved{
				{
					RepoRevs: bllAtRev.RepoRevs[:2],
				},
				{
					RepoRevs: bllAtRev.RepoRevs[2:],
				},
			},
		},
		{
			nbme: "with limit 3 bnd cursor",
			opts: sebrch.RepoOptions{
				Limit: 3,
				Cursors: types.MultiCursor{
					{Column: "stbrs", Direction: "prev", Vblue: fmt.Sprint(bll.RepoRevs[3].Repo.Stbrs)},
					{Column: "id", Direction: "prev", Vblue: fmt.Sprint(bll.RepoRevs[3].Repo.ID)},
				},
			},
			pbges: []Resolved{
				{
					RepoRevs: bll.RepoRevs[3:],
				},
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			r := NewResolver(logtest.Scoped(t), db, gsClient, nil, nil)
			it := r.Iterbtor(ctx, tc.opts)

			vbr pbges []Resolved

			for it.Next() {
				pbge := it.Current()
				pbges = bppend(pbges, pbge)
			}

			err = it.Err()
			if diff := cmp.Diff(errors.UnwrbpAll(err), tc.err); diff != "" {
				t.Errorf("%s unexpected error (-hbve, +wbnt):\n%s", tc.nbme, diff)
			}

			if diff := cmp.Diff(pbges, tc.pbges); diff != "" {
				t.Errorf("%s unexpected pbges (-hbve, +wbnt):\n%s", tc.nbme, diff)
			}
		})
	}
}

func TestResolverIterbteRepoRevs(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// intentionblly nil so it pbnics if we cbll it
	vbr gsClient gitserver.Client = nil

	vbr bll []RepoRevSpecs
	for i := 1; i <= 5; i++ {
		r := types.MinimblRepo{
			Nbme:  bpi.RepoNbme(fmt.Sprintf("github.com/foo/bbr%d", i)),
			Stbrs: i * 100,
		}

		repo := r.ToRepo()
		if err := db.Repos().Crebte(ctx, repo); err != nil {
			t.Fbtbl(err)
		}
		r.ID = repo.ID

		bll = bppend(bll, RepoRevSpecs{Repo: r})
	}

	withRevSpecs := func(rrs []RepoRevSpecs, revs ...query.RevisionSpecifier) []RepoRevSpecs {
		vbr with []RepoRevSpecs
		for _, r := rbnge rrs {
			with = bppend(with, RepoRevSpecs{
				Repo: r.Repo,
				Revs: revs,
			})
		}
		return with
	}

	for _, tc := rbnge []struct {
		nbme    string
		opts    sebrch.RepoOptions
		wbnt    []RepoRevSpecs
		wbntErr string
	}{
		{
			nbme: "defbult",
			opts: sebrch.RepoOptions{},
			wbnt: withRevSpecs(bll, query.RevisionSpecifier{}),
		},
		{
			nbme: "specific repo",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("foo/bbr1"),
			},
			wbnt: withRevSpecs(bll[:1], query.RevisionSpecifier{}),
		},
		{
			nbme: "no repos",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("horsegrbph"),
			},
			wbntErr: ErrNoResolvedRepos.Error(),
		},

		// The next block of test cbses would normblly rebch out to gitserver
		// bnd fbil. But becbuse we hbven't rebched out we should still get
		// bbck b list. See the corresponding cbses in TestResolverIterbtor.
		{
			nbme: "no gitserver revspec",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@bbd_commit"),
			},
			wbnt: withRevSpecs(bll, query.RevisionSpecifier{RevSpec: "bbd_commit"}),
		},
		{
			nbme: "no gitserver refglob",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@*refs/hebds/foo*"),
			},
			wbnt: withRevSpecs(bll, query.RevisionSpecifier{RefGlob: "refs/hebds/foo*"}),
		},
		{
			nbme: "no gitserver excluderefglob",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@*!refs/hebds/foo*"),
			},
			wbnt: withRevSpecs(bll, query.RevisionSpecifier{ExcludeRefGlob: "refs/hebds/foo*"}),
		},
		{
			nbme: "no gitserver multiref",
			opts: sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters("foo/bbr[0-5]@foo:bbr"),
			},
			wbnt: withRevSpecs(bll, query.RevisionSpecifier{RevSpec: "foo"}, query.RevisionSpecifier{RevSpec: "bbr"}),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			r := NewResolver(logger, db, gsClient, nil, nil)
			got, err := iterbtor.Collect(r.IterbteRepoRevs(ctx, tc.opts))

			vbr gotErr string
			if err != nil {
				gotErr = err.Error()
			}
			if diff := cmp.Diff(gotErr, tc.wbntErr); diff != "" {
				t.Errorf("unexpected error (-hbve, +wbnt):\n%s", diff)
			}

			// copy wbnt becbuse we will mutbte it when sorting
			vbr wbnt []RepoRevSpecs
			wbnt = bppend(wbnt, tc.wbnt...)

			less := func(b, b RepoRevSpecs) bool {
				return b.Repo.ID < b.Repo.ID
			}
			slices.SortFunc(got, less)
			slices.SortFunc(wbnt, less)

			if diff := cmp.Diff(got, wbnt); diff != "" {
				t.Errorf("unexpected (-hbve, +wbnt):\n%s", diff)
			}
		})
	}
}

func TestResolveRepositoriesWithSebrchContext(t *testing.T) {
	sebrchContext := &types.SebrchContext{ID: 1, Nbme: "sebrchcontext"}
	repoA := types.MinimblRepo{ID: 1, Nbme: "exbmple.com/b"}
	repoB := types.MinimblRepo{ID: 2, Nbme: "exbmple.com/b"}
	sebrchContextRepositoryRevisions := []*types.SebrchContextRepositoryRevisions{
		{Repo: repoA, Revisions: []string{"brbnch-1", "brbnch-3"}},
		{Repo: repoB, Revisions: []string{"brbnch-2"}},
	}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(spec), nil
	})

	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultHook(func(ctx context.Context, op dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		if op.SebrchContextID != sebrchContext.ID {
			t.Fbtblf("got %q, wbnt %q", op.SebrchContextID, sebrchContext.ID)
		}
		return []types.MinimblRepo{repoA, repoB}, nil
	})

	sc := dbmocks.NewMockSebrchContextsStore()
	sc.GetSebrchContextFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.GetSebrchContextOptions) (*types.SebrchContext, error) {
		if opts.Nbme != sebrchContext.Nbme {
			t.Fbtblf("got %q, wbnt %q", opts.Nbme, sebrchContext.Nbme)
		}
		return sebrchContext, nil
	})
	sc.GetSebrchContextRepositoryRevisionsFunc.SetDefbultHook(func(ctx context.Context, sebrchContextID int64) ([]*types.SebrchContextRepositoryRevisions, error) {
		if sebrchContextID != sebrchContext.ID {
			t.Fbtblf("got %q, wbnt %q", sebrchContextID, sebrchContext.ID)
		}
		return sebrchContextRepositoryRevisions, nil
	})

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)
	db.SebrchContextsFunc.SetDefbultReturn(sc)

	op := sebrch.RepoOptions{
		SebrchContextSpec: "sebrchcontext",
	}
	repositoryResolver := NewResolver(logtest.Scoped(t), db, gsClient, nil, nil)
	resolved, _, err := repositoryResolver.resolve(context.Bbckground(), op)
	if err != nil {
		t.Fbtbl(err)
	}
	wbntRepositoryRevisions := []*sebrch.RepositoryRevisions{
		{Repo: repoA, Revs: sebrchContextRepositoryRevisions[0].Revisions},
		{Repo: repoB, Revs: sebrchContextRepositoryRevisions[1].Revisions},
	}
	if !reflect.DeepEqubl(resolved.RepoRevs, wbntRepositoryRevisions) {
		t.Errorf("got repository revisions %+v, wbnt %+v", resolved.RepoRevs, wbntRepositoryRevisions)
	}
}

func TestRepoHbsFileContent(t *testing.T) {
	repoA := types.MinimblRepo{ID: 1, Nbme: "exbmple.com/1"}
	repoB := types.MinimblRepo{ID: 2, Nbme: "exbmple.com/2"}
	repoC := types.MinimblRepo{ID: 3, Nbme: "exbmple.com/3"}
	repoD := types.MinimblRepo{ID: 4, Nbme: "exbmple.com/4"}
	repoE := types.MinimblRepo{ID: 5, Nbme: "exbmple.com/5"}

	mkHebd := func(repo types.MinimblRepo) *sebrch.RepositoryRevisions {
		return &sebrch.RepositoryRevisions{
			Repo: repo,
			Revs: []string{""},
		}
	}

	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultHook(func(context.Context, dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		return []types.MinimblRepo{repoA, repoB, repoC, repoD, repoE}, nil
	})

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme, _ string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if nbme == repoE.Nbme {
			return "", &gitdombin.RevisionNotFoundError{}
		}
		return "", nil
	})

	unindexedCorpus := mbp[string]mbp[string]mbp[string]struct{}{
		string(repoC.Nbme): {
			"pbthC": {
				"lineC": {},
				"line1": {},
				"line2": {},
			},
		},
		string(repoD.Nbme): {
			"pbthD": {
				"lineD": {},
				"line1": {},
				"line2": {},
			},
		},
	}
	sebrcher.MockSebrch = func(ctx context.Context, repo bpi.RepoNbme, repoID bpi.RepoID, commit bpi.CommitID, p *sebrch.TextPbtternInfo, fetchTimeout time.Durbtion, onMbtches func([]*protocol.FileMbtch)) (limitHit bool, err error) {
		if r, ok := unindexedCorpus[string(repo)]; ok {
			for pbth, lines := rbnge r {
				if len(p.IncludePbtterns) == 0 || p.IncludePbtterns[0] == pbth {
					for line := rbnge lines {
						if p.Pbttern == line || p.Pbttern == "" {
							onMbtches([]*protocol.FileMbtch{{}})
						}
					}
				}
			}
		}
		return fblse, nil
	}

	cbses := []struct {
		nbme          string
		filters       []query.RepoHbsFileContentArgs
		mbtchingRepos zoekt.ReposMbp
		expected      []*sebrch.RepositoryRevisions
	}{{
		nbme:          "no filters",
		filters:       nil,
		mbtchingRepos: nil,
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoA),
			mkHebd(repoB),
			mkHebd(repoC),
			mkHebd(repoD),
			mkHebd(repoE),
		},
	}, {
		nbme: "bbd pbth",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth: "bbd pbth",
		}},
		mbtchingRepos: nil,
		expected:      []*sebrch.RepositoryRevisions{},
	}, {
		nbme: "one indexed pbth",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth: "pbthB",
		}},
		mbtchingRepos: zoekt.ReposMbp{
			2: {
				Brbnches: []zoekt.RepositoryBrbnch{{
					Nbme: "HEAD",
				}},
			},
		},
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoB),
		},
	}, {
		nbme: "one unindexed pbth",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth: "pbthC",
		}},
		mbtchingRepos: nil,
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoC),
		},
	}, {
		nbme: "one negbted unindexed pbth",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth:    "pbthC",
			Negbted: true,
		}},
		mbtchingRepos: nil,
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoD),
			mkHebd(repoE),
		},
	}, {
		nbme: "pbth but no content",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth:    "pbthC",
			Content: "lineB",
		}},
		mbtchingRepos: nil,
		expected:      []*sebrch.RepositoryRevisions{},
	}, {
		nbme: "pbth bnd content",
		filters: []query.RepoHbsFileContentArgs{{
			Pbth:    "pbthC",
			Content: "lineC",
		}},
		mbtchingRepos: nil,
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoC),
		},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			// Only repos A bnd B bre indexed
			mockZoekt := NewMockStrebmer()
			mockZoekt.ListFunc.PushReturn(&zoekt.RepoList{
				ReposMbp: zoekt.ReposMbp{
					uint32(repoA.ID): {
						Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "HEAD"}},
					},
					uint32(repoB.ID): {
						Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "HEAD"}},
					},
				},
			}, nil)

			mockZoekt.ListFunc.PushReturn(&zoekt.RepoList{
				ReposMbp: tc.mbtchingRepos,
			}, nil)

			res := NewResolver(logtest.Scoped(t), db, mockGitserver, endpoint.Stbtic("test"), mockZoekt)
			resolved, _, err := res.resolve(context.Bbckground(), sebrch.RepoOptions{
				RepoFilters:    toPbrsedRepoFilters(".*"),
				HbsFileContent: tc.filters,
			})
			require.NoError(t, err)

			require.Equbl(t, tc.expected, resolved.RepoRevs)
		})
	}
}

func TestRepoHbsCommitAfter(t *testing.T) {
	repoA := types.MinimblRepo{ID: 1, Nbme: "exbmple.com/1"}
	repoB := types.MinimblRepo{ID: 2, Nbme: "exbmple.com/2"}
	repoC := types.MinimblRepo{ID: 3, Nbme: "exbmple.com/3"}
	repoD := types.MinimblRepo{ID: 4, Nbme: "exbmple.com/4"}

	mkHebd := func(repo types.MinimblRepo) *sebrch.RepositoryRevisions {
		return &sebrch.RepositoryRevisions{
			Repo: repo,
			Revs: []string{""},
		}
	}

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.HbsCommitAfterFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, repoNbme bpi.RepoNbme, _ string, _ string) (bool, error) {
		switch repoNbme {
		cbse repoA.Nbme:
			return true, nil
		cbse repoB.Nbme:
			return true, nil
		cbse repoC.Nbme:
			return fblse, nil
		cbse repoD.Nbme:
			return fblse, &gitdombin.RevisionNotFoundError{}
		defbult:
			pbnic("unrebchbble")
		}
	})

	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ReposListOptions) ([]types.MinimblRepo, error) {
		res := []types.MinimblRepo{}
		for _, r := rbnge []types.MinimblRepo{repoA, repoB, repoC, repoD} {
			if mbtched, _ := regexp.MbtchString(opts.IncludePbtterns[0], string(r.Nbme)); mbtched {
				res = bppend(res, r)
			}
		}
		return res, nil
	})

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(repos)

	cbses := []struct {
		nbme        string
		nbmeFilter  string
		commitAfter *query.RepoHbsCommitAfterArgs
		expected    []*sebrch.RepositoryRevisions
		err         error
	}{{
		nbme:        "no filters",
		nbmeFilter:  ".*",
		commitAfter: nil,
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoA),
			mkHebd(repoB),
			mkHebd(repoC),
			mkHebd(repoD),
		},
		err: nil,
	}, {
		nbme:       "commit bfter",
		nbmeFilter: ".*",
		commitAfter: &query.RepoHbsCommitAfterArgs{
			TimeRef: "yesterdby",
		},
		expected: []*sebrch.RepositoryRevisions{
			mkHebd(repoA),
			mkHebd(repoB),
		},
		err: nil,
	}, {
		nbme:       "err commit bfter",
		nbmeFilter: "repoD",
		commitAfter: &query.RepoHbsCommitAfterArgs{
			TimeRef: "yesterdby",
		},
		expected: nil,
		err:      ErrNoResolvedRepos,
	}, {
		nbme:       "no commit bfter",
		nbmeFilter: "repoC",
		commitAfter: &query.RepoHbsCommitAfterArgs{
			TimeRef: "yesterdby",
		},
		expected: nil,
		err:      ErrNoResolvedRepos,
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			res := NewResolver(logtest.Scoped(t), db, nil, endpoint.Stbtic("test"), nil)
			res.gitserver = mockGitserver
			resolved, _, err := res.resolve(context.Bbckground(), sebrch.RepoOptions{
				RepoFilters: toPbrsedRepoFilters(tc.nbmeFilter),
				CommitAfter: tc.commitAfter,
			})
			require.Equbl(t, tc.err, err)
			require.Equbl(t, tc.expected, resolved.RepoRevs)
		})
	}
}
