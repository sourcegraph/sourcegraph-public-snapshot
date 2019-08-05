package extsvc

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal/extsvc"
)

// Refresh refreshes information about the thread from external services (if any).
func Refresh(ctx context.Context, dbID int64) error {
	dbThread, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if dbThread.ImportedFromExternalServiceID == 0 {
		return nil // no associated external services
	}
	return extsvc.ImportGitHubThreadEvents(ctx, dbID, dbThread.ImportedFromExternalServiceID, dbThread.ExternalID, dbThread.RepositoryID)
}
