package testutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// CreateThread creates a thread in the DB, for use in tests only.
func CreateThread(ctx context.Context, title string, repositoryID api.RepoID, authorUserID int32) (id int64, err error) {
	thread, err := internal.DBThreads{}.Create(ctx,
		&internal.DBThread{
			Type:         internal.DBThreadTypeThread,
			RepositoryID: repositoryID,
			Title:        title,
		},
		commentobjectdb.DBObjectCommentFields{AuthorUserID: authorUserID},
	)
	if err != nil {
		return 0, err
	}
	return thread.ID, nil
}
