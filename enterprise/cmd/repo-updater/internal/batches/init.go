package batches

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// InitSyncRegistry sets up and starts a sync registry.
func InitSyncRegistry(
	ctx context.Context,
	db dbutil.DB,
	key encryption.Key,
	cf *httpcli.Factory,
) SyncRegistry {
	// We use an internal actor so that we can freely load dependencies from
	// the database without repository permissions being enforced.
	// We do check for repository permissions consciously in the Rewirer when
	// creating new changesets and in the executor, when talking to the code
	// host, we manually check for BatchChangesCredentials.
	ctx = actor.WithInternalActor(ctx)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	bstore := store.New(db, observationContext, key)

	syncRegistry := NewSyncRegistry(ctx, bstore, cf, observationContext)

	go goroutine.MonitorBackgroundRoutines(ctx, syncRegistry)

	return syncRegistry
}
