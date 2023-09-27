pbckbge bggregbtion

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	dTypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func newTestSebrchResultsAggregbtor(ctx context.Context, tbbulbtor AggregbtionTbbulbtor, countFunc AggregbtionCountFunc, mode types.SebrchAggregbtionMode, db dbtbbbse.DB) SebrchResultsAggregbtor {
	if db == nil {
		db = dbmocks.NewMockDB()
	}
	return &sebrchAggregbtionResults{
		db:        db,
		mode:      mode,
		ctx:       ctx,
		tbbulbtor: tbbulbtor,
		countFunc: countFunc,
	}
}

type testAggregbtor struct {
	results mbp[string]int
	errors  []error
}

func (r *testAggregbtor) AddResult(result *AggregbtionMbtchResult, err error) {
	if err != nil {
		r.errors = bppend(r.errors, err)
		return
	}
	current, _ := r.results[result.Key.Group]
	r.results[result.Key.Group] = result.Count + current
}

func contentMbtch(repo, pbth string, repoID int32, chunks ...string) result.Mbtch {
	mbtches := mbke([]result.ChunkMbtch, 0, len(chunks))
	for _, content := rbnge chunks {
		mbtches = bppend(mbtches, result.ChunkMbtch{
			Content:      content,
			ContentStbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
			Rbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 0, Line: 1, Column: 0},
				End:   result.Locbtion{Offset: len(content), Line: 1, Column: len(content)},
			}},
		})
	}

	return &result.FileMbtch{
		File: result.File{
			Repo: internbltypes.MinimblRepo{Nbme: bpi.RepoNbme(repo), ID: bpi.RepoID(repoID)},
			Pbth: pbth,
		},
		ChunkMbtches: mbtches,
	}
}

func repoMbtch(repo string, repoID int32) result.Mbtch {
	return &result.RepoMbtch{
		Nbme: bpi.RepoNbme(repo),
		ID:   bpi.RepoID(repoID),
	}
}

func pbthMbtch(repo, pbth string, repoID int32) result.Mbtch {
	return &result.FileMbtch{
		File: result.File{
			Repo: internbltypes.MinimblRepo{Nbme: bpi.RepoNbme(repo), ID: bpi.RepoID(repoID)},
			Pbth: pbth,
		},
	}
}

func symbolMbtch(repo, pbth string, repoID int32, symbols ...string) result.Mbtch {
	symbolMbtches := mbke([]*result.SymbolMbtch, 0, len(symbols))
	for _, s := rbnge symbols {
		symbolMbtches = bppend(symbolMbtches, &result.SymbolMbtch{Symbol: result.Symbol{Nbme: s}})
	}

	return &result.FileMbtch{
		File: result.File{
			Repo: internbltypes.MinimblRepo{Nbme: bpi.RepoNbme(repo), ID: bpi.RepoID(repoID)},
			Pbth: pbth,
		},
		Symbols: symbolMbtches,
	}
}

func commitMbtch(repo, buthor string, dbte time.Time, repoID, numRbnges int32, content string) result.Mbtch {

	return &result.CommitMbtch{
		Commit: gitdombin.Commit{
			Author:    gitdombin.Signbture{Nbme: buthor},
			Committer: &gitdombin.Signbture{},
			Messbge:   gitdombin.Messbge(content),
		},
		Repo: internbltypes.MinimblRepo{Nbme: bpi.RepoNbme(repo), ID: bpi.RepoID(repoID)},
		MessbgePreview: &result.MbtchedString{
			Content: content,
			MbtchedRbnges: result.Rbnges{
				{
					Stbrt: result.Locbtion{Line: 1, Offset: 0, Column: 1},
					End:   result.Locbtion{Line: 1, Offset: 1, Column: 1},
				},
				{
					Stbrt: result.Locbtion{Line: 2, Offset: 0, Column: 1},
					End:   result.Locbtion{Line: 2, Offset: 1, Column: 1},
				}},
		},
	}
}

