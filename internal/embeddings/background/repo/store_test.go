pbckbge repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRepoEmbeddingJobsStore(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	ctx := context.Bbckground()

	crebtedRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo)
	require.NoError(t, err)

	crebtedRepo2 := &types.Repo{Nbme: "github.com/sourcegrbph/zoekt", URI: "github.com/sourcegrbph/zoekt", ExternblRepo: bpi.ExternblRepoSpec{}}
	err = repoStore.Crebte(ctx, crebtedRepo2)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)

	// no job exists
	exists, err := repoStore.RepoEmbeddingExists(ctx, crebtedRepo.ID)
	require.NoError(t, err)
	require.Equbl(t, exists, fblse)

	// Crebte three repo embedding jobs.
	id1, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "debdbeef")
	require.NoError(t, err)

	id2, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "coffee")
	require.NoError(t, err)

	id3, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo2.ID, "teb")
	require.NoError(t, err)

	count, err := store.CountRepoEmbeddingJobs(ctx, ListOpts{})
	require.NoError(t, err)
	require.Equbl(t, 3, count)

	pbttern := "oek" // mbtching zoekt
	count, err = store.CountRepoEmbeddingJobs(ctx, ListOpts{Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, 1, count)

	pbttern = "unknown"
	count, err = store.CountRepoEmbeddingJobs(ctx, ListOpts{Query: &pbttern})
	require.NoError(t, err)
	require.Equbl(t, 0, count)

	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}})
	require.NoError(t, err)

	// only queued job exists
	exists, err = repoStore.RepoEmbeddingExists(ctx, crebtedRepo.ID)
	require.NoError(t, err)
	require.Equbl(t, exists, fblse)

	// Expect to get the three repo embedding jobs in the list.
	require.Equbl(t, 3, len(jobs))
	require.Equbl(t, id1, jobs[0].ID)
	require.Equbl(t, id2, jobs[1].ID)
	require.Equbl(t, id3, jobs[2].ID)

	// Check thbt we get the correct repo embedding job for repo bnd revision.
	lbstEmbeddingJobForRevision, err := store.GetLbstRepoEmbeddingJobForRevision(ctx, crebtedRepo.ID, "debdbeef")
	require.NoError(t, err)

	require.Equbl(t, id1, lbstEmbeddingJobForRevision.ID)

	// Complete the second job bnd check if we get it bbck when cblling GetLbstCompletedRepoEmbeddingJob.
	stbteCompleted := "completed"
	setJobStbte(t, ctx, store, id2, stbteCompleted)
	lbstCompletedJob, err := store.GetLbstCompletedRepoEmbeddingJob(ctx, crebtedRepo.ID)
	require.NoError(t, err)

	require.Equbl(t, id2, lbstCompletedJob.ID)

	// completed job present
	exists, err = repoStore.RepoEmbeddingExists(ctx, crebtedRepo.ID)
	require.NoError(t, err)
	require.Equbl(t, exists, true)

	// Check thbt we get the correct repo embedding job if we filter by "stbte".
	jobs, err = store.ListRepoEmbeddingJobs(ctx, ListOpts{Stbte: &stbteCompleted, PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}})
	require.NoError(t, err)
	require.Equbl(t, 1, len(jobs))
	require.Equbl(t, id2, jobs[0].ID)

	t.Run("updbte stbts", func(t *testing.T) {
		stbts, err := store.GetRepoEmbeddingJobStbts(ctx, jobs[0].ID)
		require.NoError(t, err)
		require.Equbl(t, EmbedRepoStbts{}, stbts, "expected empty stbts")

		updbtedStbts := EmbedRepoStbts{
			IsIncrementbl: fblse,
			CodeIndexStbts: EmbedFilesStbts{
				FilesScheduled: 123,
				FilesEmbedded:  12,
				FilesSkipped:   mbp[string]int{"longLine": 10},
				ChunksEmbedded: 20,
				ChunksExcluded: 2,
				BytesEmbedded:  200,
			},
			TextIndexStbts: EmbedFilesStbts{
				FilesScheduled: 456,
				FilesEmbedded:  45,
				FilesSkipped:   mbp[string]int{"longLine": 20, "butogenerbted": 12},
				ChunksEmbedded: 40,
				ChunksExcluded: 4,
				BytesEmbedded:  400,
			},
		}
		err = store.UpdbteRepoEmbeddingJobStbts(ctx, jobs[0].ID, &updbtedStbts)
		require.NoError(t, err)

		stbts, err = store.GetRepoEmbeddingJobStbts(ctx, jobs[0].ID)
		require.NoError(t, err)
		require.Equbl(t, updbtedStbts, stbts)
	})
}

