package database

import (
	"context"
	"sort"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestRepoKVPs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()
	kvps := db.RepoKVPs()

	err := db.Repos().Create(ctx, &types.Repo{
		Name: "repo",
	})
	require.NoError(t, err)

	repo, err := db.Repos().GetByName(ctx, "repo")
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		t.Run("non-nil value", func(t *testing.T) {
			err := kvps.Create(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: pointers.Ptr("value1"),
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
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: pointers.Ptr("value1")})
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

	t.Run("ListKeys", func(t *testing.T) {
		t.Run("returns all", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{}, PaginationArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			sort.Strings(keys)
			require.Equal(t, []string{"key1", "tag1"}, keys)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("tag")}, PaginationArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			require.Equal(t, []string{"tag1"}, keys)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			keys, err := kvps.ListKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("nonexisting")}, PaginationArgs{
				First: pointers.Ptr(10), OrderBy: OrderBy{{Field: string(RepoKVPListKeyColumn)}},
			})
			require.NoError(t, err)
			require.Empty(t, keys)
		})
	})

	t.Run("CountKeys", func(t *testing.T) {
		t.Run("returns all", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{})
			require.NoError(t, err)
			require.Equal(t, count, 2)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("ey")})
			require.NoError(t, err)
			require.Equal(t, 1, count)

			count, err = kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("1")})
			require.NoError(t, err)
			require.Equal(t, 2, count)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			count, err := kvps.CountKeys(ctx, RepoKVPListKeysOptions{Query: pointers.Ptr("nonexisting")})
			require.NoError(t, err)
			require.Empty(t, count)
		})
	})

	t.Run("ListValues", func(t *testing.T) {
		t.Run("returns all", func(t *testing.T) {
			values, err := kvps.ListValues(ctx, RepoKVPListValuesOptions{Key: "key1"}, PaginationArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListValueColumn)}},
			})
			require.NoError(t, err)
			require.Equal(t, []string{"value1"}, values)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			keys, err := kvps.ListValues(ctx, RepoKVPListValuesOptions{Key: "key1", Query: pointers.Ptr("val")}, PaginationArgs{
				First:   pointers.Ptr(10),
				OrderBy: OrderBy{{Field: string(RepoKVPListValueColumn)}},
			})
			require.NoError(t, err)
			require.Equal(t, []string{"value1"}, keys)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			keys, err := kvps.ListValues(ctx, RepoKVPListValuesOptions{Key: "key1", Query: pointers.Ptr("nonexisting")}, PaginationArgs{
				First: pointers.Ptr(10), OrderBy: OrderBy{{Field: string(RepoKVPListValueColumn)}},
			})
			require.NoError(t, err)
			require.Empty(t, keys)
		})
	})

	t.Run("CountValues", func(t *testing.T) {
		t.Run("returns all", func(t *testing.T) {
			count, err := kvps.CountValues(ctx, RepoKVPListValuesOptions{Key: "key1"})
			require.NoError(t, err)
			require.Equal(t, count, 1)
		})

		t.Run("returns when found match by query", func(t *testing.T) {
			count, err := kvps.CountValues(ctx, RepoKVPListValuesOptions{Key: "key1", Query: pointers.Ptr("value")})
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("returns empty when found no match by query", func(t *testing.T) {
			count, err := kvps.CountValues(ctx, RepoKVPListValuesOptions{Key: "key1", Query: pointers.Ptr("nonexisting")})
			require.NoError(t, err)
			require.Empty(t, count)
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			kvp, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "key1",
				Value: pointers.Ptr("value2"),
			})
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: pointers.Ptr("value2")})
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
				Value: pointers.Ptr("value3"),
			})
			require.NoError(t, err)
			require.Equal(t, kvp, KeyValuePair{Key: "key1", Value: pointers.Ptr("value3")})
		})

		t.Run("does not exist", func(t *testing.T) {
			_, err := kvps.Update(ctx, repo.ID, KeyValuePair{
				Key:   "noexist",
				Value: pointers.Ptr("value3"),
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
