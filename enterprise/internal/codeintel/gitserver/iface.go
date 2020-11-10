package gitserver

import "context"

type DBStore interface {
	RepoName(ctx context.Context, repositoryID int) (string, error)
}
