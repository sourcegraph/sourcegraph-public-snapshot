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
) (errs error) {
	// we track skipped repo errors in the errs return value so the transaction should not rollback based on skip errors.
	// we need a separate error to pass to Done()
	var txErr error

	tx, txErr := repoEmbeddingJobsStore.Transact(ctx)
	if txErr != nil {
		return
	}
	defer func() {
		txErr = tx.Done(txErr)

		// if we did not commit the transaction then that error is the only value to return
		if txErr != nil {
			errs = txErr
		}
	}()

	repoStore := db.Repos()
	for _, repoName := range repoNames {
		// Scope the iteration to an anonymous function, so we can capture all errors and properly rollback tx in defer above.
		txErr = func() error {
			r, err := repoStore.GetByName(ctx, repoName)
			if err != nil {
				errs = errors.Append(errs, errors.Wrap(err, "getting repo by name"))
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
				}

				if job.IsRepoEmbeddingJobScheduledOrCompleted() {
					errs = errors.Append(errs, errors.Newf("Embedding job is already scheduled or completed for repo %v at the latest revision %v", repoName, latestRevision.Short()))
					return nil
				}
			}

			_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision)
			return err
		}()
		if txErr != nil {
			return
		}
	}
	return
}
