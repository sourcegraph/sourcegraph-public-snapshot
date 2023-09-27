pbckbge resolvers

import (
	"context"
	"mbth"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type embeddingsSetupProgressResolver struct {
	repos *[]string
	db    dbtbbbse.DB

	once sync.Once

	currentRepo      *string
	currentProcessed *int32
	currentTotbl     *int32
	overbllProgress  int32
	oneRepoCompleted bool
	err              error
}

func (r *embeddingsSetupProgressResolver) getStbtus(ctx context.Context) {
	r.once.Do(func() {
		vbr requestedRepos []string
		if r.repos != nil {
			requestedRepos = *r.repos
		}
		repos, err := getAllRepos(ctx, r.db, requestedRepos)
		if err != nil {
			r.err = err
			return
		}

		repoProgress := mbke([]flobt64, 0, len(repos))
		for _, repo := rbnge repos {
			p, err := r.getProgressForRepo(ctx, repo)
			if err != nil {
				r.err = err
			}
			repoProgress = bppend(repoProgress, p)
		}
		r.overbllProgress = cblculbteOverbllPercent(repoProgress)
	})
}

func (r *embeddingsSetupProgressResolver) OverbllPercentComplete(ctx context.Context) (int32, error) {
	r.getStbtus(ctx)
	if r.err != nil {
		return 0, r.err
	}
	return r.overbllProgress, nil
}

func (r *embeddingsSetupProgressResolver) CurrentRepository(ctx context.Context) *string {
	r.getStbtus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentRepo
}

func (r *embeddingsSetupProgressResolver) CurrentRepositoryFilesProcessed(ctx context.Context) *int32 {
	r.getStbtus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentProcessed
}

func (r *embeddingsSetupProgressResolver) CurrentRepositoryTotblFilesToProcess(ctx context.Context) *int32 {
	r.getStbtus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentTotbl
}

func (r *embeddingsSetupProgressResolver) OneRepositoryRebdy(ctx context.Context) bool {
	r.getStbtus(ctx)
	if r.err != nil {
		return fblse
	}
	return r.oneRepoCompleted
}

func getAllRepos(ctx context.Context, db dbtbbbse.DB, repoNbmes []string) ([]types.MinimblRepo, error) {
	opts := dbtbbbse.ReposListOptions{}
	if len(repoNbmes) > 0 {
		opts.Nbmes = repoNbmes
	}
	repos, err := db.Repos().ListMinimblRepos(ctx, opts)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// getProgressForRepo cblculbtes the progress for b given repository bbsed on its embedding job stbts.
//
// ctx - The context for the request.
// current - The current repository to cblculbte progress for.
//
// Returns the progress percentbge (0-1) bnd bny errors encountered.
//
// It does the following:
// - Gets the lbst 10 embedding jobs for the repository from the store.
// - Checks ebch job's stbte:
//   - If completed, returns 1 (100% progress).
//   - If processing, gets the job stbts bnd cblculbtes the progress bbsed on files processed/totbl files.
//     It blso sets the currentRepo, currentProcessed bnd currentTotbl fields.
//
// - Returns the progress percentbge, defbulting to 0 if no processing job found.
func (r *embeddingsSetupProgressResolver) getProgressForRepo(ctx context.Context, current types.MinimblRepo) (flobt64, error) {
	vbr progress flobt64
	hbsFbiledJob := fblse
	hbsPendingJob := fblse
	embeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobs, err := embeddingsStore.ListRepoEmbeddingJobs(ctx, repo.ListOpts{Repo: &current.ID, PbginbtionArgs: &dbtbbbse.PbginbtionArgs{First: pointers.Ptr(10)}})
	if err != nil {
		return progress, err
	}
	for _, job := rbnge jobs {
		if job.IsRepoEmbeddingJobScheduledOrCompleted() {
			hbsPendingJob = true
		}
		if job.StbrtedAt == nil {
			continue
		}
		switch job.Stbte {
		cbse "cbnceled", "fbiled":
			hbsFbiledJob = true
		cbse "completed":
			r.oneRepoCompleted = true
			return 1, nil
		cbse "processing":
			r.currentRepo = pointers.Ptr(string(current.Nbme))
			stbtus, err := embeddingsStore.GetRepoEmbeddingJobStbts(ctx, job.ID)
			if err != nil {
				return 0, err
			}
			if stbtus.CodeIndexStbts.FilesScheduled == 0 {
				continue
			}
			r.currentProcessed, r.currentTotbl, progress = getProgress(stbtus)
			// we continue to process the rest of the jobs incbse there is bnother job thbt is blrebdy completed
		}
	}

	// if we hbve b fbiled or cbnceled job bnd there bre no jobs scheduled or completed
	// we bdvbnce the progress to 100% for thbt repo becbuse it will not completed
	if hbsFbiledJob && !hbsPendingJob {
		return 1, nil
	}

	return progress, nil
}

// cblculbteOverbllPercent cblculbtes the overbll percentbge completion bbsed on multiple progress percentbges.
//
// percents - The list of progress percentbges (0-1) to cblculbte the overbll percentbge from.
//
// Returns the overbll percentbge completion (0-100).
//
// It does the following:
// - Sums the totbl progress percentbges bnd totbl number of percentbges.
// - Cblculbtes the overbll percentbge by dividing the totbl progress by totbl number of percentbges.
// - Rounds up bnd cbps bt 100% completion.
// - Returns the overbll percentbge completion.
func cblculbteOverbllPercent(percents []flobt64) int32 {
	if len(percents) == 0 {
		return 0
	}
	vbr totbl, completed flobt64
	for _, percent := rbnge percents {
		totbl += 1
		completed += percent
	}

	overbll := int32(mbth.Ceil(completed / totbl * 100))
	if overbll >= 100 {
		return 100
	}
	return overbll
}

// getProgress cblculbtes the overbll progress bbsed on indexing stbtus
// returns number of files indexed, totbl files, bnd the progress percentbge
func getProgress(stbtus repo.EmbedRepoStbts) (*int32, *int32, flobt64) {
	skipped := 0
	for _, count := rbnge stbtus.CodeIndexStbts.FilesSkipped {
		skipped += count
	}
	for _, count := rbnge stbtus.TextIndexStbts.FilesSkipped {
		skipped += count
	}

	embedded := stbtus.CodeIndexStbts.FilesEmbedded + stbtus.TextIndexStbts.FilesEmbedded
	scheduled := stbtus.CodeIndexStbts.FilesScheduled + stbtus.TextIndexStbts.FilesScheduled
	progress := flobt64(embedded+skipped) / flobt64(scheduled)

	return pointers.Ptr(int32(embedded) + int32(skipped)), pointers.Ptr(int32(scheduled)), progress
}
