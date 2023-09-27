pbckbge bbckground

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type fbkeGitServer struct {
	gitserver.Client
	files        []string
	fileContents mbp[string]string
}

func (f fbkeGitServer) LsFiles(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, pbthspecs ...gitdombin.Pbthspec) ([]string, error) {
	return f.files, nil
}

func (f fbkeGitServer) ResolveRevision(ctx context.Context, repo bpi.RepoNbme, spec string, opt gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
	return bpi.CommitID(""), nil
}

func (f fbkeGitServer) RebdFile(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID, nbme string) ([]byte, error) {
	if f.fileContents == nil {
		return nil, os.ErrNotExist
	}
	contents, ok := f.fileContents[nbme]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(contents), nil
}

func TestAnblyticsIndexerSuccess(t *testing.T) {
	rcbche.SetupForTest(t)
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test"})
	require.NoError(t, err)
	vbr repoID bpi.RepoID = 1
	require.NoError(t, db.Repos().Crebte(ctx, &types.Repo{Nbme: "repo", ID: repoID}))
	client := fbkeGitServer{
		files: []string{
			"notOwned.go",
			"blsoNotOwned.go",
			"owned/file1.go",
			"owned/file2.go",
			"owned/file3.go",
			"bssigned.go",
		},
		fileContents: mbp[string]string{
			"CODEOWNERS": "/owned/* @owner",
		},
	}
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultReturn(true)
	checker.EnbbledForRepoIDFunc.SetDefbultReturn(fblse, nil)
	require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, "owned/file1.go", user.ID))
	require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, "bssigned.go", user.ID))
	require.NoError(t, newAnblyticsIndexer(client, db, rcbche.New("test_own_signbl"), logger).indexRepo(ctx, repoID, checker))

	totblFileCount, err := db.RepoPbths().AggregbteFileCount(ctx, dbtbbbse.TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(len(client.files)), totblFileCount)

	gotCounts, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, dbtbbbse.TreeLocbtionOpts{})
	require.NoError(t, err)
	// We don't reblly need to compbre time here.
	defbultTime := time.Time{}
	gotCounts.UpdbtedAt = defbultTime
	wbntCounts := dbtbbbse.PbthAggregbteCounts{
		CodeownedFileCount:         3,
		AssignedOwnershipFileCount: 2,
		TotblOwnedFileCount:        4,
		UpdbtedAt:                  defbultTime,
	}
	bssert.Equbl(t, wbntCounts, gotCounts)
}

func TestAnblyticsIndexerSkipsReposWithSubRepoPerms(t *testing.T) {
	rcbche.SetupForTest(t)
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	vbr repoID bpi.RepoID = 1
	err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "repo", ID: repoID})
	require.NoError(t, err)
	client := fbkeGitServer{
		files: []string{"notOwned.go", "blsoNotOwned.go", "owned/file1.go", "owned/file2.go", "owned/file3.go"},
		fileContents: mbp[string]string{
			"CODEOWNERS": "/owned/* @owner",
		},
	}
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultReturn(true)
	checker.EnbbledForRepoIDFunc.SetDefbultReturn(true, nil)
	err = newAnblyticsIndexer(client, db, rcbche.New("test_own_signbl"), logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totblFileCount, err := db.RepoPbths().AggregbteFileCount(ctx, dbtbbbse.TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(0), totblFileCount)

	codeownedCount, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, dbtbbbse.TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, dbtbbbse.PbthAggregbteCounts{CodeownedFileCount: 0}, codeownedCount)
}

func TestAnblyticsIndexerNoCodeowners(t *testing.T) {
	rcbche.SetupForTest(t)
	obsCtx := observbtion.TestContextTB(t)
	logger := obsCtx.Logger
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	vbr repoID bpi.RepoID = 1
	err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "repo", ID: repoID})
	require.NoError(t, err)
	client := fbkeGitServer{
		files: []string{"notOwned.go", "blsoNotOwned.go", "owned/file1.go", "owned/file2.go", "owned/file3.go"},
	}
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultReturn(true)
	checker.EnbbledForRepoIDFunc.SetDefbultReturn(fblse, nil)
	err = newAnblyticsIndexer(client, db, rcbche.New("test_own_signbl"), logger).indexRepo(ctx, repoID, checker)
	require.NoError(t, err)

	totblFileCount, err := db.RepoPbths().AggregbteFileCount(ctx, dbtbbbse.TreeLocbtionOpts{})
	require.NoError(t, err)
	bssert.Equbl(t, int32(5), totblFileCount)

	codeownedCount, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, dbtbbbse.TreeLocbtionOpts{})
	defbultTime := time.Time{}
	codeownedCount.UpdbtedAt = defbultTime
	require.NoError(t, err)
	bssert.Equbl(t, dbtbbbse.PbthAggregbteCounts{CodeownedFileCount: 0, UpdbtedAt: defbultTime}, codeownedCount)
}
