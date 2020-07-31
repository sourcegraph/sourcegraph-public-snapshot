package gitserver

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// Client is an interface that wraps all of the queries to gitserver needed by the
// precise-code-intel services.
type Client interface {
	// Head determines the tip commit of the default branch for the given repository.
	Head(ctx context.Context, store store.Store, repositoryID int) (string, error)

	// CommitGraph returns the commit graph for the given repository as a mapping from a commit
	// to its parents.
	CommitGraph(ctx context.Context, store store.Store, repositoryID int) (map[string][]string, error)

	// DirectoryChildren determines all children known to git for the given directory names via an invocation
	// of git ls-tree. The keys of the resulting map are the input (unsanitized) dirnames, and the value of
	// that key are the files nested under that directory.
	DirectoryChildren(ctx context.Context, store store.Store, repositoryID int, commit string, dirnames []string) (map[string][]string, error)

	// Archive retrieves a tar-formatted archive of the given commit.
	Archive(ctx context.Context, store store.Store, repositoryID int, commit string) (io.Reader, error)

	// FileExists determines whether a file exists in a particular commit of a repository.
	FileExists(ctx context.Context, store store.Store, repositoryID int, commit, file string) (bool, error)

	// Tags returns the git tags associated with the given commit along with a boolean indicating whether
	// or not the tag was attached directly to the commit. If no tags exist at or before this commit, the
	// tag is an empty string.
	Tags(ctx context.Context, store store.Store, repositoryID int, commit string) (string, bool, error)
}

type defaultClient struct{}

var DefaultClient Client = &defaultClient{}

func (c *defaultClient) Head(ctx context.Context, store store.Store, repositoryID int) (string, error) {
	return Head(ctx, store, repositoryID)
}

func (c *defaultClient) CommitGraph(ctx context.Context, store store.Store, repositoryID int) (map[string][]string, error) {
	return CommitGraph(ctx, store, repositoryID)
}

func (c *defaultClient) DirectoryChildren(ctx context.Context, store store.Store, repositoryID int, commit string, dirnames []string) (map[string][]string, error) {
	return DirectoryChildren(ctx, store, repositoryID, commit, dirnames)
}

func (c *defaultClient) Archive(ctx context.Context, store store.Store, repositoryID int, commit string) (io.Reader, error) {
	return Archive(ctx, store, repositoryID, commit)
}

func (c *defaultClient) FileExists(ctx context.Context, store store.Store, repositoryID int, commit, file string) (bool, error) {
	return FileExists(ctx, store, repositoryID, commit, file)
}

func (c *defaultClient) Tags(ctx context.Context, store store.Store, repositoryID int, commit string) (string, bool, error) {
	return Tags(ctx, store, repositoryID, commit)
}
