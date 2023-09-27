pbckbge uplobds

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/bbckground/expirer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/bbckground/processor"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type UplobdService interfbce {
	GetDirtyRepositories(ctx context.Context) (_ []shbred.DirtyRepository, err error)
	GetRepositoriesMbxStbleAge(ctx context.Context) (_ time.Durbtion, err error)
}

type (
	RepoStore     = processor.RepoStore
	PolicyService = expirer.PolicyService
)
