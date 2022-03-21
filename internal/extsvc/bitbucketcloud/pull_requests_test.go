package bitbucketcloud

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

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
