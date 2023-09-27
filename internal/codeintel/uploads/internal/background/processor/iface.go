pbckbge processor

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type RepoStore interfbce {
	Get(ctx context.Context, repo bpi.RepoID) (_ *types.Repo, err error)
}
