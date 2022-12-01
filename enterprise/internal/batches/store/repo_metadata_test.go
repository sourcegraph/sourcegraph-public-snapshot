package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	irepos "github.com/sourcegraph/sourcegraph/internal/repos"
)

func testRepoMetadata(t *testing.T, ctx context.Context, s *Store, _ bt.Clock) {
	repos, _ := bt.CreateTestRepos(t, ctx, s.DatabaseDB(), 5)
	repoIDs := make([]api.RepoID, len(repos))
	for i := range repos {
		repoIDs[i] = repos[i].ID
	}

	// Set up a user who only has access to repos[0].
	user := bt.CreateTestUser(t, s.DatabaseDB(), false)
	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))
	bt.MockRepoPermissions(t, s.DatabaseDB(), user.ID, repos[0].ID)

	// Also set up a context with an internal actor for more functional testing.
	internalCtx := actor.WithInternalActor(ctx)

	forEachContext := func(t *testing.T, tc func(*testing.T, context.Context)) {
		t.Helper()
		for name, ctx := range map[string]context.Context{
			"internal": internalCtx,
			"user":     userCtx,
		} {
			t.Run(name, func(t *testing.T) { tc(t, ctx) })
		}
	}

	t.Run("list before creation", func(t *testing.T) {
		ids, cursor, err := s.ListRepoIDsMissingMetadata(internalCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Equal(t, repoIDs, ids)

		ids, cursor, err = s.ListRepoIDsMissingMetadata(userCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Equal(t, []api.RepoID{repos[0].ID}, ids)

		forEachContext(t, func(t *testing.T, ctx context.Context) {
			metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
			assert.NoError(t, err)
			assert.Zero(t, cursor)
			assert.Empty(t, metas)
		})
	})

	t.Run("upsert", func(t *testing.T) {
		for _, repo := range repos {
			meta := btypes.RepoMetadata{RepoID: repo.ID, Ignored: true}
			err := s.UpsertRepoMetadata(internalCtx, &meta)
			assert.NoError(t, err)
			assert.NotZero(t, meta.CreatedAt)
			assert.NotZero(t, meta.UpdatedAt)
		}
	})

	t.Run("list after creation", func(t *testing.T) {
		forEachContext(t, func(t *testing.T, ctx context.Context) {
			ids, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
			assert.NoError(t, err)
			assert.Zero(t, cursor)
			assert.Empty(t, ids)

			metas, cursor, err := s.ListReposWithOutdatedMetadata(ctx, CursorOpts{})
			assert.NoError(t, err)
			assert.Zero(t, cursor)
			assert.Empty(t, metas)
		})
	})

	t.Run("list after updating repos", func(t *testing.T) {
		rs := irepos.NewStore(s.logger, s.DatabaseDB())
		for i, repo := range repos {
			if i%2 == 0 {
				repo.UpdatedAt = repo.UpdatedAt.Add(1 * time.Second)
			} else {
				repo.CreatedAt = repo.CreatedAt.Add(1 * time.Second)
			}
			rs.UpdateRepo(internalCtx, repo)
		}

		forEachContext(t, func(t *testing.T, ctx context.Context) {
			missing, cursor, err := s.ListRepoIDsMissingMetadata(ctx, CursorOpts{})
			assert.NoError(t, err)
			assert.Zero(t, cursor)
			assert.Empty(t, missing)
		})

		metas, cursor, err := s.ListReposWithOutdatedMetadata(internalCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)

		ids := make([]api.RepoID, len(metas))
		for i := range metas {
			assert.True(t, metas[i].Ignored)
			ids[i] = metas[i].RepoID
		}
		assert.Equal(t, repoIDs, ids)

		metas, cursor, err = s.ListReposWithOutdatedMetadata(userCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Len(t, metas, 1)
		assert.Equal(t, repos[0].ID, metas[0].RepoID)
	})

	t.Run("upsert", func(t *testing.T) {
		metas, cursor, err := s.ListReposWithOutdatedMetadata(internalCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)

		for _, meta := range metas {
			meta.Ignored = false
			err := s.UpsertRepoMetadata(internalCtx, meta)
			assert.NoError(t, err)
		}
	})

	t.Run("list after upserting repo metadata", func(t *testing.T) {
		missing, cursor, err := s.ListRepoIDsMissingMetadata(internalCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.Empty(t, missing)

		metas, cursor, err := s.ListReposWithOutdatedMetadata(internalCtx, CursorOpts{})
		assert.NoError(t, err)
		assert.Zero(t, cursor)
		assert.NotEmpty(t, metas)
	})

	t.Run("get", func(t *testing.T) {
		assertValidMetadata := func(t *testing.T, want api.RepoID, meta *btypes.RepoMetadata) {
			t.Helper()
			assert.NotNil(t, meta)
			assert.Equal(t, want, meta.RepoID)
			assert.NotZero(t, meta.CreatedAt)
			assert.NotZero(t, meta.UpdatedAt)
			assert.False(t, meta.Ignored)
		}

		t.Run("internal", func(t *testing.T) {
			for _, repo := range repos {
				meta, err := s.GetRepoMetadata(internalCtx, repo.ID)
				assert.NoError(t, err)
				assertValidMetadata(t, repo.ID, meta)
			}

			meta, err := s.GetRepoMetadata(internalCtx, 0)
			assert.Nil(t, meta)
			assert.Equal(t, ErrNoResults, err)
		})

		t.Run("user", func(t *testing.T) {
			meta, err := s.GetRepoMetadata(userCtx, repos[0].ID)
			assert.NoError(t, err)
			assertValidMetadata(t, repos[0].ID, meta)

			for _, repo := range repos[1:] {
				meta, err := s.GetRepoMetadata(userCtx, repo.ID)
				assert.Nil(t, meta)
				assert.Equal(t, ErrNoResults, err)
			}
		})
	})
}
