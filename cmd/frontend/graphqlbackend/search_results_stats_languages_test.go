pbckbge grbphqlbbckend

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSebrchResultsStbtsLbngubges(t *testing.T) {
	logger := logtest.Scoped(t)
	wbntCommitID := bpi.CommitID(strings.Repebt("b", 40))
	rcbche.SetupForTest(t)

	gsClient := gitserver.NewMockClient()
	gsClient.NewFileRebderFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, commit bpi.CommitID, nbme string) (io.RebdCloser, error) {
		if commit != wbntCommitID {
			t.Errorf("got commit %q, wbnt %q", commit, wbntCommitID)
		}
		vbr dbtb []byte
		switch nbme {
		cbse "two.go":
			dbtb = []byte("b\nb\n")
		cbse "three.go":
			dbtb = []byte("b\nb\nc\n")
		defbult:
			pbnic("unhbndled mock NewFileRebder " + nbme)
		}
		return io.NopCloser(bytes.NewRebder(dbtb)), nil
	})
	const wbntDefbultBrbnchRef = "refs/hebds/foo"
	gsClient.GetDefbultBrbnchFunc.SetDefbultHook(func(context.Context, bpi.RepoNbme, bool) (string, bpi.CommitID, error) {
		// Mock defbult brbnch lookup in (*RepositoryResolver).DefbultBrbnch.
		return wbntDefbultBrbnchRef, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", nil
	})
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if wbnt := "HEAD"; spec != wbnt {
			t.Errorf("got spec %q, wbnt %q", spec, wbnt)
		}
		return wbntCommitID, nil
	})

	gsClient.StbtFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, _ bpi.CommitID, pbth string) (fs.FileInfo, error) {
		return &fileutil.FileInfo{Nbme_: pbth, Mode_: os.ModeDir}, nil
	})

	mkResult := func(pbth string, lineNumbers ...int) *result.FileMbtch {
		rn := types.MinimblRepo{
			Nbme: "r",
		}
		fm := mkFileMbtch(rn, pbth, lineNumbers...)
		fm.CommitID = wbntCommitID
		return fm
	}

	tests := mbp[string]struct {
		results  []result.Mbtch
		getFiles []fs.FileInfo
		wbnt     []inventory.Lbng // TotblBytes vblues bre incorrect (known issue doc'd in GrbphQL schemb)
	}{
		"empty": {
			results: nil,
			wbnt:    []inventory.Lbng{},
		},
		"1 entire file": {
			results: []result.Mbtch{
				mkResult("three.go"),
			},
			wbnt: []inventory.Lbng{{Nbme: "Go", TotblBytes: 6, TotblLines: 3}},
		},
		"line mbtches in 1 file": {
			results: []result.Mbtch{
				mkResult("three.go", 1),
			},
			wbnt: []inventory.Lbng{{Nbme: "Go", TotblBytes: 6, TotblLines: 1}},
		},
		"line mbtches in 2 files": {
			results: []result.Mbtch{
				mkResult("two.go", 1, 2),
				mkResult("three.go", 1),
			},
			wbnt: []inventory.Lbng{{Nbme: "Go", TotblBytes: 10, TotblLines: 3}},
		},
		"1 entire repo": {
			results: []result.Mbtch{
				&result.RepoMbtch{Nbme: "r"},
			},
			getFiles: []fs.FileInfo{
				fileInfo{pbth: "two.go"},
				fileInfo{pbth: "three.go"},
			},
			wbnt: []inventory.Lbng{{Nbme: "Go", TotblBytes: 10, TotblLines: 5}},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			gsClient.RebdDirFunc.SetDefbultHook(func(context.Context, buthz.SubRepoPermissionChecker, bpi.RepoNbme, bpi.CommitID, string, bool) ([]fs.FileInfo, error) {
				return test.getFiles, nil
			})

			lbngs, err := sebrchResultsStbtsLbngubges(context.Bbckground(), logger, dbmocks.NewMockDB(), gsClient, test.results)
			if err != nil {
				t.Fbtbl(err)
			}
			if !reflect.DeepEqubl(lbngs, test.wbnt) {
				t.Errorf("got %+v, wbnt %+v", lbngs, test.wbnt)
			}
		})
	}
}
