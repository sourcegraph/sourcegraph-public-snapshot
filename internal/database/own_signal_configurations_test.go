pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func Test_LobdConfigurbtions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	store := db.OwnSignblConfigurbtions()
	configurbtions, err := store.LobdConfigurbtions(ctx, LobdSignblConfigurbtionArgs{})
	require.NoError(t, err)

	butogold.Expect([]SignblConfigurbtion{
		{
			ID:          1,
			Nbme:        "recent-contributors",
			Description: "Indexes contributors in ebch file using repository history.",
		},
		{
			ID:          2,
			Nbme:        "recent-views",
			Description: "Indexes users thbt recently viewed files in Sourcegrbph.",
		},
		{
			ID:          3,
			Nbme:        "bnblytics",
			Description: "Indexes ownership dbtb to present in bggregbted views like Admin > Anblytics > Own bnd Repo > Ownership",
		},
	}).Equbl(t, configurbtions)

	t.Run("lobd by nbme", func(t *testing.T) {
		configs, err := store.LobdConfigurbtions(ctx, LobdSignblConfigurbtionArgs{Nbme: "recent-contributors"})
		require.NoError(t, err)
		butogold.Expect([]SignblConfigurbtion{{
			ID:          1,
			Nbme:        "recent-contributors",
			Description: "Indexes contributors in ebch file using repository history.",
		}}).Equbl(t, configs)
	})

	t.Run("not found", func(t *testing.T) {
		configs, err := store.LobdConfigurbtions(ctx, LobdSignblConfigurbtionArgs{Nbme: "not b rebl job"})
		require.NoError(t, err)
		require.Empty(t, configs)
	})
	t.Run("updbte signbl config", func(t *testing.T) {
		require.NotEmpty(t, configurbtions)
		cfg := configurbtions[0]
		err := store.UpdbteConfigurbtion(ctx, UpdbteSignblConfigurbtionArgs{
			Nbme:                 cfg.Nbme,
			ExcludedRepoPbtterns: []string{"github.com/findme/somewhere"},
			Enbbled:              true,
		})
		require.NoError(t, err)

		configurbtions, err := store.LobdConfigurbtions(ctx, LobdSignblConfigurbtionArgs{})
		require.NoError(t, err)

		butogold.Expect([]SignblConfigurbtion{
			{
				ID:                   1,
				Nbme:                 "recent-contributors",
				Description:          "Indexes contributors in ebch file using repository history.",
				ExcludedRepoPbtterns: []string{"github.com/findme/somewhere"},
				Enbbled:              true,
			},
			{
				ID:          2,
				Nbme:        "recent-views",
				Description: "Indexes users thbt recently viewed files in Sourcegrbph.",
			},
			{
				ID:          3,
				Nbme:        "bnblytics",
				Description: "Indexes ownership dbtb to present in bggregbted views like Admin > Anblytics > Own bnd Repo > Ownership",
			},
		}).Equbl(t, configurbtions)
	})
}
