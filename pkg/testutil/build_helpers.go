package testutil

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// These build helpers will need to be moved to another package so
// they can be called by other package's build tests when those
// exist. Until so, they live here.

// BuildRepoAndWait triggers a build of the repo commit and waits
// until it is done (either because it succeeded or failed). If the
// build fails, or if there is an error creating or waiting for the
// build, a non-nil error is returned.
//
// t is only used for logging; t.Error*/t.Fatal* are never
// called. (This is so that the line numbers in errors refer to the
// actual caller, not to this helper func.) This means you probably
// should check the error value returned by this func in your test.
func BuildRepoAndWait(t *testing.T, ctx context.Context, repo string, commitID string) (*sourcegraph.Build, *sourcegraph.BuildSpec, error) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	// Create the build.
	b, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{
		Repo:     repo,
		CommitID: commitID,
		Config:   sourcegraph.BuildConfig{Queue: true},
	})
	if err != nil {
		return nil, nil, err
	}
	t.Logf("created build %s; waiting for it to complete", b.Spec().IDString())

	buildSpec := b.Spec()
	b, err = WaitForBuild(t, ctx, buildSpec)
	return b, &buildSpec, err
}

func WaitForBuild(t *testing.T, ctx context.Context, buildSpec sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	cl, _ := sourcegraph.NewClientFromContext(ctx)

	// Wait for the build to complete.
	start, waitStart, waitEnd := time.Now(), 5*time.Second*ciFactor, 20*time.Second*ciFactor
	for {
		elapsed := time.Since(start)
		if elapsed > waitEnd {
			return nil, fmt.Errorf("build did not complete within %s", waitEnd)
		}

		b, err := cl.Builds.Get(ctx, &buildSpec)
		if grpc.Code(err) == codes.NotFound {
			// Maybe the build hasn't actually been created yet; don't
			// treat that as a fatal error, just wait a bit.
			time.Sleep(250 * time.Millisecond)
			continue
		} else if err != nil {
			return nil, err
		}

		if b.StartedAt == nil && elapsed > waitStart {
			return nil, fmt.Errorf("build did not start within %s", waitStart)
		}

		if b.EndedAt != nil {
			return b, nil
		}

		time.Sleep(250 * time.Millisecond)
	}
}
