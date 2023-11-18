package usagestats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAggregatedRepoMetadataSummary(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	err := db.Repos().Create(ctx, &types.Repo{
		Name: "repo",
	})
	require.NoError(t, err)

	repo, err := db.Repos().GetByName(ctx, "repo")
	require.NoError(t, err)

	kvps := db.RepoKVPs()

	t.Run("no data", func(t *testing.T) {
		summary, err := getAggregatedRepoMetadataSummary(ctx, db)

		require.NoError(t, err)
		require.Equal(t, &types.RepoMetadataAggregatedSummary{
			RepoMetadataCount:      int32Ptr(0),
			ReposWithMetadataCount: int32Ptr(0),
			IsEnabled:              true,
		}, summary)
	})

	t.Run("with data", func(t *testing.T) {
		err = kvps.Create(ctx, repo.ID, database.KeyValuePair{
			Key:   "tag1",
			Value: nil,
		})
		require.NoError(t, err)

		value1 := "value1"
		err = kvps.Create(ctx, repo.ID, database.KeyValuePair{
			Key:   "key1",
			Value: &value1,
		})
		require.NoError(t, err)

		summary, err := getAggregatedRepoMetadataSummary(ctx, db)

		require.NoError(t, err)
		require.Equal(t, &types.RepoMetadataAggregatedSummary{
			RepoMetadataCount:      int32Ptr(2),
			ReposWithMetadataCount: int32Ptr(1),
			IsEnabled:              true,
		}, summary)
	})

	t.Run("feature flag is set to true", func(t *testing.T) {
		db.FeatureFlags().CreateBool(ctx, "repository-metadata", true)

		summary, err := getAggregatedRepoMetadataSummary(ctx, db)

		require.NoError(t, err)
		require.Equal(t, &types.RepoMetadataAggregatedSummary{
			RepoMetadataCount:      int32Ptr(2),
			ReposWithMetadataCount: int32Ptr(1),
			IsEnabled:              true,
		}, summary)
	})

	t.Run("feature flag is set to false", func(t *testing.T) {
		db.FeatureFlags().DeleteFeatureFlag(ctx, "repository-metadata")
		db.FeatureFlags().CreateBool(ctx, "repository-metadata", false)

		summary, err := getAggregatedRepoMetadataSummary(ctx, db)

		require.NoError(t, err)
		require.Equal(t, &types.RepoMetadataAggregatedSummary{
			RepoMetadataCount:      int32Ptr(2),
			ReposWithMetadataCount: int32Ptr(1),
			IsEnabled:              false,
		}, summary)
	})
}
