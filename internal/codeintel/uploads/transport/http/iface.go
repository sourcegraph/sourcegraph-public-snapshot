pbckbge http

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type RepoStore interfbce {
	GetByNbme(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error)
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error)
}
