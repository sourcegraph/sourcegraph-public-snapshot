package gitserver

import (
	"context"
	"io"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func FetchTar(ctx context.Context, client *Client, repo Repo, commit api.CommitID) (r io.ReadCloser, err error) {
	// gitcmd.Repository.Archive returns a zip file read into
	// memory. However, we do not need to read into memory and we want a
	// tar, so we directly run the gitserver Command.
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenTar")
	ext.Component.Set(span, "git")
	span.SetTag("URL", repo)
	span.SetTag("Commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
		}
		span.Finish()
	}()

	if strings.HasPrefix(string(commit), "-") {
		return nil, badRequestError{("invalid git revision spec (begins with '-')")}
	}

	cmd := client.Command("git", "archive", "--format=tar", string(commit))
	cmd.Repo = repo
	cmd.EnsureRevision = string(commit)
	r, err = StdoutReader(ctx, cmd)
	if err != nil {
		if vcs.IsRepoNotExist(err) || err == vcs.ErrRevisionNotFound {
			err = badRequestError{err.Error()}
		}
		return nil, err
	}
	return r, nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
