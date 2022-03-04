package git

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// ArchiveFormatZip indicates a zip archive is desired.
	ArchiveFormatZip = "zip"

	// ArchiveFormatTar indicates a tar archive is desired.
	ArchiveFormatTar = "tar"
)

// ArchiveReader streams back the file contents of an archived git repo.
func ArchiveReader(
	ctx context.Context,
	repoName api.RepoName,
	options gitserver.ArchiveOptions,
) (io.ReadCloser, error) {
	return gitserver.DefaultClient.Archive(ctx, repoName, options)
}

func ArchiveReaderWithSubRepo(
	ctx context.Context,
	checker authz.SubRepoPermissionChecker,
	repo *types.Repo,
	options gitserver.ArchiveOptions,
) (io.ReadCloser, error) {
	if authz.SubRepoEnabled(checker) {
		enabled, err := authz.SubRepoEnabledForRepoID(ctx, checker, repo.ID)
		if err != nil {
			return nil, errors.Wrap(err, "sub-repo permissions check:")
		}
		if enabled {
			return nil, errors.New("archiveReader invoked for a repo with sub-repo permissions")
		}
	}
	return ArchiveReader(ctx, repo.Name, options)
}
