package git

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// ArchiveOptions contains options for (Repository).Archive.
type ArchiveOptions struct {
	Treeish string   // the tree or commit to produce an archive for
	Format  string   // format of the resulting archive (usually "tar" or "zip")
	Paths   []string // if nonempty, only include these paths
}

func Archive(ctx context.Context, repo gitserver.Repo, commitID api.CommitID) (zipData []byte, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Archive")
	span.SetTag("Repo", repo.Name)
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
	cmd := gitserver.DefaultClient.Command("git", "archive", "--format=zip", "-0", string(commitID))
	cmd.Repo = repo
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, stderr)
	}
	return stdout, nil
}
