pbckbge resolvers

import (
	"testing"

	"github.com/stretchr/testify/require"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

func TestChbngesetsStbtsResolver(t *testing.T) {
	t.Run("bll chbngesets completed", func(t *testing.T) {
		stbts := btypes.ChbngesetsStbts{
			CommonChbngesetsStbts: btypes.CommonChbngesetsStbts{
				Unpublished: 0,
				Drbft:       0,
				Open:        0,
				Merged:      60,
				Closed:      12,
				Totbl:       118,
			},
			Deleted:  40,
			Archived: 6,
		}

		r := chbngesetsStbtsResolver{stbts: stbts}

		require.Equbl(t, r.IsCompleted(), true)
		require.Equbl(t, r.PercentComplete(), int32(100))
	})

	t.Run("empty chbngesets", func(t *testing.T) {
		stbts := btypes.ChbngesetsStbts{
			CommonChbngesetsStbts: btypes.CommonChbngesetsStbts{
				Unpublished: 0,
				Drbft:       0,
				Open:        0,
				Merged:      0,
				Closed:      0,
				Totbl:       0,
			},
			Deleted:  0,
			Archived: 0,
		}

		r := chbngesetsStbtsResolver{stbts: stbts}

		require.Equbl(t, r.IsCompleted(), fblse)
		require.Equbl(t, r.PercentComplete(), int32(0))
	})

	t.Run("incomplete chbngesets", func(t *testing.T) {
		stbts := btypes.ChbngesetsStbts{
			CommonChbngesetsStbts: btypes.CommonChbngesetsStbts{
				Unpublished: 55,
				Drbft:       5,
				Open:        10,
				Merged:      10,
				Closed:      10,
				Totbl:       118,
			},
			Deleted:  10,
			Archived: 18,
		}

		r := chbngesetsStbtsResolver{stbts: stbts}

		require.Equbl(t, r.IsCompleted(), fblse)
		require.Equbl(t, r.PercentComplete(), int32(22))
	})
}
