package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestRepoKVPs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	kvps := db.RepoKVPs()

	err := db.Repos().Create(ctx, &types.Repo{
		Name: "repo",
	})
	require.NoError(t, err)

	repo, err := db.Repos().GetByName(ctx, "repo")
	require.NoError(t, err)

	strPtr := func(s string) *string { return &s }

	t.Run("Create", func(t *testing.T) {
		t.Run("non-nil value", func(t *testing.T) {
			err := kvps.Create(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: strPtr("value1"),
			})
			require.NoError(t, err)
		})

		t.Run("nil value", func(t *testing.T) {
			err = kvps.Create(ctx, repo.ID, KeyValuePair{
				Key:   "tag1",
				Value: nil,
			})
			require.NoError(t, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			kvp, err := kvps.Get(ctx, repo.ID, "key1")
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: strPtr("value1")})
		})

		t.Run("exists with nil value", func(t *testing.T) {
			kvp, err := kvps.Get(ctx, repo.ID, "tag1")
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "tag1", Value: nil})
		})

		t.Run("does not exist", func(t *testing.T) {
			_, err := kvps.Get(ctx, repo.ID, "noexist")
			require.Error(t, err)
		})

		t.Run("repo does not exist", func(t *testing.T) {
			_, err := kvps.Get(ctx, repo.ID+1, "key1")
			require.Error(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			kvps, err := kvps.List(ctx, repo.ID)
			require.NoError(t, err)
			require.Equal(t, kvps, []KeyValuePair{
				{Key: "key1", Value: strPtr("value1")},
				{Key: "tag1", Value: nil},
			})
		})

		t.Run("repo does not exist", func(t *testing.T) {
			kvps, err := kvps.List(ctx, repo.ID+1)
			require.NoError(t, err)
			require.Empty(t, kvps)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			kvp, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: strPtr("value2"),
			})
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: strPtr("value2")})
		})

		t.Run("into tag", func(t *testing.T) {
			kvp, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: nil,
			})
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: nil})
		})

		t.Run("from tag", func(t *testing.T) {
			kvp, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: strPtr("value3"),
			})
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: strPtr("value3")})
		})

		t.Run("does not exist", func(t *testing.T) {
			_, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "noexist",
				Value: strPtr("value3"),
			})
			require.Error(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			err := kvps.Delete(ctx, repo.ID, "key1")
			require.NoError(t, err)

			_, err = kvps.Get(ctx, repo.ID, "key1")
			require.Error(t, err)
		})

		t.Run("does not exist", func(t *testing.T) {
			err := kvps.Delete(ctx, repo.ID, "noexist")
			require.NoError(t, err)
		})
	})
}
