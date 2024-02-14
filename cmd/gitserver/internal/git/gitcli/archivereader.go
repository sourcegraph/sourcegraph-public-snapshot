package gitcli

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) ArchiveReader(ctx context.Context, format git.ArchiveFormat, treeish string, paths []string) (io.ReadCloser, error) {
	if err := g.verifyPaths(ctx, treeish, paths); err != nil {
		return nil, err
	}

	archiveArgs := buildArchiveArgs(format, treeish, paths)
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

func (g *gitCLIBackend) verifyPaths(ctx context.Context, treeish string, paths []string) error {
	args := []string{"ls-tree", treeish, "--"}
	args = append(args, paths...)
	cmd, cancel, err := g.gitCommand(ctx, args...)
	defer cancel()
	if err != nil {
		return err
	}

	out, err := g.runGitCommand(ctx, cmd)
	if err != nil {
		return err
	}
	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// If exit code is 128 and `not a tree object` is part of stderr, most likely we
		// are referencing a commit that does not exist.
		// We want to return a gitdomain.RevisionNotFoundError in that case.
		var e *CommandFailedError
		if errors.As(err, &e) && e.ExitStatus == 128 && (bytes.Contains(e.Stderr, []byte("not a tree object")) || bytes.Contains(e.Stderr, []byte("Not a valid object name"))) {
			return &gitdomain.RevisionNotFoundError{Repo: g.repoName, Spec: treeish}
		}

		return err
	}

	if len(paths) == 0 {
		return nil
	}

	// Check if the resulting objects match the requested
	// paths. If not, one or more of the requested
	// file paths don't exist.
	gotPaths := bytes.Split(bytes.TrimSpace(stdout), []byte("\n"))
	fileSet := collections.NewSet[string]()
	for _, p := range gotPaths {
		if len(p) == 0 {
			continue
		}
		pathSegments := bytes.Fields(p)
		fileSet.Add(string(pathSegments[len(pathSegments)-1]))
	}

	pathsSet := collections.NewSet[string]()
	for _, path := range paths {
		pathsSet.Add(path)
	}
	diff := pathsSet.Difference(fileSet)

	if len(diff) != 0 {
		return &os.PathError{Op: "open", Path: diff.Values()[0], Err: os.ErrNotExist}
	}

	return nil
}
