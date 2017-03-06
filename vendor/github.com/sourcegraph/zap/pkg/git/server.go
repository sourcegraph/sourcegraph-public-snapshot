package git

import (
	"context"
	"fmt"

	logpkg "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/sourcegraph/zap/ot"
	"github.com/sourcegraph/zap/pkg/gitutil"
	"github.com/sourcegraph/zap/ws"
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
// repository. It implements the zap.ServerBackend interface.
type ServerBackend struct {
	// OpenBareRepo is called to open a bare git repository. The value of
	// repo is opaque to zap and only needs to be created and
	// interpreted by the client and by this OpenBareRepo func.
	OpenBareRepo func(ctx context.Context, log *logpkg.Context, repo string) (ServerRepo, error)

	CanAccessRepo     func(ctx context.Context, log *logpkg.Context, repo string) (bool, error)
	CanAutoCreateRepo func() bool
}

// Create implements zap.ServerBackend.Create.
func (s ServerBackend) Create(ctx context.Context, log *logpkg.Context, repo, base string) (*ws.Proxy, error) {
	if repo == "" {
		panic("empty repo")
	}
	if base == "" {
		panic("empty base")
	}

	level.Debug(log).Log("create-repo", repo+"@"+base)

	if ok, err := s.CanAccess(ctx, log, repo); err != nil {
		return nil, fmt.Errorf("access check for repo %q: %s", repo, err)
	} else if !ok {
		return nil, fmt.Errorf("access denied for client to repo %q", repo)
	}

	gitRepo, err := s.OpenBareRepo(ctx, log, repo)
	if err != nil {
		return nil, err
	}

	var fbuf FileBuffer
	snapshot := base
	return &ws.Proxy{
		Apply: func(log *logpkg.Context, op ot.WorkspaceOp) error {
			prevSnapshot := snapshot

			// fmt.Fprintf(os.Stderr, "# server workspace: applying op %s on top of snapshot %s\n", op, prevSnapshot)
			fbufCopy := fbuf.Copy().(*FileBuffer)
			newGitSnapshot, err := CreateTreeForOp(log, gitRepo, fbufCopy, prevSnapshot, op)
			if err != nil {
				return err
			}
			fbuf = *fbufCopy
			if newGitSnapshot != "" {
				snapshot = newGitSnapshot
			}
			// fmt.Fprintf(os.Stderr, "# server workspace snapshot: %s → %s\n", prevSnapshot, newGitSnapshot)

			return nil
		},
	}, nil
}

// CanAccess implements zap.ServerBackend.
func (s ServerBackend) CanAccess(ctx context.Context, log *logpkg.Context, repo string) (bool, error) {
	return s.CanAccessRepo(ctx, log, repo)
}

// CanAutoCreate implements zap.ServerBackend.
func (s ServerBackend) CanAutoCreate() bool {
	return s.CanAutoCreateRepo()
}
