package git

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

// GitBackend is the interface through which operations on a git repository can
// be performed. It encapsulates the underlying git implementation and allows
// us to test out alternative backends.
// A GitBackend is expected to be scoped to a specific repository directory at
// initialization time, ie. it should not be shared across various repositories.
type GitBackend interface {
	// Config returns a backend for interacting with git configuration at .git/config.
	Config() GitConfigBackend
	// GetObject allows to read a git object from the git object database.
	GetObject(ctx context.Context, objectName string) (*gitdomain.GitObject, error)
	// MergeBase finds the merge base commit for the given base and head revspecs.
	// Returns an empty string and no error if no common merge-base was found.
	MergeBase(ctx context.Context, baseRevspec, headRevspec string) (api.CommitID, error)
	// Blame returns a reader for the blame info of the given path.
	// BlameHunkReader must always be closed.
	Blame(ctx context.Context, path string, opt BlameOptions) (BlameHunkReader, error)
	// SymbolicRefHead resolves what the HEAD symbolic ref points to. This is also
	// commonly referred to as the default branch within Sourcegraph.
	// If short is true, the returned ref name will be shortened when possible
	// without ambiguity.
	SymbolicRefHead(ctx context.Context, short bool) (string, error)
	// RevParseHead resolves at what commit HEAD points to. If HEAD doesn't point
	// to anything, a RevisionNotFoundError is returned. This can occur, for example,
	// when the repository is empty (ie. has no commits).
	RevParseHead(ctx context.Context) (api.CommitID, error)

	// Exec is a temporary helper to run arbitrary git commands from the exec endpoint.
	// No new usages of it should be introduced and once the migration is done we will
	// remove this method.
	Exec(ctx context.Context, args ...string) (io.ReadCloser, error)
}

// GitConfigBackend provides methods for interacting with git configuration.
type GitConfigBackend interface {
	// Get reads a given config value. If the value is not set, it returns an
	// empty string and no error.
	Get(ctx context.Context, key string) (string, error)
	// Set sets a config value for the given key.
	Set(ctx context.Context, key, value string) error
	// Unset removes a config value of the given key. If the key wasn't present,
	// no error is returned.
	Unset(ctx context.Context, key string) error
}

// BlameOptions are options for git blame.
type BlameOptions struct {
	NewestCommit     api.CommitID
	IgnoreWhitespace bool
	// 1-indexed start line (or 0 for beginning of file)
	StartLine int
	// 1-indexed end line (or 0 for end of file)
	EndLine int
}

// BlameHunkReader is a reader for git blame hunks.
type BlameHunkReader interface {
	// Consume the next hunk. io.EOF is returned at the end of the stream.
	Read() (*gitdomain.Hunk, error)
	Close() error
}