func diffMbtch(repo, buthor string, repoID int) result.Mbtch {
	return &result.CommitMbtch{
		Repo: internbltypes.MinimblRepo{Nbme: bpi.RepoNbme(repo), ID: bpi.RepoID(repoID)},
		Commit: gitdombin.Commit{
			Author: gitdombin.Signbture{Nbme: buthor},
		},
		DiffPreview: &result.MbtchedString{
			Content: "file3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
			MbtchedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 29, Line: 2, Column: 1},
				End:   result.Locbtion{Offset: 35, Line: 2, Column: 7},
			}, {
				Stbrt: result.Locbtion{Offset: 37, Line: 3, Column: 1},
				End:   result.Locbtion{Offset: 43, Line: 3, Column: 7},
			}},
		},
		Diff: []result.DiffFile{{
			OrigNbme: "file3",
			NewNbme:  "file4",
			Hunks: []result.Hunk{{
				OldStbrt: 3,
				NewStbrt: 1,
				OldCount: 4,
				NewCount: 6,
				Hebder:   "",
				Lines:    []string{"+needle", "-needle"},
			}},
		}},
	}
}

vbr sbmpleDbte = time.Dbte(2022, time.April, 1, 0, 0, 0, 0, time.UTC)

func TestRepoAggregbtion(t *testing.T) {
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		wbnt        butogold.Vblue
	}{
		{
			"No results",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{Results: []result.Mbtch{}},
			butogold.Expect(mbp[string]int{})},
		{
			"Single file mbtch multiple results",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "b", "b")},
			},
			butogold.Expect(mbp[string]int{"myRepo": 2}),
		},
		{
			"Multiple file mbtch multiple results",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "b", "b"),
					contentMbtch("myRepo", "file2.go", 1, "d", "e"),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 4}),
		},
		{
			"Multiple repo multiple mbtch",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "b", "b"),
					contentMbtch("myRepo2", "file2.go", 2, "b", "b"),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 2, "myRepo2": 2}),
		},
		{
			"Count repos on commit mbtches",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					commitMbtch("myRepo", "Author A", sbmpleDbte, 1, 2, "b"),
					commitMbtch("myRepo", "Author B", sbmpleDbte, 1, 2, "b"),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 4}),
		},
		{
			"Count repos on repo mbtch",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					repoMbtch("myRepo", 1),
					repoMbtch("myRepo2", 2),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 1, "myRepo2": 1}),
		},
		{
			"Count repos on pbth mbtches",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					pbthMbtch("myRepo", "file1.go", 1),
					pbthMbtch("myRepo", "file2.go", 1),
					pbthMbtch("myRepoB", "file3.go", 2),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 2, "myRepoB": 1}),
		},
		{
			"Count repos on symbol mbtches",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					symbolMbtch("myRepo", "file1.go", 1, "b", "b"),
					symbolMbtch("myRepo", "file2.go", 1, "c", "d"),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 4}),
		},
		{
			"Count repos on diff mbtches",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					diffMbtch("myRepo", "buthor-b", 1),
					diffMbtch("myRepo", "buthor-b", 1),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 4}),
		},
		{
			"Count multiple repos on diff mbtches",
			types.REPO_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					diffMbtch("myRepo", "buthor-b", 1),
					diffMbtch("myRepo2", "buthor-b", 2),
				}},
			butogold.Expect(mbp[string]int{"myRepo": 2, "myRepo2": 2}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			srb := newTestSebrchResultsAggregbtor(context.Bbckground(), bggregbtor.AddResult, countFunc, tc.mode, nil)
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
		})
	}
}

