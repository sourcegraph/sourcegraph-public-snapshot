package migration

import "context"

type GitserverClient interface {
	DefaultBranchContains(ctx context.Context, repositoryID int, commit string) (bool, error)
}
