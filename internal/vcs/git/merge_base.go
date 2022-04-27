package git

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// MergeBase returns the merge base commit for the specified commits.
func MergeBase(ctx context.Context, db database.DB, repo api.RepoName, a, b api.CommitID) (api.CommitID, error) {
	if Mocks.MergeBase != nil {
		return Mocks.MergeBase(repo, a, b)
	}
	span, ctx := ot.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	cmd := gitserver.NewClient(db).GitCommand(repo, "merge-base", "--", string(a), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}
