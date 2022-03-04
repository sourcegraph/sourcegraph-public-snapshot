package git

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	options gitserver.ArchiveOptions,
) (io.ReadCloser, error) {
	if authz.SubRepoEnabled(checker) {
		if enabled, err := authz.SubRepoEnabledForRepo(ctx, checker, repo); err != nil {
			return nil, errors.Wrap(err, "sub-repo permissions check:")
		} else if enabled {
			// todo: handle case if options nil or if Treeish empty
			if len(options.Paths) == 0 {
				options.Paths = []string{"."}
			}
			filteredFiles, err := LsFiles(ctx, checker, repo, api.CommitID(options.Treeish), options.Paths...)
			if err != nil {
				return nil, errors.Wrap(err, "LsFiles in ArchiveReader")
			}
			options.Paths = filteredFiles
		}
	}
	return gitserver.DefaultClient.Archive(ctx, repo, options)
}
