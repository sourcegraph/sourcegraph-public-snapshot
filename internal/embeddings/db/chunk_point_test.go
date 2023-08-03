package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayload(t *testing.T) {
	t.Run("roundtrip", func(t *testing.T) {
		pp := ChunkPayload{
			RepoName:  "a",
			RepoID:    2,
			Revision:  "c",
			FilePath:  "d",
			StartLine: 5,
			EndLine:   6,
			IsCode:    false,
		}

		qp := pp.ToQdrantPayload()
		var newPP ChunkPayload
		newPP.FromQdrantPayload(qp)
		require.Equal(t, pp, newPP)
	})
}
