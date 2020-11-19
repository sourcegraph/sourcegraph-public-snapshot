package correlation

import (
	"sort"

	"github.com/google/uuid"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
)

func sortedRangeIDs(ranges map[lsifstore.ID]lsifstore.RangeData) []lsifstore.ID {
	var rngIDs []lsifstore.ID
	for rngID := range ranges {
		rngIDs = append(rngIDs, rngID)
	}

	sort.Slice(rngIDs, func(i, j int) bool {
		return lsifstore.CompareRanges(ranges[rngIDs[i]], ranges[rngIDs[j]]) < 0
	})

	return rngIDs
}

func getDefRef(resultID lsifstore.ID, meta lsifstore.MetaData, resultChunks map[int]lsifstore.ResultChunkData) ([]lsifstore.DocumentIDRangeID, lsifstore.ResultChunkData) {
	chunkID := lsifstore.HashKey(resultID, meta.NumResultChunks)
	chunk := resultChunks[chunkID]
	docRngIDs := chunk.DocumentIDRangeIDs[resultID]
	return docRngIDs, chunk
}

func newID() (lsifstore.ID, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return lsifstore.ID(uuid.String()), nil
}
