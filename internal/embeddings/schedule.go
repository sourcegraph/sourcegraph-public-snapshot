package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ScheduleRepositoriesForEmbedding(
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

	var errs error

	repoStore := db.Repos()
	for _, repoName := range repoNames {
		// Scope the iteration to an anonymous function, so we can capture all errors and properly rollback tx in defer above.
		err = func() error {
			r, err := repoStore.GetByName(ctx, repoName)
			if err != nil {
				errs = errors.Append(errs, errors.Newf("Repo not found: %v", repoName))
				return nil
			}

			refName, latestRevision, err := gitserverClient.GetDefaultBranch(ctx, r.Name, false)
			// enqueue with an empty revision and let handler determine whether job can execute
			if err != nil || refName == "" {
				latestRevision = ""
			}

			// Skip creating a repo embedding job for a repo at revision, if there already exists
			// an identical job that has been completed, or is scheduled to run (processing or queued).
			if !forceReschedule {
				job, _ := tx.GetLastRepoEmbeddingJobForRevision(ctx, r.ID, latestRevision)

				// if job previously failed then only resubmit if revision is non-empty
				if job.EmptyRepoEmbeddingJob() {
					errs = errors.Append(errs, errors.Newf("Embedding job cannot be scheduled because the latest revision or default branch cannot be resolved for repo: %v", repoName))
					return nil
					//return errors.Newf("Embedding job cannot be scheduled because the latest revision or default branch cannot be resolved for %v", repoName)
				}

				if job.IsRepoEmbeddingJobScheduledOrCompleted() {
					errs = errors.Append(errs, errors.Newf("Embedding job is already scheduled or completed for repo %v at the latest revision %v", repoName, latestRevision))
					return nil
					//return errors.Newf("Embedding job is already scheduled or completed for the latest revision %v", latestRevision)
				}
			}

			_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision)
			return err
		}()
		if err != nil {
			return err
		}
	}
	return errs
}
