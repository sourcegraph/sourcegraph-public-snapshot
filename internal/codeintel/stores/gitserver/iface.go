package gitserver

import "context"

type DBStore interface {
	RepoName(ctx context.Context, repositoryID int) (string, error)
	RepoNames(ctx context.Context, repositoryIDs ...int) (map[int]string, error)
}
