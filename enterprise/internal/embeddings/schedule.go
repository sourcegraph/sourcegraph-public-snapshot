package embeddings

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ScheduleRepositoriesForEmbedding(
	ctx context.Context,
	repoNames []api.RepoName,
	forceReindex bool,
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
		// Scope the iteration to an anonymous function, so we can capture all errors and properly rollback tx in defer above.
		err = func() error {
			r, err := repoStore.GetByName(ctx, repoName)
			if err != nil {
				return err
			}

			refName, latestRevision, err := gitserverClient.GetDefaultBranch(ctx, r.Name, false)
			if err != nil {
				return err
			}
			if refName == "" {
				return errors.Newf("could not get latest commit for repo %s", r.Name)
			}

			// If the user has forced a reindex, then we always start a new job. Otherwise, we skip creating a job for
			// a revision if there's already an identical job that's completed or scheduled to run.
			if !forceReindex {
				job, _ := tx.GetLastRepoEmbeddingJobForRevision(ctx, r.ID, latestRevision)
				if job.IsRepoEmbeddingJobScheduledOrCompleted() {
					return nil
				}
			}

			_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision)
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
