package database

import (
	"context"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func Test_LoadConfigurations(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	store := db.OwnSignalConfigurations()
	configurations, err := store.LoadConfigurations(ctx, LoadSignalConfigurationArgs{})
	require.NoError(t, err)

	autogold.Expect([]SignalConfiguration{
		{
			ID:          1,
			Name:        "recent-contributors",
			Description: "Indexes contributors in each file using repository history.",
		},
		{
			ID:          2,
			Name:        "recent-views",
			Description: "Indexes users that recently viewed files in Sourcegraph.",
		},
		{
			ID:          3,
			Name:        "analytics",
			Description: "Indexes ownership data to present in aggregated views like Admin > Analytics > Own and Repo > Ownership",
		},
	}).Equal(t, configurations)

	t.Run("load by name", func(t *testing.T) {
		configs, err := store.LoadConfigurations(ctx, LoadSignalConfigurationArgs{Name: "recent-contributors"})
		require.NoError(t, err)
		autogold.Expect([]SignalConfiguration{{
			ID:          1,
			Name:        "recent-contributors",
			Description: "Indexes contributors in each file using repository history.",
		}}).Equal(t, configs)
	})

	t.Run("not found", func(t *testing.T) {
		configs, err := store.LoadConfigurations(ctx, LoadSignalConfigurationArgs{Name: "not a real job"})
		require.NoError(t, err)
		require.Empty(t, configs)
	})
	t.Run("update signal config", func(t *testing.T) {
		require.NotEmpty(t, configurations)
		cfg := configurations[0]
		err := store.UpdateConfiguration(ctx, UpdateSignalConfigurationArgs{
			Name:                 cfg.Name,
			ExcludedRepoPatterns: []string{"github.com/findme/somewhere"},
			Enabled:              true,
		})
		require.NoError(t, err)

		configurations, err := store.LoadConfigurations(ctx, LoadSignalConfigurationArgs{})
		require.NoError(t, err)

		autogold.Expect([]SignalConfiguration{
			{
				ID:                   1,
				Name:                 "recent-contributors",
				Description:          "Indexes contributors in each file using repository history.",
				ExcludedRepoPatterns: []string{"github.com/findme/somewhere"},
				Enabled:              true,
			},
			{
				ID:          2,
				Name:        "recent-views",
				Description: "Indexes users that recently viewed files in Sourcegraph.",
			},
			{
				ID:          3,
				Name:        "analytics",
				Description: "Indexes ownership data to present in aggregated views like Admin > Analytics > Own and Repo > Ownership",
			},
		}).Equal(t, configurations)
	})
}
