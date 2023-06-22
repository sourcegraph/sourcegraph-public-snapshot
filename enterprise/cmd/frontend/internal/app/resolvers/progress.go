package resolvers

import (
	"context"
	"math"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type embeddingsSetupProgressResolver struct {
	repos *[]string
	db    database.DB
}

func (r *embeddingsSetupProgressResolver) PercentComplete(ctx context.Context) (int32, error) {
	repos, err := getAllRepos(ctx, r.db)
	if err != nil {
		return 0, err
	}

	repoProgress := make([]float64, 0, len(repos))
	for _, repo := range repos {
		p, err := getProgressForRepo(ctx, r.db, repo.ID)
		if err != nil {
			return 0, err
		}
		repoProgress = append(repoProgress, p)
	}
	return calculateOverallPercent(repoProgress), nil

}

func getAllRepos(ctx context.Context, db database.DB) ([]types.MinimalRepo, error) {
	repos, err := db.Repos().ListMinimalRepos(ctx, database.ReposListOptions{})
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func getProgressForRepo(ctx context.Context, db database.DB, repoID api.RepoID) (float64, error) {
	var progress float64
	embeddingsStore := repo.NewRepoEmbeddingJobsStore(db)
	jobs, err := embeddingsStore.ListRepoEmbeddingJobs(ctx, repo.ListOpts{Repo: &repoID, PaginationArgs: &database.PaginationArgs{First: pointers.Ptr(10)}})
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
			status, err := embeddingsStore.GetRepoEmbeddingJobStats(ctx, job.ID)
			if err != nil {
				return 0, err
			}
			if status.CodeIndexStats.FilesScheduled == 0 {
				continue
			}
			progress = getProgress(status)

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

func getProgress(status repo.EmbedRepoStats) float64 {
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

	return progress
}