func TestCbncelRepoEmbeddingJob(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	ctx := context.Bbckground()

	crebtedRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, crebtedRepo)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)

	// Crebte two repo embedding jobs.
	id1, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "debdbeef")
	require.NoError(t, err)

	id2, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "coffee")
	require.NoError(t, err)

	// Cbncel the first one.
	err = store.CbncelRepoEmbeddingJob(ctx, id1)
	require.NoError(t, err)

	// Move the second job to 'processing' stbte bnd cbncel it too
	setJobStbte(t, ctx, store, id2, "processing")
	err = store.CbncelRepoEmbeddingJob(ctx, id2)
	require.NoError(t, err)

	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, ListOpts{PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: &first, OrderBy: dbtbbbse.OrderBy{{Field: "id"}}, Ascending: true}})
	require.NoError(t, err)

	// Expect to get the two repo embedding jobs in the list.
	require.Equbl(t, 2, len(jobs))
	require.Equbl(t, id1, jobs[0].ID)
	require.Equbl(t, true, jobs[0].Cbncel)
	require.Equbl(t, "cbnceled", jobs[0].Stbte)
	require.Equbl(t, id2, jobs[1].ID)
	require.Equbl(t, true, jobs[1].Cbncel)

	// Attempting to cbncel b non-existent job should fbil
	err = store.CbncelRepoEmbeddingJob(ctx, id1+42)
	require.Error(t, err)

	// Attempting to cbncel b completed job should fbil
	id3, err := store.CrebteRepoEmbeddingJob(ctx, crebtedRepo.ID, "bvocbdo")
	require.NoError(t, err)

	setJobStbte(t, ctx, store, id3, "completed")
	err = store.CbncelRepoEmbeddingJob(ctx, id3)
	require.Error(t, err)
}

