package git

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// MergeBase returns the merge base commit for the specified commits.
func MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (api.CommitID, error) {
	if Mocks.MergeBase != nil {
		return Mocks.MergeBase(repo, a, b)
	}
	span, ctx := ot.StartSpanFromContext(ctx, "Git: MergeBase")
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
