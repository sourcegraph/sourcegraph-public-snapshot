package threads

import (
	"context"
)

// Refresh refreshes information about the thread from external services (if any).
func Refresh(ctx context.Context, dbID int64) error {
	dbThread, err := dbThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if dbThread.ExternalServiceID == 0 {
		return nil // no associated external services
	}

	// Refresh this thread's metadata.
	client, _, err := getClientForRepo(ctx, dbThread.RepositoryID)
	if err != nil {
		return nil
	}
	if err := client.RefreshThreadMetadata(ctx, dbThread.ID, dbThread.ExternalServiceID, dbThread.ExternalID, dbThread.RepositoryID); err != nil {
		return err
	}

	return ImportGitHubThreadEvents(ctx, dbID, dbThread.ExternalServiceID, dbThread.ExternalID, dbThread.RepositoryID)
}