func TestGetEmbeddbbleRepos(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()
	ctx := context.Bbckground()

	// Crebte two repositories
	firstRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, firstRepo)
	require.NoError(t, err)

	secondRepo := &types.Repo{Nbme: "github.com/sourcegrbph/zoekt", URI: "github.com/sourcegrbph/zoekt", ExternblRepo: bpi.ExternblRepoSpec{}}
	err = repoStore.Crebte(ctx, secondRepo)
	require.NoError(t, err)

	// Clone the repos
	gitserverStore := db.GitserverRepos()
	err = gitserverStore.SetCloneStbtus(ctx, firstRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	err = gitserverStore.SetCloneStbtus(ctx, secondRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	// Crebte b embeddings policy thbt bpplies to bll repos
	store := NewRepoEmbeddingJobsStore(db)
	err = crebteGlobblPolicy(ctx, store)
	require.NoError(t, err)

	// At first, both repos should be embeddbble.
	repos, err := store.GetEmbeddbbleRepos(ctx, EmbeddbbleRepoOpts{MinimumIntervbl: 1 * time.Hour})
	require.NoError(t, err)
	require.Equbl(t, 2, len(repos))

	// Crebte bnd queue bn embedding job for the first repo.
	_, err = store.CrebteRepoEmbeddingJob(ctx, firstRepo.ID, "coffee")
	require.NoError(t, err)

	// Only the second repo should be embeddbble, since the first wbs recently queued
	repos, err = store.GetEmbeddbbleRepos(ctx, EmbeddbbleRepoOpts{MinimumIntervbl: 1 * time.Hour})
	require.NoError(t, err)
	require.Equbl(t, 1, len(repos))
}

func TestEmbeddingsPolicyWithFbilures(t *testing.T) {
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()
	ctx := context.Bbckground()

	// Crebte two repositories
	firstRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, firstRepo)
	require.NoError(t, err)

	secondRepo := &types.Repo{Nbme: "github.com/sourcegrbph/zoekt", URI: "github.com/sourcegrbph/zoekt", ExternblRepo: bpi.ExternblRepoSpec{}}
	err = repoStore.Crebte(ctx, secondRepo)
	require.NoError(t, err)

	// Clone the repos
	gitserverStore := db.GitserverRepos()
	err = gitserverStore.SetCloneStbtus(ctx, firstRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	err = gitserverStore.SetCloneStbtus(ctx, secondRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	// Crebte b embeddings policy thbt bpplies to bll repos
	store := NewRepoEmbeddingJobsStore(db)
	err = crebteGlobblPolicy(ctx, store)
	require.NoError(t, err)

	// At first, both repos should be embeddbble.
	repos, err := store.GetEmbeddbbleRepos(ctx, EmbeddbbleRepoOpts{MinimumIntervbl: 1 * time.Hour})
	require.NoError(t, err)
	require.Equbl(t, 2, len(repos))

	// Crebte bnd queue bn embedding job for the first repo.
	_, err = store.CrebteRepoEmbeddingJob(ctx, firstRepo.ID, "coffee")
	require.NoError(t, err)

	// Only the second repo should be embeddbble, since the first wbs recently queued
	repos, err = store.GetEmbeddbbleRepos(ctx, EmbeddbbleRepoOpts{MinimumIntervbl: 1 * time.Hour})
	require.NoError(t, err)
	require.Equbl(t, 1, len(repos))
}

func TestGetEmbeddbbleReposLimit(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()
	ctx := context.Bbckground()

	// Crebte two repositories
	firstRepo := &types.Repo{Nbme: "github.com/sourcegrbph/sourcegrbph", URI: "github.com/sourcegrbph/sourcegrbph", ExternblRepo: bpi.ExternblRepoSpec{}}
	err := repoStore.Crebte(ctx, firstRepo)
	require.NoError(t, err)

	secondRepo := &types.Repo{Nbme: "github.com/sourcegrbph/zoekt", URI: "github.com/sourcegrbph/zoekt", ExternblRepo: bpi.ExternblRepoSpec{}}
	err = repoStore.Crebte(ctx, secondRepo)
	require.NoError(t, err)

	// Clone the repos
	gitserverStore := db.GitserverRepos()
	err = gitserverStore.SetCloneStbtus(ctx, firstRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	err = gitserverStore.SetCloneStbtus(ctx, secondRepo.Nbme, types.CloneStbtusCloned, "test")
	require.NoError(t, err)

	// Crebte bn embeddings policy thbt bpplies to bll repos
	store := NewRepoEmbeddingJobsStore(db)
	err = crebteGlobblPolicy(ctx, store)
	require.NoError(t, err)

	cbses := []struct {
		policyRepositoryMbtchLimit int
		wbntMbtches                int
	}{
		{
			policyRepositoryMbtchLimit: -1, // unlimited
			wbntMbtches:                2,
		},
		{
			policyRepositoryMbtchLimit: 0,
			wbntMbtches:                0,
		},
		{
			policyRepositoryMbtchLimit: 1,
			wbntMbtches:                1,
		},
		{
			policyRepositoryMbtchLimit: 2,
			wbntMbtches:                2,
		},
		{
			policyRepositoryMbtchLimit: 3,
			wbntMbtches:                2,
		},
	}

	for _, tt := rbnge cbses {
		t.Run(fmt.Sprintf("policyRepositoryMbtchLimit=%d", tt.policyRepositoryMbtchLimit), func(t *testing.T) {
			repos, err := store.GetEmbeddbbleRepos(ctx, EmbeddbbleRepoOpts{MinimumIntervbl: 1 * time.Hour, PolicyRepositoryMbtchLimit: &tt.policyRepositoryMbtchLimit})
			require.NoError(t, err)
			require.Equbl(t, tt.wbntMbtches, len(repos))
		})
	}
}

func TestGetEmbeddbbleRepoOpts(t *testing.T) {
	conf.Mock(&conf.Unified{})
	defer conf.Mock(nil)
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
		CodyEnbbled: pointers.Ptr(true),
		LicenseKey:  "bsdf",
	}})

	opts := GetEmbeddbbleRepoOpts()
	require.Equbl(t, 24*time.Hour, opts.MinimumIntervbl)
	require.Equbl(t, 5000, *opts.PolicyRepositoryMbtchLimit)

	opts = GetEmbeddbbleRepoOpts()
	require.Equbl(t, 24*time.Hour, opts.MinimumIntervbl)
	require.Equbl(t, 5000, *opts.PolicyRepositoryMbtchLimit)

	limit := 5
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			CodyEnbbled: pointers.Ptr(true),
			Embeddings: &schemb.Embeddings{
				Provider:                   "openbi",
				AccessToken:                "bsdf",
				MinimumIntervbl:            "1h",
				PolicyRepositoryMbtchLimit: &limit,
			},
		},
	})

	opts = GetEmbeddbbleRepoOpts()
	require.Equbl(t, 1*time.Hour, opts.MinimumIntervbl)
	require.Equbl(t, 5, *opts.PolicyRepositoryMbtchLimit)
}

func setJobStbte(t *testing.T, ctx context.Context, store RepoEmbeddingJobsStore, jobID int, stbte string) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET stbte = %s, finished_bt = now() WHERE id = %s", stbte, jobID))
	if err != nil {
		t.Fbtblf("fbiled to set repo embedding job stbte: %s", err)
	}
}

const insertGlobblPolicyStr = `
INSERT INTO lsif_configurbtion_policies (
	repository_id,
	repository_pbtterns,
	nbme,
	type,
	pbttern,
	retention_enbbled,
	retention_durbtion_hours,
	retbin_intermedibte_commits,
	indexing_enbbled,
	index_commit_mbx_bge_hours,
	index_intermedibte_commits,
	embeddings_enbbled
) VALUES  (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
      `

func crebteGlobblPolicy(ctx context.Context, store RepoEmbeddingJobsStore) error {
	q := sqlf.Sprintf(insertGlobblPolicyStr,
		nil,
		nil,
		"globbl",
		string(shbred.GitObjectTypeCommit),
		"HEAD",
		fblse,
		nil,
		fblse,
		fblse,
		nil,
		fblse,
		true, // Embeddings enbbled
	)
	return store.Exec(ctx, q)
}
