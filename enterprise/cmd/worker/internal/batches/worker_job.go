package batches

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/internal/batches/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewBatchesJob() shared.Job {
	return &batchesJob{}
}

type batchesJob struct{}

func (s *batchesJob) Config() []env.Config {
	return []env.Config{}
}

// TODO: The passed down context is canceled once this function returns, so it's
// not safe to use it for our purposes.
func (s *batchesJob) Routines(_ context.Context) ([]goroutine.BackgroundRoutine, error) {
	// No Batch Changes on dotcom, so we don't need to spawn the
	// background jobs for this feature.
	if envvar.SourcegraphDotComMode() {
		log15.Info("Found dot-com instance, not starting Batch Changes routines")
		return []goroutine.BackgroundRoutine{}, nil
	}

	// We use an internal actor so that we can freely load dependencies from
	// the database without repository permissions being enforced.
	// We do check for repository permissions consciously in the Rewirer when
	// creating new changesets and in the executor, when talking to the code
	// host, we manually check for BatchChangesCredentials.
	ctx := actor.WithInternalActor(context.Background())

	db, err := shared.InitDatabase()
	if err != nil {
		return nil, err
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	bstore := store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey)

	return background.Routines(ctx, bstore, httpcli.NewExternalClientFactory(), observationContext), nil
}
