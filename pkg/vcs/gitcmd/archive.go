package gitcmd

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// Archive implements vcs.Archiver.
func (r *Repository) Archive(ctx context.Context, commitID api.CommitID) (zipData []byte, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Archive")
	span.SetTag("URL", r.repoURI)
	span.SetTag("Commit", commitID)
	defer func() {
		if err == nil {
			span.SetTag("byteSize", len(zipData))
		} else {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	if err := checkSpecArgSafety(string(commitID)); err != nil {
		return nil, err
	}

	// Compression level of 0 (no compression) seems to perform the
	// best overall on fast network links, but this has not been tuned
	// thoroughly.
	cmd := r.command("git", "archive", "--format=zip", "-0", string(commitID))
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, stderr)
	}
	return stdout, nil
}
