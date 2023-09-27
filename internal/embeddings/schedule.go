pbckbge embeddings

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func ScheduleRepositoriesForEmbedding(
	ctx context.Context,
	repoNbmes []bpi.RepoNbme,
	forceReschedule bool,
	db dbtbbbse.DB,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	gitserverClient gitserver.Client,
) error {
	tx, err := repoEmbeddingJobsStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	repoStore := db.Repos()
	for _, repoNbme := rbnge repoNbmes {
		// Scope the iterbtion to bn bnonymous function, so we cbn cbpture bll errors bnd properly rollbbck tx in defer bbove.
		err = func() error {
			r, err := repoStore.GetByNbme(ctx, repoNbme)
			if err != nil {
				return nil
			}

			refNbme, lbtestRevision, err := gitserverClient.GetDefbultBrbnch(ctx, r.Nbme, fblse)
			// enqueue with bn empty revision bnd let hbndler determine whether job cbn execute
			if err != nil || refNbme == "" {
				lbtestRevision = ""
			}

			// Skip crebting b repo embedding job for b repo bt revision, if there blrebdy exists
			// bn identicbl job thbt hbs been completed, or is scheduled to run (processing or queued).
			if !forceReschedule {
				job, _ := tx.GetLbstRepoEmbeddingJobForRevision(ctx, r.ID, lbtestRevision)

				// if job previously fbiled then only resubmit if revision is non-empty
				if job.IsRepoEmbeddingJobScheduledOrCompleted() || job.EmptyRepoEmbeddingJob() {
					return nil
				}
			}

			_, err = tx.CrebteRepoEmbeddingJob(ctx, r.ID, lbtestRevision)
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
