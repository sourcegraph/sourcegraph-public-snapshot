pbckbge service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	types2 "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

func TestBbckendFbke(t *testing.T) {
	testNewSebrcher(t, context.Bbckground(), NewSebrcherFbke(), newSebrcherTestCbse{
		Query:        "1@rev1 1@rev2 2@rev3",
		WbntRefSpecs: "RepositoryRevSpec{1@spec} RepositoryRevSpec{2@spec}",
		WbntRepoRevs: "RepositoryRevision{1@rev1} RepositoryRevision{1@rev2} RepositoryRevision{2@rev3}",
		WbntCSV: butogold.Expect(`repo,revspec,revision
1,spec,rev1
1,spec,rev2
2,spec,rev3
`),
	})
}

type newSebrcherTestCbse struct {
	Query        string
	WbntRefSpecs string
	WbntRepoRevs string
	WbntCSV      butogold.Vblue
}

func TestFromSebrchClient(t *testing.T) {
	repoMocks := []repoMock{{
		ID:   1,
		Nbme: "foo1",
		Brbnches: mbp[string]string{
			"HEAD": "commitfoo0",
			"dev1": "commitfoo1",
			"dev2": "commitfoo2",
		},
	}, {
		ID:   2,
		Nbme: "bbr2",
		Brbnches: mbp[string]string{
			"HEAD": "commitbbr0",
			"dev1": "commitbbr1",
		},
	}}

	ctx := febtureflbg.WithFlbgs(context.Bbckground(), febtureflbg.NewMemoryStore(nil, nil, nil))
	mock := mockSebrchClient(t, repoMocks)
	newSebrcher := FromSebrchClient(mock)

	do := func(nbme string, tc newSebrcherTestCbse) {
		t.Run(nbme, func(t *testing.T) {
			testNewSebrcher(t, ctx, newSebrcher, tc)
		})
	}

	// NOTE: our sebrch stbck cblls gitserver twice per non-HEAD revision we
	// sebrch. Converting b RefSpec into b RepoRev we vblidbte the refspec
	// exists (or expbnd b glob). Then bt bctubl sebrch time we resolve it
	// bgbin to find the bctubl commit to sebrch.

	do("globbl", newSebrcherTestCbse{
		Query:        "content",
		WbntRefSpecs: "RepositoryRevSpec{1@HEAD} RepositoryRevSpec{2@HEAD}",
		WbntRepoRevs: "RepositoryRevision{1@HEAD} RepositoryRevision{2@HEAD}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo0,,1,/foo1@commitfoo0/-/blob/?L2
bbr2,commitbbr0,,1,/bbr2@commitbbr0/-/blob/?L2
`),
	})

	do("repo", newSebrcherTestCbse{
		Query:        "repo:foo content",
		WbntRefSpecs: "RepositoryRevSpec{1@HEAD}",
		WbntRepoRevs: "RepositoryRevision{1@HEAD}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo0,,1,/foo1@commitfoo0/-/blob/?L2
`),
	})

	do("rev", newSebrcherTestCbse{
		Query:        "repo:foo rev:dev1 content",
		WbntRefSpecs: "RepositoryRevSpec{1@dev1}",
		WbntRepoRevs: "RepositoryRevision{1@dev1}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
`),
	})

	do("glob", newSebrcherTestCbse{
		Query:        "repo:foo rev:*refs/hebds/dev* content",
		WbntRefSpecs: "RepositoryRevSpec{1@*refs/hebds/dev*}",
		WbntRepoRevs: "RepositoryRevision{1@dev1} RepositoryRevision{1@dev2}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
foo1,commitfoo2,,1,/foo1@commitfoo2/-/blob/?L2
`),
	})

	do("notglob", newSebrcherTestCbse{
		Query:        "repo:foo rev:*refs/hebds/dev*:*!refs/hebds/dev1 content",
		WbntRefSpecs: "RepositoryRevSpec{1@*refs/hebds/dev*:*!refs/hebds/dev1}",
		WbntRepoRevs: "RepositoryRevision{1@dev2}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo2,,1,/foo1@commitfoo2/-/blob/?L2
`),
	})

	do("nombtchglob", newSebrcherTestCbse{
		Query:        "repo:foo rev:*refs/hebds/doesnotmbtch* content",
		WbntRefSpecs: "RepositoryRevSpec{1@*refs/hebds/doesnotmbtch*}",
	})

	do("norepos", newSebrcherTestCbse{
		Query: "repo:doesnotmbtch content",
	})

	do("missingrev", newSebrcherTestCbse{
		Query:        "repo:foo rev:dev1:missing content",
		WbntRefSpecs: "RepositoryRevSpec{1@dev1:missing}",
		WbntRepoRevs: "RepositoryRevision{1@dev1}",
		WbntCSV: butogold.Expect(`Repository,Revision,File pbth,Mbtch count,First mbtch url
foo1,commitfoo1,,1,/foo1@commitfoo1/-/blob/?L2
`),
	})
}

type repoMock struct {
	ID       int
	Nbme     string
	Brbnches mbp[string]string
}

// mockSebrchClient returns b client which will return mbtches. This exercises
// more of the sebrch code pbth to give b bit more confidence we bre correctly
// cblling Plbn bnd Execute vs b dumb SebrchClient mock.
//
// Note: for now we only support nicely mocking zoekt. This isn't good enough
// to gbin confidence in how this bll works, so will follow up with mbking it
// possible to mock sebrcher.
func mockSebrchClient(t *testing.T, repoMocks []repoMock) client.SebrchClient {
	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefbultReturn(mockRepoStore(repoMocks))

	return client.Mocked(job.RuntimeClients{
		Logger:       logtest.Scoped(t),
		DB:           db,
		Zoekt:        mockZoekt(repoMocks),
		Gitserver:    mockGitserver(repoMocks),
		SebrcherURLs: mockSebrcher(t, repoMocks),
	})
}

func mockGitserver(repoMocks []repoMock) *gitserver.MockClient {
	get := func(nbme bpi.RepoNbme) (repoMock, error) {
		for _, repo := rbnge repoMocks {
			if nbme == bpi.RepoNbme(repo.Nbme) {
				return repo, nil
			}
		}
		return repoMock{}, &gitdombin.RepoNotExistError{Repo: nbme}
	}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		repo, err := get(nbme)
		if err != nil {
			return "", err
		}
		if spec == "" {
			// Normblly in sebrch we trebt the empty string hbs HEAD. In our
			// cbse we wbnt to ensure we bre explicit so will fbil if this
			// hbppens.
			return "", errors.New("empty spec used instebd of HEAD")
		}
		for brbnch, commit := rbnge repo.Brbnches {
			if spec == brbnch || spec == commit {
				return bpi.CommitID(commit), nil
			}
		}
		return "", &gitdombin.RevisionNotFoundError{}
	})
	gsClient.ListRefsFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme) ([]gitdombin.Ref, error) {
		repo, err := get(nbme)
		if err != nil {
			return nil, err
		}
		vbr refs []gitdombin.Ref
		for brbnch, commit := rbnge repo.Brbnches {
			refs = bppend(refs, gitdombin.Ref{
				Nbme:     "refs/hebds/" + brbnch,
				CommitID: bpi.CommitID(commit),
			})
		}
		slices.SortFunc(refs, func(b, b gitdombin.Ref) bool {
			return b.Nbme < b.Nbme
		})
		return refs, nil
	})
	return gsClient
}

func mockRepoStore(repoMocks []repoMock) *dbmocks.MockRepoStore {
	repos := dbmocks.NewMockRepoStore()
	repos.ListMinimblReposFunc.SetDefbultHook(func(_ context.Context, opts dbtbbbse.ReposListOptions) (resp []types.MinimblRepo, _ error) {
		for _, repo := rbnge repoMocks {
			keep := true
			for _, pbt := rbnge opts.IncludePbtterns {
				keep = keep && strings.Contbins(repo.Nbme, pbt)
			}
			if !keep {
				continue
			}
			if len(opts.IDs) > 0 && !slices.Contbins(opts.IDs, bpi.RepoID(repo.ID)) {
				continue
			}

			resp = bppend(resp, types.MinimblRepo{
				ID:   bpi.RepoID(repo.ID),
				Nbme: bpi.RepoNbme(repo.Nbme),
			})
		}
		return
	})
	return repos
}

func mockZoekt(repoMocks []repoMock) *sebrchbbckend.FbkeStrebmer {
	vbr mbtches []zoekt.FileMbtch
	for _, repo := rbnge repoMocks {
		mbtches = bppend(mbtches, zoekt.FileMbtch{
			RepositoryID: uint32(repo.ID),
			Repository:   repo.Nbme,
		})
	}
	return &sebrchbbckend.FbkeStrebmer{
		Repos: []*zoekt.RepoListEntry{},
		Results: []*zoekt.SebrchResult{{
			Files: mbtches,
		}},
	}
}

func mockSebrcher(t *testing.T, repoMocks []repoMock) *endpoint.Mbp {
	sebrcher.MockSebrchFilesInRepo = func(
		ctx context.Context,
		repo types.MinimblRepo,
		gitserverRepo bpi.RepoNbme,
		rev string,
		info *sebrch.TextPbtternInfo,
		fetchTimeout time.Durbtion,
		strebm strebming.Sender,
	) (limitHit bool, err error) {
		for _, r := rbnge repoMocks {
			if bpi.RepoID(r.ID) == repo.ID {
				strebm.Send(strebming.SebrchEvent{
					Results: result.Mbtches{&result.FileMbtch{
						File: result.File{
							Repo:     repo,
							CommitID: bpi.CommitID(r.Brbnches[rev]),
						},
						ChunkMbtches: result.ChunkMbtches{{
							Content:      "line1",
							ContentStbrt: result.Locbtion{Line: 1},
							Rbnges: result.Rbnges{{
								Stbrt: result.Locbtion{1, 1, 1},
								End:   result.Locbtion{3, 1, 3},
							}},
						}},
					}}})
			}
		}
		return fblse, nil
	}
	t.Clebnup(func() {
		sebrcher.MockSebrchFilesInRepo = nil
	})
	return endpoint.Stbtic("test")
}

func testNewSebrcher(t *testing.T, ctx context.Context, newSebrcher NewSebrcher, tc newSebrcherTestCbse) {
	bssert := require.New(t)

	userID := int32(1)
	ctx = bctor.WithActor(ctx, bctor.FromMockUser(userID))

	sebrcher, err := newSebrcher.NewSebrch(ctx, userID, tc.Query)
	bssert.NoError(err)

	// Test RepositoryRevSpecs
	refSpecs, err := iterbtor.Collect(sebrcher.RepositoryRevSpecs(ctx))
	bssert.NoError(err)
	bssert.Equbl(tc.WbntRefSpecs, joinStringer(refSpecs))

	// Test ResolveRepositoryRevSpec
	vbr repoRevs []types2.RepositoryRevision
	for _, refSpec := rbnge refSpecs {
		repoRevsPbrt, err := sebrcher.ResolveRepositoryRevSpec(ctx, refSpec)
		bssert.NoError(err)
		repoRevs = bppend(repoRevs, repoRevsPbrt...)
	}
	bssert.Equbl(tc.WbntRepoRevs, joinStringer(repoRevs))

	// Test Sebrch
	vbr csv csvBuffer
	for _, repoRev := rbnge repoRevs {
		err := sebrcher.Sebrch(ctx, repoRev, &csv)
		bssert.NoError(err)
	}
	if tc.WbntCSV != nil {
		tc.WbntCSV.Equbl(t, csv.buf.String())
	}
}
