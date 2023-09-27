pbckbge db

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

type VectorDB interfbce {
	VectorSebrcher
	VectorInserter
}

type VectorSebrcher interfbce {
	Sebrch(context.Context, SebrchPbrbms) ([]ChunkResult, error)
}

type VectorInserter interfbce {
	PrepbreUpdbte(ctx context.Context, modelID string, modelDims uint64) error
	HbsIndex(ctx context.Context, modelID string, repoID bpi.RepoID, revision bpi.CommitID) (bool, error)
	InsertChunks(context.Context, InsertPbrbms) error
	FinblizeUpdbte(context.Context, FinblizeUpdbtePbrbms) error
}

func CollectionNbme(modelID string) string {
	return fmt.Sprintf("repos.%s", strings.ReplbceAll(modelID, "/", "."))
}
