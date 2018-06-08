package git

import (
	"context"
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// FetchTar returns a reader for running "git archive --format=tar
// commit".
func FetchTar(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (rc io.ReadCloser, err error) {
	// Archive returns a zip file read into
	// memory. However, we do not need to read into memory and we want a
	// tar, so we directly run the gitserver Command.
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenTar")
	ext.Component.Set(span, "git")
	span.SetTag("Repo", repo.Name)
	span.SetTag("Commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
		}
		span.Finish()
	}()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git", "archive", "--format=tar", string(commit))
	cmd.Repo = repo
	rc, err = gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		if errcode.IsNotFound(err) {
			err = badRequestError{err.Error()}
		}
		return nil, err
	}
	return rc, nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
