package git

import (
	"bytes"
	"context"
	"fmt"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

// ReadFile returns the content of the named file at commit.
func ReadFile(ctx context.Context, repo gitserver.Repo, commit api.CommitID, name string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadFile")
	span.SetTag("Name", name)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	name = util.Rel(name)
	b, err := readFileBytes(ctx, repo, commit, name)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readFileBytes(ctx context.Context, repo gitserver.Repo, commit api.CommitID, name string) ([]byte, error) {
	ensureAbsCommit(commit)

	cmd := gitserver.DefaultClient.Command("git", "show", string(commit)+":"+name)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) || bytes.Contains(out, []byte("does not exist")) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		if bytes.HasPrefix(out, []byte("fatal: bad object ")) {
			// Could be a git submodule.
			fi, err := Stat(ctx, repo, commit, name)
			if err != nil {
				return nil, err
			}
			// Return empty for a submodule for now.
			if fi.Mode()&ModeSubmodule != 0 {
				return nil, nil
			}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}
	return out, nil
}
