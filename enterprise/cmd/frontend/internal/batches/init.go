package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var config = &store.Config{}

func init() {
	config.Load()
}

// Init initializes the given enterpriseServices to include the required
// resolvers for Batch Changes and sets up webhook handlers for changeset
// events.
func Init(ctx context.Context, db database.DB, _ conftypes.UnifiedWatchable, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	// Validate site configuration.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if _, err := window.NewConfiguration(c.SiteConfig().BatchChangesRolloutWindows); err != nil {
			problems = append(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})

	// Initialize store.
	cstore := store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey)

	// Initialize upload store
	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "failed to load batches config")
	}
	uploadStore, err := store.NewUploadStore(ctx, config, observationContext)
	if err != nil {
		return errors.Wrap(err, "initialize upload store")
	}

	// Register enterprise services.
	enterpriseServices.BatchChangesResolver = resolvers.New(cstore)
	enterpriseServices.GitHubWebhook = webhooks.NewGitHubWebhook(cstore)
	enterpriseServices.BitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(cstore)
	enterpriseServices.BitbucketCloudWebhook = webhooks.NewBitbucketCloudWebhook(cstore)
	enterpriseServices.GitLabWebhook = webhooks.NewGitLabWebhook(cstore)
	operations := httpapi.NewOperations(observationContext)
	enterpriseServices.BatchesMountUploadHandler = httpapi.NewMountUploadHandler(cstore, uploadStore, operations)
	enterpriseServices.BatchesMountRetrievalHandler = httpapi.NewMountRetrievalHandler(cstore, uploadStore, operations)

	return nil
}
