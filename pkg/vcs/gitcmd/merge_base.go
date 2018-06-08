package gitcmd

import (
	"bytes"
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// MergeBase returns the merge base commit for the specified commits.
func (r *Repository) MergeBase(ctx context.Context, a, b api.CommitID) (api.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	cmd := r.command("git", "merge-base", "--", string(a), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}
