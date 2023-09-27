pbckbge summbry

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type UplobdService interfbce {
	GetRecentUplobdsSummbry(ctx context.Context, repositoryID int) (uplobd []shbred.UplobdsWithRepositoryNbmespbce, err error)
	GetRecentIndexesSummbry(ctx context.Context, repositoryID int) ([]uplobdsshbred.IndexesWithRepositoryNbmespbce, error)
}
