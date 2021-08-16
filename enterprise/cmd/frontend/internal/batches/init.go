package batches

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/migrations"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Init initializes the given enterpriseServices to include the required
// resolvers for Batch Changes and sets up webhook handlers for changeset
// events.
func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	// Validate site configuration.
	conf.ContributeValidator(func(c conf.Unified) (problems conf.Problems) {
		if _, err := window.NewConfiguration(c.BatchChangesRolloutWindows); err != nil {
			problems = append(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize store.
	bstore := store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey)

	// Register enterprise services.
	enterpriseServices.BatchChangesResolver = resolvers.New(bstore)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(bstore, observationContext)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(bstore, observationContext)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(bstore, observationContext)

	// Register Batch Changes OOB migrations.
	return migrations.Register(bstore, outOfBandMigrationRunner)
}
