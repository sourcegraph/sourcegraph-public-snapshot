package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		ten := 10
		t.Run("returns all", func(t *testing.T) {
			tuples, err := kvps.List(ctx, RepoKVPListOptions{}, PaginationArgs{First: &ten})
			require.NoError(t, err)
			require.Equal(t, []KeyValuePair{
				{Key: "key1", Value: strPtr("value1")},
				{Key: "tag1", Value: nil},
			}, tuples)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			tuples, err := kvps.List(ctx, RepoKVPListOptions{QueryKey: strPtr("tag")}, PaginationArgs{First: &ten})
			require.NoError(t, err)
			require.Equal(t, []KeyValuePair{
				{Key: "tag1", Value: nil},
			}, tuples)

			tuples, err = kvps.List(ctx, RepoKVPListOptions{QueryKey: strPtr("1"), QueryValue: strPtr("val")}, PaginationArgs{First: &ten})
			require.NoError(t, err)
			require.Equal(t, []KeyValuePair{
				{Key: "key1", Value: strPtr("value1")},
			}, tuples)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			tuples, err := kvps.List(ctx, RepoKVPListOptions{QueryKey: strPtr("nonexisting")}, PaginationArgs{First: &ten})
			require.NoError(t, err)
			require.Empty(t, tuples)
		})
	})

	t.Run("Count", func(t *testing.T) {
		ten := 10
		t.Run("returns all", func(t *testing.T) {
			count, err := kvps.Count(ctx, RepoKVPListOptions{})
			require.NoError(t, err)
			require.Equal(t, count, 2)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			count, err := kvps.Count(ctx, RepoKVPListOptions{QueryKey: strPtr("ey")})
			require.NoError(t, err)
			require.Equal(t, 1, count)

			count, err = kvps.Count(ctx, RepoKVPListOptions{QueryKey: strPtr("1"), QueryValue: strPtr("val")})
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			count, err := kvps.List(ctx, RepoKVPListOptions{QueryKey: strPtr("nonexisting")}, PaginationArgs{First: &ten})
			require.NoError(t, err)
			require.Empty(t, count)
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
