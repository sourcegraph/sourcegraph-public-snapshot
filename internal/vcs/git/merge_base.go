package git

import (
	"bytes"
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// MergeBase returns the merge base commit for the specified commits.
func MergeBase(ctx context.Context, repo gitserver.Repo, a, b api.CommitID) (api.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	cmd := gitserver.DefaultClient.Command("git", "merge-base", "--", string(a), string(b))
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}
