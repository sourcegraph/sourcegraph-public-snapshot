package git

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// GitBackend is the interface through which operations on a git repository can
// be performed. It encapsulates the underlying git implementation and allows
// us to test out alternative backends.
// A GitBackend is expected to be scoped to a specific repository directory at
// initialization time, ie. it should not be shared across various repositories.
type GitBackend interface {
	// Config returns a backend for interacting with git configuration.
	Config() GitConfigBackend
	GetObject(ctx context.Context, objectName string) (*gitdomain.GitObject, error)
	// MergeBase finds the merge base commit for the given base and head SHAs.
	// Both baseSHA and headSHA are expected to be valid SHAs and are not validated
	// for safety.
	MergeBase(ctx context.Context, baseRevspec, headRevspec string) (api.CommitID, error)
}

// GitConfigBackend provides methods for interacting with git configuration.
type GitConfigBackend interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Unset(ctx context.Context, key string) error
}
