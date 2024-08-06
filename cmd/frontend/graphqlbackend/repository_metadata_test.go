package graphqlbackend

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/log/logtest"

	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

func TestRepositoryMetadata(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDBFrom(database.NewDB(logger, dbtest.NewDB(t)))

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{}, nil)
	db.UsersFunc.SetDefaultReturn(users)

	permissions := dbmocks.NewMockPermissionStore()
	permissions.GetPermissionForUserFunc.SetDefaultReturn(&types.Permission{
		ID:        1,
		Namespace: rtypes.RepoMetadataNamespace,
		Action:    rtypes.RepoMetadataWriteAction,
		CreatedAt: time.Now(),
	}, nil)
	db.PermissionsFunc.SetDefaultReturn(permissions)

	err := db.Repos().Create(ctx, &types.Repo{
		Name: "testrepo",
	})
	require.NoError(t, err)
	repo, err := db.Repos().GetByName(ctx, "testrepo")
	require.NoError(t, err)

	schema := newSchemaResolver(db, gitserver.NewTestClient(t), nil)
	gqlID := MarshalRepositoryID(repo.ID)

	t.Run("add", func(t *testing.T) {
		_, err = schema.AddRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val1"),
		})
		require.NoError(t, err)

		_, err = schema.AddRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: pointers.Ptr(" 	"),
		})
		require.Error(t, err)
		require.Equal(t, emptyNonNilValueError{value: " 	"}, err)

		_, err = schema.AddRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: nil,
		})
		require.NoError(t, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metadata(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equal(t, []KeyValuePair{{
			key:   "key1",
			value: pointers.Ptr("val1"),
		}, {
			key:   "tag1",
			value: nil,
		}}, kvps)
	})

	t.Run("update", func(t *testing.T) {
		_, err = schema.UpdateRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val2"),
		})
		require.NoError(t, err)

		_, err = schema.UpdateRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: pointers.Ptr("val3"),
		})
		require.NoError(t, err)

		_, err = schema.UpdateRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "tag1",
			Value: pointers.Ptr("     "),
		})
		require.Error(t, err)
		require.Equal(t, emptyNonNilValueError{value: "     "}, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metadata(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Equal(t, []KeyValuePair{{
			key:   "key1",
			value: pointers.Ptr("val2"),
		}, {
			key:   "tag1",
			value: pointers.Ptr("val3"),
		}}, kvps)
	})

	t.Run("delete", func(t *testing.T) {
		_, err = schema.DeleteRepoMetadata(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.NoError(t, err)

		_, err = schema.DeleteRepoMetadata(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "tag1",
		})
		require.NoError(t, err)

		repoResolver, err := schema.repositoryByID(ctx, gqlID)
		require.NoError(t, err)

		kvps, err := repoResolver.Metadata(ctx)
		require.NoError(t, err)
		sort.Slice(kvps, func(i, j int) bool {
			return kvps[i].key < kvps[j].key
		})
		require.Empty(t, kvps)
	})

	t.Run("handles feature flag", func(t *testing.T) {
		flags := map[string]bool{"repository-metadata": false}
		ctx = featureflag.WithFlags(ctx, featureflag.NewMemoryStore(flags, flags, flags))
		_, err = schema.AddRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val1"),
		})
		require.Error(t, err)
		require.Equal(t, featureDisabledError, err)

		_, err = schema.UpdateRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val2"),
		})
		require.Error(t, err)
		require.Equal(t, featureDisabledError, err)

		_, err = schema.DeleteRepoMetadata(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.Error(t, err)
		require.Equal(t, featureDisabledError, err)
	})

	t.Run("handles rbac", func(t *testing.T) {
		permissions.GetPermissionForUserFunc.SetDefaultReturn(nil, nil)

		// add
		_, err = schema.AddRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val1"),
		})
		require.Error(t, err)
		require.Equal(t, err, &rbac.ErrNotAuthorized{Permission: string(rbac.RepoMetadataWritePermission)})

		// update
		_, err = schema.UpdateRepoMetadata(ctx, struct {
			Repo  graphql.ID
			Key   string
			Value *string
		}{
			Repo:  gqlID,
			Key:   "key1",
			Value: pointers.Ptr("val2"),
		})
		require.Error(t, err)
		require.Equal(t, err, &rbac.ErrNotAuthorized{Permission: string(rbac.RepoMetadataWritePermission)})

		// delete
		_, err = schema.DeleteRepoMetadata(ctx, struct {
			Repo graphql.ID
			Key  string
		}{
			Repo: gqlID,
			Key:  "key1",
		})
		require.Error(t, err)
		require.Equal(t, err, &rbac.ErrNotAuthorized{Permission: string(rbac.RepoMetadataWritePermission)})
	})

}
