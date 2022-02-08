package git

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ArchiveFormat represents an archive format (zip, tar, etc).
type ArchiveFormat string

const (
	// ArchiveFormatZip indicates a zip archive is desired.
	ArchiveFormatZip ArchiveFormat = "zip"

	// ArchiveFormatTar indicates a tar archive is desired.
	ArchiveFormatTar ArchiveFormat = "tar"
)

// ArchiveReader streams back the file contents of an archived git repo.
func ArchiveReader(
	ctx context.Context,
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	format ArchiveFormat,
	commit api.CommitID,
	relativePath string,
) (io.ReadCloser, error) {
	a := actor.FromContext(ctx)
	if authz.SubRepoEnabled(checker) && a.IsAuthenticated() {
		return nil, errors.New("ArchiveReader invoked by user on a repo with sub-repo permissions")
	}
	cmd := gitserver.DefaultClient.Command("git", "archive", "--format="+string(format), string(commit), relativePath)
	cmd.Repo = repo
	return gitserver.StdoutReader(ctx, cmd)
}
