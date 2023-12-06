package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func ScheduleRepositories(
	ctx context.Context,
	repoNames []api.RepoName,
	forceReschedule bool,
	db database.DB,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	gitserverClient gitserver.Client,
) error {
	tx, err := repoEmbeddingJobsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	repoStore := db.Repos()
	for _, repoName := range repoNames {
		r, err := repoStore.GetByName(ctx, repoName)
		if err != nil {
			return err
		}

		refName, latestRevision, err := gitserverClient.GetDefaultBranch(ctx, r.Name, false)
		if err != nil || refName == "" {
			return err
		}

		if !forceReschedule {
			job, _ := tx.GetLastRepoEmbeddingJobForRevision(ctx, r.ID, latestRevision)
			if job != nil && isScheduledOrCompleted(job) {
				continue
			}
		}

		_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision)
		if err != nil {
			return err
		}
	}
	return nil
}

func ScheduleRepositoriesForPolicy(
	ctx context.Context,
	repoIDs []api.RepoID,
	db database.DB,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	gitserverClient gitserver.Client,
) error {
	tx, err := repoEmbeddingJobsStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	repoStore := db.Repos()
	repos, err := repoStore.ListMinimalRepos(ctx, database.ReposListOptions{IDs: repoIDs})

	if err != nil {
		return err
	}

	for _, r := range repos {
		refName, latestRevision, err := gitserverClient.GetDefaultBranch(ctx, r.Name, false)
		// enqueue with an empty revision and let handler determine whether job can execute
		if err != nil || refName == "" {
			latestRevision = ""
		}

		job, _ := tx.GetLastRepoEmbeddingJobForRevision(ctx, r.ID, latestRevision)
		if job != nil && isScheduledOrCompleted(job) {
			continue
		}

		_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision)
		if err != nil {
			return err
		}
	}
	return nil
}

// We skip creating a repo embedding job for a repo at revision if there already exists
// an identical job that has been completed, or is scheduled to run (processing or queued).
// In the case that a repo is empty, a job is considered "completed" even if it failed.
func isScheduledOrCompleted(job *repo.RepoEmbeddingJob) bool {
	return job.IsRepoEmbeddingJobScheduledOrCompleted() || job.EmptyRepoEmbeddingJob()
}
