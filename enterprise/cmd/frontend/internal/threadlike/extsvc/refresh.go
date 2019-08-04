package extsvc

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal/extsvc"
)

// Refresh refreshes information about the thread from external services (if any).
func Refresh(ctx context.Context, dbID int64) error {
	dbThread, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if dbThread.ExternalURL == nil {
		return nil // no associated external services
	}

	repo, err := graphqlbackend.RepositoryByDBID(ctx, dbThread.RepositoryID)
	if err != nil {
		return err
	}
	return extsvc.ImportGitHubThreadEvents(ctx, dbID, repo.DBID(), repo.DBExternalRepo(), *dbThread.ExternalURL)
}
