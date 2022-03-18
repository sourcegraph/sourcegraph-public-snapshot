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
			if shouldFilterPaths(options.Paths) {
				filteredPaths, err := filterRequestedPaths(ctx, checker, repo, options.Treeish, options.Paths)
				if err != nil {
					return nil, errors.Wrap(err, "error filtering the requested paths")
				}
				options.Paths = filteredPaths
			} else {
				pathSpec, err := authz.CreateGitPathSpecFromSubRepoPerms(ctx, checker, repo)
				if err != nil {
					return nil, errors.Wrap(err, "creating git pathspec from sub-repo perms")
				}
				options.Paths = pathSpec
			}
		}
	}
	return gitserver.DefaultClient.Archive(ctx, repo, options)
}

func shouldFilterPaths(paths []string) bool {
	if len(paths) == 0 {
		return false
	}
	if len(paths) == 1 && paths[0] == "." {
		return false
	}
	return true
}

func filterRequestedPaths(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit string, requestedPaths []string) ([]string, error) {
	if commit == "" {
		return []string{}, errors.New("empty commit id")
	}
	// Call LsFiles which will list the files requested and filter out any the user doesn't have access to.
	return LsFiles(ctx, checker, repo, api.CommitID(commit), requestedPaths...)
}
