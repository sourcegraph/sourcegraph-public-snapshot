package batches

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDisableBatchChangesJobs(t *testing.T) {
	os.Setenv("DISABLE_BATCH_CHANGES", "true")
	defer os.Unsetenv("DISABLE_BATCH_CHANGES")

	newJobFunc := []func() job.Job{
		NewJanitorJob,
		NewReconcilerJob,
		NewWorkspaceResolverJob,
		NewSchedulerJob,
	}

	for _, newJob := range newJobFunc {
		job := newJob()
		routines, err := job.Routines(context.Background(), &observation.TestContext)
		require.NoError(t, err)
		require.Nil(t, routines)
	}
}
