package gitcli

import (
	"bufio"
	"context"
	"io"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) RawDiff(ctx context.Context, base string, head string, typ git.GitDiffComparisonType, paths ...string) (io.ReadCloser, error) {
	baseOID, err := g.revParse(ctx, base)
	if err != nil {
		return nil, err
	}
	headOID, err := g.revParse(ctx, head)
	if err != nil {
		return nil, err
	}

	return g.NewCommand(ctx, WithArguments(buildRawDiffArgs(baseOID, headOID, typ, paths)...))
}

func buildRawDiffArgs(base, head api.CommitID, typ git.GitDiffComparisonType, paths []string) []string {
	var rangeType string
	switch typ {
	case git.GitDiffComparisonTypeIntersection:
		rangeType = "..."
	case git.GitDiffComparisonTypeOnlyInHead:
		rangeType = ".."
	}
	rangeSpec := string(base) + rangeType + string(head)

	return append([]string{
		"diff",
		"--find-renames",
		"--full-index",
		"--inter-hunk-context=3",
		"--no-prefix",
		rangeSpec,
		"--",
	}, paths...)
}

func (g *gitCLIBackend) ChangedFiles(ctx context.Context, base, head string) (git.ChangedFilesIterator, error) {
	args := []string{
		"diff-tree",
		"-r",
		"--root",
		"--format=format:",
		"--no-prefix",
		"--name-status",
		"--no-renames",
		"-z",
	}

	if base != "" {
		baseOID, err := g.revParse(ctx, base)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve base commit %q", base)
		}

		args = append(args, string(baseOID))
	}

	headOID, err := g.revParse(ctx, head)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve head commit %q", head)
	}

	args = append(args, string(headOID))

	rc, err := g.NewCommand(ctx, WithArguments(args...))
	if err != nil {
		return nil, errors.Wrap(err, "failed to run git diff-tree command")
	}

	return newGitDiffIterator(rc), nil
}

func newGitDiffIterator(rc io.ReadCloser) git.ChangedFilesIterator {
	scanner := bufio.NewScanner(rc)
	scanner.Split(byteutils.ScanNullLines)

	closer := sync.OnceValue(func() error {
		return rc.Close()
	})

	return &gitDiffIterator{
		rc:             rc,
		scanner:        scanner,
		onceFuncCloser: closer,
	}
}

type gitDiffIterator struct {
	rc      io.ReadCloser
	scanner *bufio.Scanner

	onceFuncCloser func() error
}

func (i *gitDiffIterator) Next() (gitdomain.PathStatus, error) {
	for i.scanner.Scan() {
		status := i.scanner.Text()
		if len(status) == 0 {
			continue
		}

		if !i.scanner.Scan() {
			return gitdomain.PathStatus{}, errors.New("uneven pairs")
		}
		path := i.scanner.Text()

		switch status[0] {
		case 'A':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.AddedAMD}, nil
		case 'M':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.ModifiedAMD}, nil
		case 'D':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.DeletedAMD}, nil
		default:
			return gitdomain.PathStatus{}, errors.Errorf("encountered unknown file status %q for file %q", status, path)
		}
	}

	if err := i.scanner.Err(); err != nil {
		return gitdomain.PathStatus{}, errors.Wrap(err, "failed to scan git diff output")
	}

	return gitdomain.PathStatus{}, io.EOF
}

func (i *gitDiffIterator) Close() error {
	return i.onceFuncCloser()
}

var _ git.ChangedFilesIterator = &gitDiffIterator{}