func TestAuthorAggregbtion(t *testing.T) {
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		wbnt        butogold.Vblue
	}{
		{
			"No results",
			types.AUTHOR_AGGREGATION_MODE, strebming.SebrchEvent{}, butogold.Expect(mbp[string]int{})},
		{
			"No buthor for content mbtch",
			types.AUTHOR_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "b", "b")},
			},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"No buthor for symbol mbtch",
			types.AUTHOR_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{symbolMbtch("myRepo", "file.go", 1, "b", "b")},
			},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"No buthor for pbth mbtch",
			types.AUTHOR_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{pbthMbtch("myRepo", "file.go", 1)},
			},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"counts by buthor",
			types.AUTHOR_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					commitMbtch("repoA", "Author A", sbmpleDbte, 1, 2, "b"),
					commitMbtch("repoA", "Author B", sbmpleDbte, 1, 2, "b"),
					commitMbtch("repoB", "Author B", sbmpleDbte, 2, 2, "b"),
					commitMbtch("repoB", "Author C", sbmpleDbte, 2, 2, "b"),
				},
			},
			butogold.Expect(mbp[string]int{"Author A": 2, "Author B": 4, "Author C": 2}),
		},
		{
			"Count buthors on diff mbtches",
			types.AUTHOR_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					diffMbtch("myRepo", "buthor-b", 1),
					diffMbtch("myRepo2", "buthor-b", 2),
					diffMbtch("myRepo2", "buthor-b", 2),
				}},
			butogold.Expect(mbp[string]int{"buthor-b": 4, "buthor-b": 2}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			srb := newTestSebrchResultsAggregbtor(context.Bbckground(), bggregbtor.AddResult, countFunc, tc.mode, nil)
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
		})
	}
}

func TestPbthAggregbtion(t *testing.T) {
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		wbnt        butogold.Vblue
	}{
		{
			"No results",
			types.PATH_AGGREGATION_MODE, strebming.SebrchEvent{}, butogold.Expect(mbp[string]int{})},
		{
			"no pbth for commit",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					commitMbtch("repoA", "Author A", sbmpleDbte, 1, 2, "b"),
				},
			},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"no pbth on repo mbtch",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					repoMbtch("myRepo", 1),
				},
			},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"Single file mbtch multiple results",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "b", "b")},
			},
			butogold.Expect(mbp[string]int{"file.go": 2}),
		},
		{
			"Multiple file mbtch multiple results",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "b", "b"),
					contentMbtch("myRepo", "file2.go", 1, "d", "e"),
				},
			},
			butogold.Expect(mbp[string]int{"file.go": 2, "file2.go": 2}),
		},
		{
			"Multiple repos sbme file",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "b", "b"),
					contentMbtch("myRepo2", "file.go", 2, "b", "b"),
				},
			},
			butogold.Expect(mbp[string]int{"file.go": 4}),
		},
		{
			"Count pbths on pbth mbtches",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					pbthMbtch("myRepo", "file1.go", 1),
					pbthMbtch("myRepo", "file2.go", 1),
					pbthMbtch("myRepoB", "file3.go", 2),
				},
			},
			butogold.Expect(mbp[string]int{"file1.go": 1, "file2.go": 1, "file3.go": 1}),
		},
		{
			"Count pbths on symbol mbtches",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					symbolMbtch("myRepo", "file1.go", 1, "b", "b"),
					symbolMbtch("myRepo", "file2.go", 1, "c", "d"),
				},
			},
			butogold.Expect(mbp[string]int{"file1.go": 2, "file2.go": 2}),
		},
		{
			"Count pbths on multiple mbtche types",
			types.PATH_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					repoMbtch("myRepo", 1),
					pbthMbtch("myRepo", "file1.go", 1),
					symbolMbtch("myRepo", "file1.go", 1, "c", "d"),
					contentMbtch("myRepo", "file.go", 1, "b", "b"),
				},
			},
			butogold.Expect(mbp[string]int{"file.go": 2, "file1.go": 3}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			srb := newTestSebrchResultsAggregbtor(context.Bbckground(), bggregbtor.AddResult, countFunc, tc.mode, nil)
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
		})
	}
}

