package gitcli

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func (g *gitCLIBackend) ArchiveReader(ctx context.Context, format, treeish string, pathspecs []string) (io.ReadCloser, error) {
	// This is a long time, but this never blocks a user request for this
	// long. Even repos that are not that large can take a long time, for
	// example a search over all repos in an organization may have several
	// large repos. All of those repos will be competing for IO => we need
	// a larger timeout.
	ctx, cancel := context.WithTimeout(ctx, conf.GitLongCommandTimeout())

	archiveArgs := buildArchiveArgs(format, treeish, pathspecs)
	cmd, _, err := g.gitCommand(ctx, archiveArgs...)
	if err != nil {
		cancel()
		return nil, err
	}

	r, err := g.runGitCommand(ctx, cmd)
	if err != nil {
		cancel()
		return nil, err
	}

	return &closingFileReader{
		ReadCloser: r,
		onClose:    func() { cancel() },
	}, nil
}

func buildArchiveArgs(format, treeish string, pathspecs []string) []string {
	args := []string{"archive", "--worktree-attributes", "--format=" + format}

	if format == string(gitserver.ArchiveFormatZip) {
		args = append(args, "-0")
	}

	args = append(args, treeish, "--")
	args = append(args, pathspecs...)

	return args
}
