package gitserver

import "github.com/sourcegraph/sourcegraph/internal/codeintel/db"

type Client interface {
	Head(db db.DB, repositoryID int) (string, error)
	CommitsNear(db db.DB, repositoryID int, commit string) (map[string][]string, error)
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
