package resolvers

import (
	"context"
	"math"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type embeddingsSetupProgressResolver struct {
	repos *[]string
	db    database.DB

	once sync.Once

	currentRepo      *string
	currentProcessed *int32
	currentTotal     *int32
	overallProgress  int32
	oneRepoCompleted bool
	err              error
}

func (r *embeddingsSetupProgressResolver) getStatus(ctx context.Context) {
	r.once.Do(func() {
		var requestedRepos []string
		if r.repos != nil {
			requestedRepos = *r.repos
		}
		repos, err := getAllRepos(ctx, r.db, requestedRepos)
		if err != nil {
			r.err = err
			return
		}

		repoProgress := make([]float64, 0, len(repos))
		for _, repo := range repos {
			p, err := r.getProgressForRepo(ctx, repo)
			if err != nil {
				r.err = err
			}
			repoProgress = append(repoProgress, p)
		}
		r.overallProgress = calculateOverallPercent(repoProgress)
	})
}

func (r *embeddingsSetupProgressResolver) OverallPercentComplete(ctx context.Context) (int32, error) {
	r.getStatus(ctx)
	if r.err != nil {
		return 0, r.err
	}
	return r.overallProgress, nil
}

func (r *embeddingsSetupProgressResolver) CurrentRepository(ctx context.Context) *string {
	r.getStatus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentRepo
}

func (r *embeddingsSetupProgressResolver) CurrentRepositoryFilesProcessed(ctx context.Context) *int32 {
	r.getStatus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentProcessed
}

func (r *embeddingsSetupProgressResolver) CurrentRepositoryTotalFilesToProcess(ctx context.Context) *int32 {
	r.getStatus(ctx)
	if r.err != nil {
		return nil
	}
	return r.currentTotal
}

func (r *embeddingsSetupProgressResolver) OneRepositoryReady(ctx context.Context) bool {
	r.getStatus(ctx)
	if r.err != nil {
		return false
	}
	return r.oneRepoCompleted
}

func getAllRepos(ctx context.Context, db database.DB, repoNames []string) ([]types.MinimalRepo, error) {
	opts := database.ReposListOptions{}
	if len(repoNames) > 0 {
		opts.Names = repoNames
	}
	repos, err := db.Repos().ListMinimalRepos(ctx, opts)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// getProgressForRepo calculates the progress for a given repository based on its embedding job stats.
//
// ctx - The context for the request.
// current - The current repository to calculate progress for.
//
// Returns the progress percentage (0-1) and any errors encountered.
//
// It does the following:
// - Gets the last 10 embedding jobs for the repository from the store.
// - Checks each job's state:
//   - If completed, returns 1 (100% progress).
//   - If processing, gets the job stats and calculates the progress based on files processed/total files.
//     It also sets the currentRepo, currentProcessed and currentTotal fields.
//
// - Returns the progress percentage, defaulting to 0 if no processing job found.
func (r *embeddingsSetupProgressResolver) getProgressForRepo(ctx context.Context, current types.MinimalRepo) (float64, error) {
	var progress float64
	hasFailedJob := false
	hasPendingJob := false
	embeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobs, err := embeddingsStore.ListRepoEmbeddingJobs(ctx, repo.ListOpts{Repo: &current.ID, PaginationArgs: &database.PaginationArgs{First: pointers.Ptr(10)}})
	if err != nil {
		return progress, err
	}
	for _, job := range jobs {
		if job.IsRepoEmbeddingJobScheduledOrCompleted() {
			hasPendingJob = true
		}
		if job.StartedAt == nil {
			continue
		}
		switch job.State {
		case "canceled", "failed":
			hasFailedJob = true
		case "completed":
			r.oneRepoCompleted = true
			return 1, nil
		case "processing":
			r.currentRepo = pointers.Ptr(string(current.Name))
			status, err := embeddingsStore.GetRepoEmbeddingJobStats(ctx, job.ID)
			if err != nil {
				return 0, err
			}
			if status.CodeIndexStats.FilesScheduled == 0 {
				continue
			}
			r.currentProcessed, r.currentTotal, progress = getProgress(status)
			// we continue to process the rest of the jobs incase there is another job that is already completed
		}
	}

	// if we have a failed or canceled job and there are no jobs scheduled or completed
	// we advance the progress to 100% for that repo because it will not completed
	if hasFailedJob && !hasPendingJob {
		return 1, nil
	}

	return progress, nil
}

// calculateOverallPercent calculates the overall percentage completion based on multiple progress percentages.
//
// percents - The list of progress percentages (0-1) to calculate the overall percentage from.
//
// Returns the overall percentage completion (0-100).
//
// It does the following:
// - Sums the total progress percentages and total number of percentages.
// - Calculates the overall percentage by dividing the total progress by total number of percentages.
// - Rounds up and caps at 100% completion.
// - Returns the overall percentage completion.
func calculateOverallPercent(percents []float64) int32 {
	if len(percents) == 0 {
		return 0
	}
	var total, completed float64
	for _, percent := range percents {
		total += 1
		completed += percent
	}

	overall := int32(math.Ceil(completed / total * 100))
	if overall >= 100 {
		return 100
	}
	return overall
}

// getProgress calculates the overall progress based on indexing status
// returns number of files indexed, total files, and the progress percentage
func getProgress(status repo.EmbedRepoStats) (*int32, *int32, float64) {
	skipped := 0
	for _, count := range status.CodeIndexStats.FilesSkipped {
		skipped += count
	}
	for _, count := range status.TextIndexStats.FilesSkipped {
		skipped += count
	}

	embedded := status.CodeIndexStats.FilesEmbedded + status.TextIndexStats.FilesEmbedded
	scheduled := status.CodeIndexStats.FilesScheduled + status.TextIndexStats.FilesScheduled
	progress := float64(embedded+skipped) / float64(scheduled)

	return pointers.Ptr(int32(embedded) + int32(skipped)), pointers.Ptr(int32(scheduled)), progress
}
