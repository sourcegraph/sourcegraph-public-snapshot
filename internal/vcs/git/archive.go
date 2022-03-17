package git

import (
	"context"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/actor"
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
			pathSpec, err := createPathSpecFromSubRepoPerms(ctx, checker, repo)
			if err != nil {
				return nil, errors.Wrap(err, "creating git pathspec from sub-repo perms")
			}
			options.Paths = pathSpec
		}
	}
	return gitserver.DefaultClient.Archive(ctx, repo, options)
}

// TODO: handle case where options.Paths is non-empty
func createPathSpecFromSubRepoPerms(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName) ([]string, error) {
	a := actor.FromContext(ctx)
	perms, err := authz.ActorRawPermissions(ctx, checker, a, repo)
	if err != nil {
		return []string{}, err
	}
	pathSpecs := make([]string, 0, len(perms.PathExcludes)+len(perms.PathIncludes))
	for _, p := range perms.PathIncludes {
		pathSpec := fmt.Sprintf(":(glob)%s", p)
		pathSpecs = append(pathSpecs, pathSpec)
	}
	for _, p := range perms.PathExcludes {
		pathSpec := fmt.Sprintf(":(glob,exclude)%s", p)
		pathSpecs = append(pathSpecs, pathSpec)
	}
	return pathSpecs, nil
}
