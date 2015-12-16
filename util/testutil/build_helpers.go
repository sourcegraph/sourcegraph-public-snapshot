package testutil

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/grpccache"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: repo}, Rev: commitID, CommitID: commitID}

	cl := sourcegraph.NewClientFromContext(ctx)

	// Create the build.
	b, err := cl.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &sourcegraph.BuildCreateOptions{
		BuildConfig: sourcegraph.BuildConfig{
			Import: true, Queue: true, UseCache: false, Priority: 1,
		},
		Force: true,
	}})

	if err != nil {
		return nil, nil, err
	}
	t.Logf("created build %s; waiting for it to complete", b.Spec().IDString())

	// Wait for the build to complete.
	buildSpec := b.Spec()
	start, waitStart, waitEnd := time.Now(), 2*time.Second*ciFactor, 10*time.Second*ciFactor
	for {
		elapsed := time.Since(start)
		if elapsed > waitEnd {
			return nil, nil, fmt.Errorf("build did not complete within %s", waitEnd)
		}

		b, err := cl.Builds.Get(grpccache.NoCache(ctx), &buildSpec)
		if err != nil {
			return nil, nil, err
		}

		if b.StartedAt == nil && elapsed > waitStart {
			return nil, nil, fmt.Errorf("build did not start within %s", waitStart)
		}

		if b.EndedAt != nil {
			return b, &buildSpec, nil
		}

		time.Sleep(250 * time.Millisecond)
	}
}
