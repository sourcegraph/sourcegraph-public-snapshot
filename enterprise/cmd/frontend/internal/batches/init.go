package batches

import (
	"context"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/httpapi"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/batches/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types/scheduler/window"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
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
	observationCtx *observation.Context,
	db database.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	// Validate site configuration.
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if _, err := window.NewConfiguration(c.SiteConfig().BatchChangesRolloutWindows); err != nil {
			problems = append(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})

	// Initialize store.
	bstore := store.New(db, observationCtx, keyring.Default().BatchChangesCredentialKey)

	// Register enterprise services.
	gitserverClient := gitserver.NewClient()
	logger := sglog.Scoped("Batches", "batch changes webhooks")
	enterpriseServices.BatchChangesResolver = resolvers.New(edb.NewEnterpriseDB(db), bstore, gitserverClient, logger)
	enterpriseServices.BatchesGitHubWebhook = webhooks.NewGitHubWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BatchesBitbucketServerWebhook = webhooks.NewBitbucketServerWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BatchesBitbucketCloudWebhook = webhooks.NewBitbucketCloudWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BatchesGitLabWebhook = webhooks.NewGitLabWebhook(bstore, gitserverClient, logger)
	enterpriseServices.BatchesAzureDevOpsWebhook = webhooks.NewAzureDevOpsWebhook(bstore, gitserverClient, logger)

	operations := httpapi.NewOperations(observationCtx)
	fileHandler := httpapi.NewFileHandler(db, bstore, operations)
	enterpriseServices.BatchesChangesFileGetHandler = fileHandler.Get()
	enterpriseServices.BatchesChangesFileExistsHandler = fileHandler.Exists()
	enterpriseServices.BatchesChangesFileUploadHandler = fileHandler.Upload()

	return nil
}