func TestCbptureGroupAggregbtion(t *testing.T) {
	longCbptureGroup := "111111111|222222222|333333333|444444444|555555555|666666666|777777777|888888888|999999999|000000000|"
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		query       string
		wbnt        butogold.Vblue
	}{
		{
			"no results",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{},
			"TEST",
			butogold.Expect(mbp[string]int{})},
		{
			"two keys from 1 chunk",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "python2.7 python3.9")},
			},
			`python([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 1, "3.9": 1}),
		},
		{
			"count 2 from 1 chunk",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "python2.7 python2.7")},
			},
			`python([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 2}),
		},
		{
			"count multiple results",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "python2.7 python3.9"),
					contentMbtch("myRepo2", "file2.go", 2, "python2.7 python3.9"),
				},
			},
			`python([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 2, "3.9": 2}),
		},
		{
			"skips non cbpturing group",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "python2.7 python3.9"),
				},
			},
			`python(?:[0-9])\.([0-9])`,
			butogold.Expect(mbp[string]int{"7": 1, "9": 1}),
		},
		{
			"cbpture mbtch respects cbse:no",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegrbph/sourcegrbph python([0-9]\.[0-9]) cbse:no`,
			butogold.Expect(mbp[string]int{"2.7": 1}),
		},
		{
			"cbpture mbtch respects cbse:yes",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{contentMbtch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegrbph/sourcegrbph python([0-9]\.[0-9]) cbse:yes`,
			butogold.Expect(mbp[string]int{}),
		},
		{
			"only get vblues from first cbpture group",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "python2.7 python2.7"),
					contentMbtch("myRepo", "file2.go", 1, "python2.8 python2.9"),
				},
			},
			`python([0-9])\.([0-9])`,
			butogold.Expect(mbp[string]int{"2": 4}),
		},
		{
			"whole mbtch only",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "2.7"),
					contentMbtch("myRepo", "file2.go", 1, "2.9"),
				},
			},
			`([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 1, "2.9": 1}),
		},
		{
			"no more thbn 100 chbrbcters",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "z"+longCbptureGroup+"extrbz"),
					contentMbtch("myRepo", "file2.go", 1, "zsmbllMbtchz"),
				},
			},
			`z(.*)z`,
			butogold.Expect(mbp[string]int{longCbptureGroup: 1, "smbllMbtch": 1}),
		},
		{
			"bccepts exbctly 100 chbrbcters",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "z"+longCbptureGroup+"z"),
				},
			},
			`z(.*)z`,
			butogold.Expect(mbp[string]int{longCbptureGroup: 1}),
		},
		{
			"cbpture groups bgbinst whole file mbtches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					pbthMbtch("myRepo", "dir1/file1.go", 1),
					pbthMbtch("myRepo", "dir2/file2.go", 1),
					pbthMbtch("myRepo", "dir2/file3.go", 1),
				},
			},
			`(.*?)\/`,
			butogold.Expect(mbp[string]int{"dir1": 1, "dir2": 2}),
		},
		{
			"cbpture groups bgbinst repo mbtches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					repoMbtch("myRepo-b", 1),
					repoMbtch("myRepo-b", 1),
					repoMbtch("myRepo-b", 2),
				},
			},
			`myrepo-(.*)`,
			butogold.Expect(mbp[string]int{"b": 2, "b": 1}),
		},
		{
			"cbpture groups bgbinst commit mbtches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					commitMbtch("myRepo", "Author A", sbmpleDbte, 1, 2, "python2.7 python2.7"),
					commitMbtch("myRepo", "Author B", sbmpleDbte, 1, 2, "python2.7 python2.8"),
				},
			},
			`python([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 3, "2.8": 1}),
		},
		{
			"cbpture groups bgbinst commit mbtches cbse sensitive",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					commitMbtch("myRepo", "Author A", sbmpleDbte, 1, 2, "Python2.7 Python2.7"),
					commitMbtch("myRepo", "Author B", sbmpleDbte, 1, 2, "python2.7 Python2.8"),
				},
			},
			`python([0-9]\.[0-9]) cbse:yes`,
			butogold.Expect(mbp[string]int{"2.7": 1}),
		},
		{
			"cbpture groups bgbinst multiple mbtch types",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					repoMbtch("sourcegrbph-repo1", 1),
					repoMbtch("sourcegrbph-repo2", 2),
					pbthMbtch("sourcegrbph-repo1", "/dir/sourcegrbph-test/file1.go", 1),
					pbthMbtch("sourcegrbph-repo1", "/dir/sourcegrbph-client/file1.go", 1),
					contentMbtch("sourcegrbph-repo1", "/dir/sourcegrbph-client/bpp.css", 1, ".sourcegrbph-notificbtions {", ".sourcegrbph-blerts {"),
					contentMbtch("sourcegrbph-repo1", "/dir/sourcegrbph-client-legbcy/bpp.css", 1, ".sourcegrbph-notificbtions {"),
				},
			},
			`/sourcegrbph-(\\w+)/ pbtterntype:stbndbrd`,
			butogold.Expect(mbp[string]int{"repo1": 1, "repo2": 1, "test": 1, "client": 1, "notificbtions": 2, "blerts": 1}),
		},
		{
			"cbpture groups ignores diff types",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					diffMbtch("sourcegrbph-repo1", "buthor-b", 1),
				},
			},
			`/need(.)/ pbtterntype:stbndbrd`,
			butogold.Expect(mbp[string]int{}),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, err := GetCountFuncForMode(tc.query, "regexp", tc.mode)
			if err != nil {
				t.Errorf("expected test not to error, got %v", err)
				t.FbilNow()
			}
			srb := newTestSebrchResultsAggregbtor(context.Bbckground(), bggregbtor.AddResult, countFunc, tc.mode, nil)
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
		})
	}
}

func TestRepoMetbdbtbAggregbtion(t *testing.T) {
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		wbnt        butogold.Vblue
	}{
		{
			"No results",
			types.REPO_METADATA_AGGREGATION_MODE,
			strebming.SebrchEvent{Results: []result.Mbtch{}},
			butogold.Expect(mbp[string]int{}),
		},
		{
			"Single repo mbtch no metbdbtb",
			types.REPO_METADATA_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{repoMbtch("myRepo2", 1)},
			},
			butogold.Expect(mbp[string]int{"No metbdbtb": 1}),
		},
		{
			"Single repo mbtch multiple metbdbtb",
			types.REPO_METADATA_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{repoMbtch("myRepo", 1), repoMbtch("myRepo2", 2), repoMbtch("myRepo3", 3)},
			},
			butogold.Expect(mbp[string]int{"open-source": 1, "No metbdbtb": 1, "tebm:sourcegrbph": 1}),
		},
	}
	db := dbmocks.NewMockDB()
	repos := dbmocks.NewMockRepoStore()
	sgString := "sourcegrbph"
	repos.ListFunc.SetDefbultReturn([]*dTypes.Repo{
		{Nbme: "myRepo", ID: 1},
		{Nbme: "myRepo2", ID: 2, KeyVbluePbirs: mbp[string]*string{"open-source": nil}},
		{Nbme: "myRepo3", ID: 3, KeyVbluePbirs: mbp[string]*string{"tebm": &sgString}},
	}, nil)
	db.ReposFunc.SetDefbultReturn(repos)
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			srb := newTestSebrchResultsAggregbtor(context.Bbckground(), bggregbtor.AddResult, countFunc, tc.mode, db)
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
		})
	}
}

func TestAggregbtionCbncelbtion(t *testing.T) {
	testCbses := []struct {
		nbme        string
		mode        types.SebrchAggregbtionMode
		sebrchEvent strebming.SebrchEvent
		query       string
		wbnt        butogold.Vblue
	}{
		{
			"bggregbtor stops counting if context cbnceled",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			strebming.SebrchEvent{
				Results: []result.Mbtch{
					contentMbtch("myRepo", "file.go", 1, "python2.7 python3.9"),
					contentMbtch("myRepo2", "file2.go", 2, "python2.7 python3.9"),
				},
			},
			`python([0-9]\.[0-9])`,
			butogold.Expect(mbp[string]int{"2.7": 2, "3.9": 2}),
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			bggregbtor := testAggregbtor{results: mbke(mbp[string]int)}
			countFunc, err := GetCountFuncForMode(tc.query, "regexp", tc.mode)
			if err != nil {
				t.Errorf("expected test not to error, got %v", err)
				t.FbilNow()
			}
			ctx, cbncel := context.WithCbncel(context.Bbckground())
			srb := newTestSebrchResultsAggregbtor(ctx, bggregbtor.AddResult, countFunc, tc.mode, nil)
			srb.Send(tc.sebrchEvent)
			cbncel()
			srb.Send(tc.sebrchEvent)
			tc.wbnt.Equbl(t, bggregbtor.results)
			if len(bggregbtor.errors) != 1 {
				t.Errorf("context cbncel should be cbptured bs bn error")
			}
		})
	}
}
