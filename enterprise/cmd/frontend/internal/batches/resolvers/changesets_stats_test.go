package resolvers

import (
	"testing"

	"github.com/stretchr/testify/require"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
)

func TestChangesetsStatsResolver(t *testing.T) {
	t.Run("all changesets completed", func(t *testing.T) {
		stats := btypes.ChangesetsStats{
			CommonChangesetsStats: btypes.CommonChangesetsStats{
				Unpublished: 0,
				Draft:       0,
				Open:        0,
				Merged:      60,
				Closed:      12,
				Total:       118,
			},
			Deleted:  40,
			Archived: 6,
		}

		r := changesetsStatsResolver{stats: stats}

		require.Equal(t, r.IsCompleted(), true)
		require.Equal(t, r.PercentComplete(), int32(100))
	})

	t.Run("empty changesets", func(t *testing.T) {
		stats := btypes.ChangesetsStats{
			CommonChangesetsStats: btypes.CommonChangesetsStats{
				Unpublished: 0,
				Draft:       0,
				Open:        0,
				Merged:      0,
				Closed:      0,
				Total:       0,
			},
			Deleted:  0,
			Archived: 0,
		}

		r := changesetsStatsResolver{stats: stats}

		require.Equal(t, r.IsCompleted(), false)
		require.Equal(t, r.PercentComplete(), int32(0))
	})

	t.Run("incomplete changesets", func(t *testing.T) {
		stats := btypes.ChangesetsStats{
			CommonChangesetsStats: btypes.CommonChangesetsStats{
				Unpublished: 55,
				Draft:       5,
				Open:        10,
				Merged:      10,
				Closed:      10,
				Total:       118,
			},
			Deleted:  10,
			Archived: 18,
		}

		r := changesetsStatsResolver{stats: stats}

		require.Equal(t, r.IsCompleted(), false)
		require.Equal(t, r.PercentComplete(), int32(22))
	})
}
