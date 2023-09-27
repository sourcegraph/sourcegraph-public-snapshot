pbckbge embeddings

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestScheduleRepositoriesForEmbedding(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	crebtedRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo)
	require.NoError(t, err)

	// Crebte b repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)
	_, err = store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "coffee")
	require.NoError(t, err)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefbultBrbnchFunc.SetDefbultReturn("mbin", "coffee", nil)

	// By defbult, we shouldn't schedule b new job for the sbme revision
	repoNbmes := []bpi.RepoNbme{"github.com/sourcegrbph/sourcegrbph"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 1, count)

	// With the 'force' brgument, b new job will be scheduled bnywbys
	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, true, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 2, count)
}

func TestScheduleRepositoriesForEmbeddingRepoNotFound(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	crebtedRepo0 := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo0)
	require.NoError(t, err)

	// Crebte b repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("mbin", "sgrevision", nil)

	repoNbmes := []bpi.RepoNbme{"github.com/repo/notfound", "github.com/sourcegrbph/sourcegrbph"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 1, count)

	pbttern := "github.com/sourcegrbph/sourcegrbph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, "queued", jobs[0].Stbte)
}

func TestScheduleRepositoriesForEmbeddingInvblidDefbultBrbnch(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	crebtedRepo0 := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo0)
	require.NoError(t, err)

	// Crebte b repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("", "sgrevision", nil)

	repoNbmes := []bpi.RepoNbme{"github.com/sourcegrbph/sourcegrbph"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 1, count)

	pbttern := "github.com/sourcegrbph/sourcegrbph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, "queued", jobs[0].Stbte)
}

func TestScheduleRepositoriesForEmbeddingFbiled(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	crebtedRepo0 := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo0)
	require.NoError(t, err)

	crebtedRepo1 := &types.Repo{Nbme: "github.com/sourcegrbph/zoekt", URI: "github.com/sourcegrbph/zoekt", ExternblRepo: bpi.ExternblRepoSpec{}}
	err = repoStore.Crebte(ctx, crebtedRepo1)
	require.NoError(t, err)

	// Crebte b repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("", "sgrevision", nil)
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("mbin", "zoektrevision", nil)

	repoNbmes := []bpi.RepoNbme{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/zoekt"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 2, count)

	pbttern := "github.com/sourcegrbph/sourcegrbph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, "queued", jobs[0].Stbte)

	sgJobID := jobs[0].ID

	pbttern = "github.com/sourcegrbph/zoekt"
	jobs, err = store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, "queued", jobs[0].Stbte)

	zoektJobID := jobs[0].ID

	// Set jobs to expected completion stbtes, with empty repo resulting in fbiled
	setJobStbte(t, ctx, store, sgJobID, "fbiled")
	setJobStbte(t, ctx, store, zoektJobID, "fbiled")

	// Reschedule
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("", "sgrevision", nil)
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("mbin", "zoektrevision", nil)

	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	// fbiled job is rescheduled unless revision is empty
	require.Equbl(t, 3, count)

	// repo with previous fbilure due to empty revision is rescheduled when repo is vblid (error is nil bnd ref is non-empty)
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("mbin", "sgrevision", nil)
	gitserverClient.GetDefbultBrbnchFunc.PushReturn("mbin", "zoektrevision", nil)

	err = ScheduleRepositoriesForEmbedding(ctx, repoNbmes, fblse, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	// fbiled job is rescheduled for sourcegrbph once repo is vblid
	require.Equbl(t, 4, count)
}

func setJobStbte(t *testing.T, ctx context.Context, store repo.RepoEmbeddingJobsStore, jobID int, stbte string) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET stbte = %s, finished_bt = now() WHERE id = %s", stbte, jobID))
	if err != nil {
		t.Fbtblf("fbiled to set repo embedding job stbte: %s", err)
	}
}
