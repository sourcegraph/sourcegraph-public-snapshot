package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func testStoreRepoStatus(t *testing.T, ctx context.Context, s *Store, _ bt.Clock) {
	repo, _ := bt.CreateTestRepo(t, ctx, s.DatabaseDB())

	rs := &btypes.RepoStatus{
		RepoID:  repo.ID,
		Commit:  "abcdef",
		Ignored: true,
	}

	t.Run("get missing", func(t *testing.T) {
		have, err := s.GetRepoStatus(ctx, repo.ID, "abcdef")
		assert.Nil(t, have)
		assert.Equal(t, ErrNoResults, err)
	})

	t.Run("insert", func(t *testing.T) {
		err := s.UpsertRepoStatus(ctx, rs)
		assert.NoError(t, err)
	})

	t.Run("get after insert", func(t *testing.T) {
		have, err := s.GetRepoStatus(ctx, repo.ID, "abcdef")
		assert.NoError(t, err)
		assert.Equal(t, rs, have)
	})

	t.Run("update", func(t *testing.T) {
		rs.Commit = "ghijkl"
		err := s.UpsertRepoStatus(ctx, rs)
		assert.NoError(t, err)
	})

	t.Run("get after update", func(t *testing.T) {
		have, err := s.GetRepoStatus(ctx, repo.ID, "abcdef")
		assert.Nil(t, have)
		assert.Equal(t, ErrNoResults, err)

		have, err = s.GetRepoStatus(ctx, repo.ID, "ghijkl")
		assert.NoError(t, err)
		assert.Equal(t, rs, have)
	})
}
