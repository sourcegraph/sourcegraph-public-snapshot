package campaigns

import (
	"context"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// maxWorkers defines the maximum number of changeset jobs to run in parallel.
var maxWorkers = env.Get("CAMPAIGNS_MAX_WORKERS", "8", "maximum number of repository jobs to run in parallel")

const defaultWorkerCount = 8

type GitserverClient interface {
	CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error)
}

// RunWorkers should be executed in a background goroutine and is responsible
// for finding pending ChangesetJobs and executing them.
// ctx should be canceled to terminate the function.
func RunWorkers(ctx context.Context, s *Store, clock func() time.Time, gitClient GitserverClient, sourcer repos.Sourcer, backoffDuration time.Duration) {
	workerCount, err := strconv.Atoi(maxWorkers)
	if err != nil {
		log15.Error("Parsing max worker count failed. Falling back to default.", "default", defaultWorkerCount, "err", err)
		workerCount = defaultWorkerCount
	}

	worker := func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(backoffDuration)
			}
		}
	}
	for i := 0; i < workerCount; i++ {
		go worker()
	}
}
