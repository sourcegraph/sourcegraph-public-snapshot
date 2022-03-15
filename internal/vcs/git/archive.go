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

	// maxPathLength is the maximum number of characters total in the requested paths that we can request from gitserver
	maxPathLength = 100000
)

// ArchiveReader streams back the file contents of an archived git repo.
func ArchiveReader(
	ctx context.Context,
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	options gitserver.ArchiveOptions,
) (io.ReadCloser, error) {
	options, err := validateOptions(options)
	if err != nil {
		return nil, err
	}
	if authz.SubRepoEnabled(checker) {
		if enabled, err := authz.SubRepoEnabledForRepo(ctx, checker, repo); err != nil {
			return nil, errors.Wrap(err, "sub-repo permissions check:")
		} else if enabled {
			filteredFiles, err := LsFiles(ctx, checker, repo, api.CommitID(options.Treeish), options.Paths...)
			if err != nil {
				return nil, errors.Wrap(err, "LsFiles in ArchiveReader")
			}
			if tooManyPaths(filteredFiles) {
				return nil, errors.Newf("too many files to pass to git archive: %d", len(filteredFiles))
			}
			options.Paths = filteredFiles
		}
	}
	return gitserver.DefaultClient.Archive(ctx, repo, options)
}

func validateOptions(opts gitserver.ArchiveOptions) (gitserver.ArchiveOptions, error) {
	if opts.Treeish == "" {
		return opts, errors.New("must provide a tree or commit to archive")
	}
	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}
	return opts, nil
}

func tooManyPaths(paths []string) bool {
	totalLen := 0
	for _, path := range paths {
		totalLen += len(path)
		if totalLen >= maxPathLength {
			return true
		}
	}
	return false
}
