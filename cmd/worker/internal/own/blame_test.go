package own_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/worker/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRunBackgroundJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	observationContext := observation.TestContextTB(t)
	ctx := actor.WithInternalActor(context.Background())

	job := own.NewBlameJob()
	routines, err := job.Routines(ctx, observationContext)
	require.NoError(t, err)
	for _, r := range routines {
		go r.Start()
		defer r.Stop()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	id, err := own.Enqueue(ctx, db, 3, "/foo/bar/baz")
	require.NoError(t, err)
	var state string
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second * 1)
		state, err = own.State(ctx, db, id)
		require.NoError(t, err)
		if state != "queued" && state != "processing" {
			break
		}
	}
	assert.Equal(t, "completed", state)
}
