package batches

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/syncer"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// InitBackgroundJobs starts all jobs required to run batches. Currently, it is called from
// repo-updater and in the future will be the main entry point for the batch changes worker.
func InitBackgroundJobs(
	ctx context.Context,
	db database.DB,
	key encryption.Key,
	cf *httpcli.Factory,
) (batches.ChangesetSyncRegistry, batches.RepoMetadataSyncer) {
	// We use an internal actor so that we can freely load dependencies from
	// the database without repository permissions being enforced.
	// We do check for repository permissions consciously in the Rewirer when
	// creating new changesets and in the executor, when talking to the code
	// host, we manually check for BatchChangesCredentials.
	ctx = actor.WithInternalActor(ctx)

	observationContext := &observation.Context{
		Logger:     log.Scoped("batches.background", "batches background jobs"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	bstore := store.New(db, observationContext, key)

	syncRegistry := syncer.NewSyncRegistry(ctx, bstore, cf, observationContext)

	go goroutine.MonitorBackgroundRoutines(ctx, syncRegistry)

	return syncRegistry, &repoMetadataSyncer{bstore}
}

type repoMetadataSyncer struct {
	store *store.Store
}

func (s *repoMetadataSyncer) EnqueueRepos(ctx context.Context, ids []api.RepoID) error {
	return s.store.EnqueueRepoMetadataSyncs(ctx, ids)
}
