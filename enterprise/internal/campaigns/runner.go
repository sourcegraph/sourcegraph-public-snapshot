package campaigns

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"gopkg.in/inconshreveable/log15.v2"
)

// maxWorkers defines the maximum number of repository jobs to run in parallel.
var maxWorkers = env.Get("A8N_MAX_WORKERS", "8", "maximum number of repository jobs to run in parallel")

const defaultWorkerCount = 8

// RunChangesetJobs should run in a background goroutine and is responsible
// for finding pending jobs and running them.
// ctx should be canceled to terminate the function
func RunChangesetJobs(ctx context.Context, s *Store, clock func() time.Time, gitClient GitserverClient, backoffDuration time.Duration) {
	workerCount, err := strconv.Atoi(maxWorkers)
	if err != nil {
		log15.Error("Parsing max worker count failed. Falling back to default.", "default", defaultWorkerCount, "err", err)
		workerCount = defaultWorkerCount
	}
	process := func(ctx context.Context, s *Store, job campaigns.ChangesetJob) error {
		c, err := s.GetCampaign(ctx, GetCampaignOpts{
			ID: job.CampaignID,
		})
		if err != nil {
			return errors.Wrap(err, "getting campaign")
		}
		if runErr := RunChangesetJob(ctx, clock, s, gitClient, nil, c, &job); runErr != nil {
			log15.Error("RunChangesetJob", "jobID", job.ID, "err", err)
		}
		// We don't assign to err here so that we don't roll back the transaction
		// RunChangesetJob will save the error in the job row
		return nil
	}
	worker := func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				didRun, err := s.ProcessPendingChangesetJobs(context.Background(), process)
				if err != nil {
					log15.Error("Running changeset job", "err", err)
				}
				// Back off on error or when no jobs available
				if err != nil || !didRun {
					time.Sleep(backoffDuration)
				}
			}
		}
	}
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}
