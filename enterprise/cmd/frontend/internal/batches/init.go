package batches

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Init initializes the given enterpriseServices to include the required
// resolvers for Batch Changes and sets up webhook handlers for changeset
// events.
func Init(
	ctx context.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	observationContext *observation.Context,
) error {
	// Validate site configuration.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if _, err := window.NewConfiguration(c.SiteConfig().BatchChangesRolloutWindows); err != nil {
			problems = append(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})

	// Initialize store.
	bstore := store.New(db, observationContext, keyring.Default().BatchChangesCredentialKey)

	// Register enterprise services.
	gitserverClient := gitserver.NewClient(db)
	enterpriseServices.BatchChangesResolver = resolvers.New(bstore, gitserverClient)
	enterpriseServices.BatchesGitHubWebhook = webhooks.NewGitHubWebhook(bstore, gitserverClient)
	enterpriseServices.BatchesBitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(bstore, gitserverClient)
	enterpriseServices.BatchesBitbucketCloudWebhook = webhooks.NewBitbucketCloudWebhook(bstore, gitserverClient)
	enterpriseServices.BatchesGitLabWebhook = webhooks.NewGitLabWebhook(bstore, gitserverClient)

	operations := httpapi.NewOperations(observationContext)
	fileHandler := httpapi.NewFileHandler(db, bstore, operations)
	enterpriseServices.BatchesChangesFileGetHandler = fileHandler.Get()
	enterpriseServices.BatchesChangesFileExistsHandler = fileHandler.Exists()
	enterpriseServices.BatchesChangesFileUploadHandler = fileHandler.Upload()

	return nil
}
