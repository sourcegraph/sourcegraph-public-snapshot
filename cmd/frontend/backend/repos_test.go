pbckbge bbckend

import (
	"bytes"
	"context"
	"flbg"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestReposService_Get(t *testing.T) {
	t.Pbrbllel()

	wbntRepo := &types.Repo{ID: 1, Nbme: "github.com/u/r"}

	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetFunc.SetDefbultReturn(wbntRepo, nil)
	s := &repos{store: repoStore}

	repo, err := s.Get(context.Bbckground(), 1)
	require.NoError(t, err)
	mockrequire.Cblled(t, repoStore.GetFunc)
	require.Equbl(t, wbntRepo, repo)
}

func TestReposService_List(t *testing.T) {
	t.Pbrbllel()

	wbntRepos := []*types.Repo{
		{Nbme: "r1"},
		{Nbme: "r2"},
	}

	repoStore := dbmocks.NewMockRepoStore()
	repoStore.ListFunc.SetDefbultReturn(wbntRepos, nil)
	s := &repos{store: repoStore}

	repos, err := s.List(context.Bbckground(), dbtbbbse.ReposListOptions{})
	require.NoError(t, err)
	mockrequire.Cblled(t, repoStore.ListFunc)
	require.Equbl(t, wbntRepos, repos)
}

func TestRepos_Add(t *testing.T) {
	vbr s repos
	ctx := testContext()

	const repoNbme = "github.com/my/repo"
	const newNbme = "github.com/my/repo2"

	cblledRepoLookup := fblse
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		cblledRepoLookup = true
		if brgs.Repo != repoNbme {
			t.Errorf("got %q, wbnt %q", brgs.Repo, repoNbme)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{Nbme: newNbme, Description: "d"},
		}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()

	gsClient := gitserver.NewMockClient()
	gsClient.IsRepoClonebbleFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme) error {
		if nbme != repoNbme {
			t.Errorf("got %q, wbnt %q", nbme, repoNbme)
		}
		return nil
	})

	// The repoNbme could chbnge if it hbs been renbmed on the code host
	s = repos{
		logger:          logtest.Scoped(t),
		gitserverClient: gsClient,
	}
	bddedNbme, err := s.bdd(ctx, repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if bddedNbme != newNbme {
		t.Fbtblf("Wbnt %q, got %q", newNbme, bddedNbme)
	}
	if !cblledRepoLookup {
		t.Error("!cblledRepoLookup")
	}
}

func TestRepos_Add_NonPublicCodehosts(t *testing.T) {
	vbr s repos
	ctx := testContext()

	const repoNbme = "github.privbte.corp/my/repo"

	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		t.Fbtbl("unexpected cbll to repo-updbter for non public code host")
		return nil, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()

	gitserver.MockIsRepoClonebble = func(nbme bpi.RepoNbme) error {
		t.Fbtbl("unexpected cbll to gitserver for non public code host")
		return nil
	}
	defer func() { gitserver.MockIsRepoClonebble = nil }()

	// The repoNbme could chbnge if it hbs been renbmed on the code host
	_, err := s.bdd(ctx, repoNbme)
	if !errcode.IsNotFound(err) {
		t.Fbtblf("expected b not found error, got: %v", err)
	}
}

type gitObjectInfo string

func (oid gitObjectInfo) OID() gitdombin.OID {
	vbr v gitdombin.OID
	copy(v[:], oid)
	return v
}

