package git

import (
	"bytes"
	"context"
	"fmt"
	"os"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

// ReadFile returns the content of the named file at commit.
func (r *Repository) ReadFile(ctx context.Context, commit api.CommitID, name string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadFile")
	span.SetTag("Name", name)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	name = util.Rel(name)
	b, err := r.readFileBytes(ctx, commit, name)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Repository) readFileBytes(ctx context.Context, commit api.CommitID, name string) ([]byte, error) {
	r.ensureAbsCommit(commit)

	cmd := r.command("git", "show", string(commit)+":"+name)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) || bytes.Contains(out, []byte("does not exist")) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		if bytes.HasPrefix(out, []byte("fatal: bad object ")) {
			// Could be a git submodule.
			fi, err := r.Stat(ctx, commit, name)
			if err != nil {
				return nil, err
			}
			// Return empty for a submodule for now.
			if fi.Mode()&ModeSubmodule != 0 {
				return nil, nil
			}

		}
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	return out, nil
}
