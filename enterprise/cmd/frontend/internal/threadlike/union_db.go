package threadlike

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func ThreadOrIssueOrChangesetByDBID(ctx context.Context, dbID int64) (graphqlbackend.ThreadOrIssueOrChangeset, error) {
	dbThread, err := internal.DBThreads{}.GetByID(ctx, dbID)
	if err != nil {
		return graphqlbackend.ThreadOrIssueOrChangeset{}, err
	}
	return newGQLThreadOrIssueOrChangeset(dbThread), nil
}