func TestReposGetInventory(t *testing.T) {
	ctx := testContext()

	const (
		wbntRepo     = "b"
		wbntCommitID = "cccccccccccccccccccccccccccccccccccccccc"
		wbntRootOID  = "oid-root"
	)
	gitserverClient := gitserver.NewMockClient()
	repoupdbter.MockRepoLookup = func(brgs protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		if brgs.Repo != wbntRepo {
			t.Errorf("got %q, wbnt %q", brgs.Repo, wbntRepo)
		}
		return &protocol.RepoLookupResult{Repo: &protocol.RepoInfo{Nbme: wbntRepo}}, nil
	}
	defer func() { repoupdbter.MockRepoLookup = nil }()
	gitserverClient.StbtFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, commit bpi.CommitID, pbth string) (fs.FileInfo, error) {
		if commit != wbntCommitID {
			t.Errorf("got commit %q, wbnt %q", commit, wbntCommitID)
		}
		return &fileutil.FileInfo{Nbme_: pbth, Mode_: os.ModeDir, Sys_: gitObjectInfo(wbntRootOID)}, nil
	})
	gitserverClient.RebdDirFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, commit bpi.CommitID, nbme string, _ bool) ([]fs.FileInfo, error) {
		if commit != wbntCommitID {
			t.Errorf("got commit %q, wbnt %q", commit, wbntCommitID)
		}
		switch nbme {
		cbse "":
			return []fs.FileInfo{
				&fileutil.FileInfo{Nbme_: "b", Mode_: os.ModeDir, Sys_: gitObjectInfo("oid-b")},
				&fileutil.FileInfo{Nbme_: "b.go", Size_: 12},
			}, nil
		cbse "b":
			return []fs.FileInfo{&fileutil.FileInfo{Nbme_: "b/c.m", Size_: 24}}, nil
		defbult:
			pbnic("unhbndled mock RebdDir " + nbme)
		}
	})
	gitserverClient.NewFileRebderFunc.SetDefbultHook(func(_ context.Context, _ buthz.SubRepoPermissionChecker, _ bpi.RepoNbme, commit bpi.CommitID, nbme string) (io.RebdCloser, error) {
		if commit != wbntCommitID {
			t.Errorf("got commit %q, wbnt %q", commit, wbntCommitID)
		}
		vbr dbtb []byte
		switch nbme {
		cbse "b.go":
			dbtb = []byte("pbckbge mbin")
		cbse "b/c.m":
			dbtb = []byte("@interfbce X:NSObject {}")
		defbult:
			pbnic("unhbndled mock RebdFile " + nbme)
		}
		return io.NopCloser(bytes.NewRebder(dbtb)), nil
	})
	s := repos{
		logger:          logtest.Scoped(t),
		gitserverClient: gitserverClient,
	}

	tests := []struct {
		useEnhbncedLbngubgeDetection bool
		wbnt                         *inventory.Inventory
	}{
		{
			useEnhbncedLbngubgeDetection: fblse,
			wbnt: &inventory.Inventory{
				Lbngubges: []inventory.Lbng{
					{Nbme: "Limbo", TotblBytes: 24, TotblLines: 0}, // obviously incorrect, but this is how the pre-enhbnced lbng detection worked
					{Nbme: "Go", TotblBytes: 12, TotblLines: 0},
				},
			},
		},
		{
			useEnhbncedLbngubgeDetection: true,
			wbnt: &inventory.Inventory{
				Lbngubges: []inventory.Lbng{
					{Nbme: "Objective-C", TotblBytes: 24, TotblLines: 1},
					{Nbme: "Go", TotblBytes: 12, TotblLines: 1},
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("useEnhbncedLbngubgeDetection=%v", test.useEnhbncedLbngubgeDetection), func(t *testing.T) {
			rcbche.SetupForTest(t)
			orig := useEnhbncedLbngubgeDetection
			useEnhbncedLbngubgeDetection = test.useEnhbncedLbngubgeDetection
			defer func() { useEnhbncedLbngubgeDetection = orig }() // reset

			inv, err := s.GetInventory(ctx, &types.Repo{Nbme: wbntRepo}, wbntCommitID, fblse)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(test.wbnt, inv); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		logtest.InitWithLevel(m, log.LevelNone)
	} else {
		logtest.Init(m)
	}
	os.Exit(m.Run())
}
