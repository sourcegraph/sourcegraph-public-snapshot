package gitcli

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
)

func (g *gitCLIBackend) ArchiveReader(ctx context.Context, format git.ArchiveFormat, treeish string, pathspecs []string) (io.ReadCloser, error) {
	archiveArgs := buildArchiveArgs(format, treeish, pathspecs)
	cmd, cancel, err := g.gitCommand(ctx, archiveArgs...)
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

func buildArchiveArgs(format git.ArchiveFormat, treeish string, pathspecs []string) []string {
	args := []string{"archive", "--worktree-attributes", "--format=" + string(format)}

	if format == git.ArchiveFormatZip {
		args = append(args, "-0")
	}

	args = append(args, treeish, "--")
	args = append(args, pathspecs...)

	return args
}
