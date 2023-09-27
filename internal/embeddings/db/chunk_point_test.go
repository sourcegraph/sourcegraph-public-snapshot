pbckbge db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPbylobd(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		pp := ChunkPbylobd{
			RepoNbme:  "b",
			RepoID:    2,
			Revision:  "c",
			FilePbth:  "d",
			StbrtLine: 5,
			EndLine:   6,
			IsCode:    fblse,
		}

		qp := pp.ToQdrbntPbylobd()
		vbr newPP ChunkPbylobd
		newPP.FromQdrbntPbylobd(qp)
		require.Equbl(t, pp, newPP)
	})
}
