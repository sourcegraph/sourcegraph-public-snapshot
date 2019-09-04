package git

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

// ReadFile returns the first maxBytes of the named file at commit. If maxBytes <= 0, the entire
// file is read. (If you just need to check a file's existence, use Stat, not ReadFile.)
func ReadFile(ctx context.Context, repo gitserver.Repo, commit api.CommitID, name string, maxBytes int64) ([]byte, error) {
	if Mocks.ReadFile != nil {
		return Mocks.ReadFile(commit, name)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadFile")
	span.SetTag("Name", name)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	name = util.Rel(name)
	b, err := readFileBytes(ctx, repo, commit, name, maxBytes)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readFileBytes(ctx context.Context, repo gitserver.Repo, commit api.CommitID, name string, maxBytes int64) ([]byte, error) {
	ensureAbsCommit(commit)

	cmd := gitserver.DefaultClient.Command("git", "show", string(commit)+":"+name)
	cmd.Repo = repo
	stdout, err := gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	r := io.Reader(stdout)
	if maxBytes > 0 {
		r = io.LimitReader(r, maxBytes)
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		if strings.Contains(err.Error(), "exists on disk, but not in") || strings.Contains(err.Error(), "does not exist") {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		if strings.Contains(err.Error(), "fatal: bad object ") {
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
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, err))
	}
	return data, nil
}
