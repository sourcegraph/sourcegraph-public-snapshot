package git

import (
	"context"
)

// SetRepositoryType sets the type of the repository.
func SetRepositoryType(ctx context.Context, git GitConfigBackend, typ string) error {
	return git.Set(ctx, "sourcegraph.type", typ)
}

// GetRepositoryType returns the type of the repository.
func GetRepositoryType(ctx context.Context, git GitConfigBackend) (string, error) {
	val, err := git.Get(ctx, "sourcegraph.type")
	if err != nil {
		return "", err
	}
	return val, nil
}
