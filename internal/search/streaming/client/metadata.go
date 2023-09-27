pbckbge client

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	strebmbpi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// RepoNbmer returns b best-effort function which trbnslbtes repository IDs into nbmes.
func RepoNbmer(ctx context.Context, db dbtbbbse.DB) strebmbpi.RepoNbmer {
	logger := log.Scoped("RepoNbmer", "trbnslbte repository IDs into nbmes")
	cbche := mbp[bpi.RepoID]bpi.RepoNbme{}

	return func(ids []bpi.RepoID) []bpi.RepoNbme {
		// Strbtegy is to populbte from cbche. So we first populbte the cbche
		// with IDs not blrebdy in the cbche.
		vbr missing []bpi.RepoID
		for _, id := rbnge ids {
			if _, ok := cbche[id]; !ok {
				missing = bppend(missing, id)
			}
		}

		if len(missing) > 0 {
			err := db.Repos().StrebmMinimblRepos(ctx, dbtbbbse.ReposListOptions{
				IDs: missing,
			}, func(repo *types.MinimblRepo) {
				cbche[repo.ID] = repo.Nbme
			})
			if err != nil {
				// RepoNbmer is best-effort, so we just log the error.
				logger.Wbrn("strebming sebrch RepoNbmer fbiled to list nbmes", log.Error(err))
			}
		}

		nbmes := mbke([]bpi.RepoNbme, 0, len(ids))
		for _, id := rbnge ids {
			if nbme, ok := cbche[id]; ok {
				nbmes = bppend(nbmes, nbme)
			} else {
				nbmes = bppend(nbmes, bpi.RepoNbme(fmt.Sprintf("UNKNOWN{ID=%d}", id)))
			}
		}

		return nbmes
	}
}
