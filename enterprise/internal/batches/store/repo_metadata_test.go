package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	irepos "github.com/sourcegraph/sourcegraph/internal/repos"
)

func testRepoMetadata(t *testing.T, ctx context.Context, s *Store, _ bt.Clock) {
	repos, _ := bt.CreateTestRepos(t, ctx, s.DatabaseDB(), 5)
	repoIDs := make([]api.RepoID, len(repos))
	for i := range repos {
		repoIDs[i] = repos[i].ID
	}

	t.Run("list before creation", func(t *testing.T) {
		ids, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Equal(t, repoIDs, ids)

		metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, metas)
	})

	t.Run("create", func(t *testing.T) {
		for _, repo := range repos {
			meta := btypes.RepoMetadata{RepoID: repo.ID, Ignored: true}
			err := s.UpsertRepoMetadata(ctx, &meta)
			assert.NoError(t, err)
			assert.NotZero(t, meta.CreatedAt)
			assert.NotZero(t, meta.UpdatedAt)
		}
	})

	t.Run("list after creation", func(t *testing.T) {
		ids, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, ids)

		metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, metas)
	})

	t.Run("list after updating repos", func(t *testing.T) {
		rs := irepos.NewStore(s.logger, s.DatabaseDB())
		for _, repo := range repos {
			repo.UpdatedAt = repo.UpdatedAt.Add(1 * time.Second)
			rs.UpdateRepo(ctx, repo)
		}

		missing, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, missing)

		metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)

		ids := make([]api.RepoID, len(metas))
		for i := range metas {
			assert.True(t, metas[i].Ignored)
			ids[i] = metas[i].RepoID
		}
		assert.Equal(t, repoIDs, ids)
	})

	t.Run("upsert", func(t *testing.T) {
		metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)

		for _, meta := range metas {
			meta.Ignored = false
			err := s.UpsertRepoMetadata(ctx, meta)
			assert.NoError(t, err)
		}
	})

	t.Run("list after upserting repo metadata", func(t *testing.T) {
		missing, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, missing)

		metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)
	})

	t.Run("get", func(t *testing.T) {
		for _, repo := range repos {
			meta, err := s.GetRepoMetadata(ctx, repo.ID)
			assert.NoError(t, err)
			assert.NotNil(t, meta)
			assert.Equal(t, repo.ID, meta.RepoID)
			assert.NotZero(t, meta.CreatedAt)
			assert.NotZero(t, meta.UpdatedAt)
			assert.False(t, meta.Ignored)
		}

		meta, err := s.GetRepoMetadata(ctx, 0)
		assert.Nil(t, meta)
		assert.Equal(t, ErrNoResults, err)
	})
}
