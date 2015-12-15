package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shazow/go-git"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
)

func init() {
	// Overwrite the git opener to return repositories that use the
	// gogits native-go implementation.
	vcs.RegisterOpener("git", func(dir string) (vcs.Repository, error) {
		return Open(dir)
	})
}

// Repository is a git VCS repository.
type Repository struct {
	*gitcmd.Repository

	repo *git.Repository
}

func (r *Repository) String() string {
	return fmt.Sprintf("git (gogit) repo at %s", r.RepoDir())
}

func Clone(url, dir string, opt vcs.CloneOpt) (*Repository, error) {
	_, err := gitcmd.Clone(url, dir, opt)
	if err != nil {
		return nil, err
	}
	// Note: This will call Open ~3 times as it jumps between
	// gitcmd -> gogit -> gitcmd until we replace it with a native version or
	// refactor.
	return Open(dir)
}

func Open(dir string) (*Repository, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		// Append .git to path
		dir = filepath.Join(dir, ".git")
	}

	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, &os.PathError{
			Op:   fmt.Sprintf("Open git repo [%s]", err.Error()),
			Path: dir,
			Err:  os.ErrNotExist,
		}
	}

	return &Repository{
		Repository: &gitcmd.Repository{Dir: dir},
		repo:       repo,
	}, nil
}

// FileSystem opens the repository file tree at a given commit ID.
func (r *Repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	ci, err := r.repo.GetCommit(string(at))
	if err != nil {
		return nil, err
	}
	return &filesystem{
		dir:  r.repo.Path,
		oid:  string(at),
		tree: &ci.Tree,
		repo: r.repo,
	}, nil
}
