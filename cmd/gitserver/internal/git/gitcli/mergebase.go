package gitcli

import (
	"bytes"
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) MergeBase(ctx context.Context, baseRevspec, headRevspec string) (api.CommitID, error) {
	out, err := g.NewCommand(
		ctx,
		WithArguments("merge-base", "--", baseRevspec, headRevspec),
	)
	if err != nil {
		return "", err
	}

	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// Exit code 1 and empty output most likely means that no common merge-base was found.
		var e *commandFailedError
		if errors.As(err, &e) {
			if e.ExitStatus == 1 {
				if len(e.Stderr) == 0 {
					return "", nil
				}
			} else if e.ExitStatus == 128 && bytes.Contains(e.Stderr, []byte("fatal: Not a valid object name")) {
				spec := headRevspec
				if bytes.Contains(e.Stderr, []byte(baseRevspec)) {
					spec = baseRevspec
				}
				return "", &gitdomain.RevisionNotFoundError{
					Repo: g.repoName,
					Spec: spec,
				}
			}
		}

		return "", err
	}

	return api.CommitID(bytes.TrimSpace(stdout)), nil
}

func (g *gitCLIBackend) MergeBaseOctopus(ctx context.Context, revspecs ...string) (api.CommitID, error) {
	if len(revspecs) < 2 {
		return "", errors.New("at least two revspecs must be given")
	}

	args := make([]string, 0, len(revspecs)+3)
	args = append(args, "merge-base", "--octopus", "--")
	args = append(args, revspecs...)

	out, err := g.NewCommand(
		ctx,
		WithArguments(args...),
	)
	if err != nil {
		return "", err
	}

	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// Exit code 1 and empty output most likely means that no common merge-base was found.
		var e *commandFailedError
		if errors.As(err, &e) {
			if e.ExitStatus == 1 {
				if len(e.Stderr) == 0 {
					return "", nil
				}
			} else if e.ExitStatus == 128 && bytes.Contains(e.Stderr, []byte("fatal: Not a valid object name")) {
				p := octopusNotAValidObjectRegexp.FindSubmatch(e.Stderr)
				var spec string
				if len(p) > 0 {
					spec = string(p[1])
				}
				return "", &gitdomain.RevisionNotFoundError{
					Repo: g.repoName,
					Spec: spec,
				}
			}
		}

		return "", err
	}

	return api.CommitID(bytes.TrimSpace(stdout)), nil
}

var octopusNotAValidObjectRegexp = lazyregexp.New(`fatal: Not a valid object name ([^\s]+)`)
