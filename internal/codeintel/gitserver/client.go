package gitserver

import "github.com/sourcegraph/sourcegraph/internal/codeintel/db"

// Client is an interface that wraps all of the queries to gitserver needed by the
// precise-code-intel services.
type Client interface {
	// Head determines the tip commit of the default branch for the given repository.
	Head(db db.DB, repositoryID int) (string, error)

	// CommitsNear returns a map from a commit to parent commits. The commits populating the
	// map are the MaxCommitsPerUpdate closest ancestors from the given commit.
	CommitsNear(db db.DB, repositoryID int, commit string) (map[string][]string, error)

	// DirectoryChildren determines all children known to git for the given directory names via an invocation
	// of git ls-tree. The keys of the resulting map are the input (unsanitized) dirnames, and the value of
	// that key are the files nested under that directory.
	DirectoryChildren(db db.DB, repositoryID int, commit string, dirnames []string) (map[string][]string, error)
}

type defaultClient struct{}

var DefaultClient Client = &defaultClient{}

func (c *defaultClient) Head(db db.DB, repositoryID int) (string, error) {
	return Head(db, repositoryID)
}

func (c *defaultClient) CommitsNear(db db.DB, repositoryID int, commit string) (map[string][]string, error) {
	return CommitsNear(db, repositoryID, commit)
}

func (c *defaultClient) DirectoryChildren(db db.DB, repositoryID int, commit string, dirnames []string) (map[string][]string, error) {
	return DirectoryChildren(db, repositoryID, commit, dirnames)
}
