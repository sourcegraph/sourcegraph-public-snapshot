package git

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/zap/pkg/gitutil"
)

// ServerRepo is the minimal repo interface containing only those
// methods the server needs.
type ServerRepo interface {
	ReadBlob(snapshot, name string) ([]byte, string, string, error)
	ListTreeFull(string) (*gitutil.Tree, error)
	FileInfoForPath(rev, path string) (string, string, error)
	HashObject(typ, path string, data []byte) (string, error)
	CreateTree(basePath string, entries []*gitutil.TreeEntry) (string, error)
}

// ServerBackend creates server workspaces backed by a git
// repository. It implements the server.ServerBackend interface.
type ServerBackend struct {
	CanAccessRepo func(ctx context.Context, logger log.Logger, repo string) (bool, error)
}

// CanAccess implements server.ServerBackend.
func (s ServerBackend) CanAccess(ctx context.Context, logger log.Logger, repo string) (bool, error) {
	return s.CanAccessRepo(ctx, logger, repo)
}
