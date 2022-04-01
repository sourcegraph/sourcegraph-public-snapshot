package bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func TestClient_CreatePullRequest_Fork(t *testing.T) {
	// WHEN UPDATING: this test requires a new branch in a fork of
	// https://bitbucket.org/sourcegraph-testing/src-cli/src/master/ to open a
	// pull request. The simplest way to accomplish this is to do the following,
	// replacing XX with the next number after the branch currently in this
	// test, and FORK with the user and repo src-cli was forked to:
	//
	// $ cd /tmp
	// $ git clone git@bitbucket.org:FORK.git
	// $ cd src-cli
	// $ git checkout -b branch-fork-XX
	// $ git commit --allow-empty -m "new branch"
	// $ git push origin branch-fork-XX
	//
	// Then update this test with the new branch number, and run the test suite
	// with the appropriate -update flag.

	branch := "branch-fork-00"
	fork := "aharvey-sg/src-cli-testing"
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
		SourceRepo: &Repo{
			FullName: fork,
		},
	}

	t.Run("invalid destination branch", func(t *testing.T) {
		opts := commonOpts
		dest := "this-branch-should-never-exist"
		opts.DestinationBranch = &dest

		pr, err := c.CreatePullRequest(ctx, repo, opts)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
	})

	var id int64
	t.Run("valid, omitted destination branch", func(t *testing.T) {
		opts := commonOpts

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

		pr, err := c.CreatePullRequest(ctx, repo, opts)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
		assertGolden(t, pr)

		// As an extra sanity check, let's check the ID against the previous
		// creation.
		assert.Equal(t, id, pr.ID)
	})
}

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

func TestClient_DeclinePullRequest(t *testing.T) {
	// WHEN UPDATING: this test expects a PR in
	// https://bitbucket.org/sourcegraph-testing/src-cli/ to be open. Note that
	// PRs cannot be reopened after being declined, so we can't use a stable ID
	// here — this must use a PR that is open and can be safely declined, such
	// as one created in the CreatePullRequest tests above. Update the ID below
	// with such a PR before updating!

	var id int64 = 2
	ctx := context.Background()

	c, save := newTestClient(t)
	defer save()

	repo := &Repo{
		FullName: "sourcegraph-testing/src-cli",
	}

	t.Run("not found", func(t *testing.T) {
		pr, err := c.DeclinePullRequest(ctx, repo, 0)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
		assert.True(t, errcode.IsNotFound(err))
	})

	t.Run("found", func(t *testing.T) {
		pr, err := c.DeclinePullRequest(ctx, repo, id)
		assert.Nil(t, err)
		assert.NotNil(t, pr)
		assertGolden(t, pr)
	})

	t.Run("already declined", func(t *testing.T) {
		// Given the above behaviour around CreatePullRequest being able to be
		// called multiple times with no apparent effect, one might expect that
		// you could do the same with declined pull requests. One cannot:
		// repeated invocations of DeclinePullRequest for the same ID will fail.
		pr, err := c.DeclinePullRequest(ctx, repo, id)
		assert.Nil(t, pr)
		assert.NotNil(t, err)
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
