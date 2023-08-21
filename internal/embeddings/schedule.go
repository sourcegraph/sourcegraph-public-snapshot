package embeddings

import (
	"context"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/embed"
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

	embeddingsConf := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	c, err := embed.NewEmbeddingsClient(embeddingsConf)
	if err != nil {
		return errors.Wrap(err, "getting embeddings client")
	}

	modelID := c.GetModelIdentifier()

	defer func() { err = tx.Done(err) }()

	repoStore := db.Repos()
	for _, repoName := range repoNames {
		// Scope the iteration to an anonymous function, so we can capture all errors and properly rollback tx in defer above.
		err = func() error {
			r, err := repoStore.GetByName(ctx, repoName)
			if err != nil {
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

				// Skip if another job is scheduled or completed for this revision
				duplicate := job.IsRepoEmbeddingJobScheduledOrCompleted()

				// Skip if a job previously failed for empty revision and our fetched revision has not changed
				skipEmptyRepo := job.EmptyRepoEmbeddingJob()

				// Skip if we know the provider and model has not changed.
				// If our previous job has an empty model identifier then schedule a job to identify provider & model from
				// embedding index metadata.
				modelIsCurrent := job.ModelID == modelID

				if (duplicate || skipEmptyRepo) && modelIsCurrent {
					return nil
				}
			}

			_, err = tx.CreateRepoEmbeddingJob(ctx, r.ID, latestRevision, modelID)
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
