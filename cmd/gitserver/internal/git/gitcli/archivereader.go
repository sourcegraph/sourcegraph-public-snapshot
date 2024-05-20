package gitcli

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
)

func (g *gitCLIBackend) ArchiveReader(ctx context.Context, format git.ArchiveFormat, treeish string, paths []string) (io.ReadCloser, error) {
	if err := checkSpecArgSafety(treeish); err != nil {
		return nil, err
	}

	// Verify the tree-ish exists, if it doesn't this will return a RevisionNotFoundError:
	_, err := g.getObjectType(ctx, treeish)
	if err != nil {
		return nil, err
	}

	archiveArgs := buildArchiveArgs(format, treeish, paths)

	return g.NewCommand(ctx, WithArguments(archiveArgs...))
}

func buildArchiveArgs(format git.ArchiveFormat, treeish string, paths []string) []string {
	args := []string{"archive", "--worktree-attributes", "--format=" + string(format)}

	if format == git.ArchiveFormatZip {
		args = append(args, "-0")
	}

	args = append(args, treeish, "--")
	for _, p := range paths {
		args = append(args, pathspecLiteral(p))
	}

	return args
}

// pathspecLiteral constructs a pathspec that matches a path without interpreting "*" or "?" as special
// characters.
//
// See: https://git-scm.com/docs/gitglossary#Documentation/gitglossary.txt-literal
func pathspecLiteral(s string) string { return ":(literal)" + s }
