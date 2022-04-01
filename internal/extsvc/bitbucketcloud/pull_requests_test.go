package bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func TestClient_CreatePullRequest_SameOrigin(t *testing.T) {
	// WHEN UPDATING: this test requires a new branch in
	// https://bitbucket.org/sourcegraph-testing/src-cli/src/master/ to open a
	// pull request. The simplest way to accomplish this is to do the following,
	// replacing XX with the next number after the branch currently in this
	// test, assuming you have an account set up with an SSH key that can push
	// to sourcegraph-testing/src-cli:
	//
	// $ cd /tmp
	// $ git clone git@bitbucket.org:sourcegraph-testing/src-cli.git
	// $ cd src-cli
	// $ git checkout -b branch-XX
	// $ git commit --allow-empty -m "new branch"
	// $ git push origin branch-XX
	//
	// Then update this test with the new branch number, and run the test suite
	// with the appropriate -update flag.

	branch := "branch-00"
	ctx := context.Background()

	c, save := newTestClient(t)
	defer save()

	repo := &Repo{
		FullName: "sourcegraph-testing/src-cli",
	}
	commonOpts := CreatePullRequestOpts{
		Title:        "Sourcegraph test " + branch,
		Description:  "This is a PR created by the Sourcegraph test suite.",
		SourceBranch: branch,
	}

	// We'll test the two cases with an explicit destination branch: that it's
	// valid, and that it's invalid. We'll test the omitted destination branch
	// case in the fork test.

	t.Run("invalid destination branch", func(t *testing.T) {
		opts := commonOpts
		dest := "this-branch-should-never-exist"
		opts.DestinationBranch = &dest

		pr, err := c.CreatePullRequest(ctx, repo, opts)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
	})

	var id int64
	t.Run("valid destination branch", func(t *testing.T) {
		opts := commonOpts
		dest := "master"
		opts.DestinationBranch = &dest

		pr, err := c.CreatePullRequest(ctx, repo, opts)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
		assertGolden(t, pr)
		id = pr.ID
	})

	t.Run("recreated", func(t *testing.T) {
		// Bitbucket has the interesting behaviour that creating the same PR
		// multiple times succeeds, but without actually changing the PR. Let's
		// ensure that's still the case.
		opts := commonOpts
		dest := "master"
		opts.DestinationBranch = &dest

		pr, err := c.CreatePullRequest(ctx, repo, opts)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
		assertGolden(t, pr)

		// As an extra sanity check, let's check the ID against the previous
		// creation.
		assert.Equal(t, id, pr.ID)
	})
}

func TestClient_GetPullRequest(t *testing.T) {
	// WHEN UPDATING: this test expects
	// https://bitbucket.org/sourcegraph-testing/src-cli/pull-requests/1/always-open-pr
	// to be open.

	ctx := context.Background()

	c, save := newTestClient(t)
	defer save()

	repo := &Repo{
		FullName: "sourcegraph-testing/src-cli",
	}

	t.Run("not found", func(t *testing.T) {
		pr, err := c.GetPullRequest(ctx, repo, 0)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
		assert.True(t, errcode.IsNotFound(err))
	})

	t.Run("found", func(t *testing.T) {
		pr, err := c.GetPullRequest(ctx, repo, 1)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
		assertGolden(t, pr)
	})
}
