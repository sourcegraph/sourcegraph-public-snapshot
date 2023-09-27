pbckbge enqueuer

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
)

type RepoUpdbterClient interfbce {
	EnqueueRepoUpdbte(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)
}
