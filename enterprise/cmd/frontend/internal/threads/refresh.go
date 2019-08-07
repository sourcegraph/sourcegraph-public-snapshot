package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads/internal/extsvc"
)

// Refresh refreshes information about the thread from external services (if any).
func Refresh(ctx context.Context, dbID int64) error {
	dbThread, err := dbThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if dbThread.ImportedFromExternalServiceID == 0 {
		return nil // no associated external services
	}
	return extsvc.ImportGitHubThreadEvents(ctx, dbID, dbThread.ImportedFromExternalServiceID, dbThread.ExternalID, dbThread.RepositoryID)
}
