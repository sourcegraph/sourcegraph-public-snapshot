package intel

import (
	"context"
	"path"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PathState int

const (
	PathIsFile PathState = iota
	PathIsDirectory
)

type ErrPathDoesNotExist struct{}

func (ErrPathDoesNotExist) Error() string  { return "path does not exist" }
func (ErrPathDoesNotExist) NotFound() bool { return true }

func (s *IntelService) PathExists(ctx context.Context, repo, commit, path string) (PathState, error) {
	resp, err := pathExists(ctx, s.client, repo, commit, path)
	if err != nil {
		return 0, errors.Wrap(err, "checking if path exists")
	}

	if resp.Repository.Commit.Path == nil {
		return 0, ErrPathDoesNotExist{}
	}

	commitPath := resp.Repository.Commit.Path.(interface{ GetIsDirectory() bool })
	if commitPath.GetIsDirectory() {
		return PathIsDirectory, nil
	}
	return PathIsFile, nil
}

func (s *IntelService) FindWorkspaceRoot(ctx context.Context, repo, commit, fullPath string) (string, error) {
	// TODO: support anything but Go.
	for dir := path.Dir(fullPath); dir != "."; dir = path.Dir(dir) {
		target := path.Join(dir, "go.mod")
		fileType, err := s.PathExists(ctx, repo, commit, target)
		if err != nil && !errcode.IsNotFound(err) {
			return "", errors.Wrap(err, "finding workspace root")
		}
		if err == nil {
			if fileType != PathIsFile {
				return "", errors.Newf("unexpected file type at %q: %v", target, fileType)
			}
			return dir, nil
		}
	}

	return "", nil
}
