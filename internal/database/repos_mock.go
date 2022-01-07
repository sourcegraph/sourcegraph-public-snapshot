package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type MockRepos struct {
	// TODO: we're knowingly taking on a little tech debt by placing these here for now.
	ListExternalServiceUserIDsByRepoID func(ctx context.Context, repoID api.RepoID) ([]int32, error)
	ListExternalServiceRepoIDsByUserID func(ctx context.Context, userID int32) ([]api.RepoID, error)
}
