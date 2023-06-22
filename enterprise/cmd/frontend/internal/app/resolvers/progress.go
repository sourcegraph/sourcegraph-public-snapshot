package resolvers

import (
	"context"
	"math"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	err              error
}

func (r *embeddingsSetupProgressResolver) getStatus(ctx context.Context) {
	r.once.Do(func() {
		repos, err := getAllRepos(ctx, r.db)
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

func (r *embeddingsSetupProgressResolver) PercentComplete(ctx context.Context) (int32, error) {
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

func getAllRepos(ctx context.Context, db database.DB) ([]types.MinimalRepo, error) {
	repos, err := db.Repos().ListMinimalRepos(ctx, database.ReposListOptions{})
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (r *embeddingsSetupProgressResolver) getProgressForRepo(ctx context.Context, current types.MinimalRepo) (float64, error) {
	var progress float64
	embeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobs, err := embeddingsStore.ListRepoEmbeddingJobs(ctx, repo.ListOpts{Repo: &current.ID, PaginationArgs: &database.PaginationArgs{First: pointers.Ptr(10)}})
	if err != nil {
		return progress, err
	}
	for _, job := range jobs {
		if job.StartedAt == nil {
			continue
		}
		switch job.State {
		case "completed":
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

		default:
		}
	}
	return progress, nil
}

func calculateOverallPercent(percents []float64) int32 {
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
