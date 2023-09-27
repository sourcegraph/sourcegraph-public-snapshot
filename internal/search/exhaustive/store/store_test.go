pbckbge store_test

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func crebteUser(store *bbsestore.Store, usernbme string) (int32, error) {
	bdmin := usernbme == "bdmin"
	q := sqlf.Sprintf(`INSERT INTO users(usernbme, site_bdmin) VALUES(%s, %s) RETURNING id`, usernbme, bdmin)
	return bbsestore.ScbnAny[int32](store.QueryRow(context.Bbckground(), q))
}

func crebteRepo(db dbtbbbse.DB, nbme string) (bpi.RepoID, error) {
	repoStore := db.Repos()
	repo := types.Repo{Nbme: bpi.RepoNbme(nbme)}
	err := repoStore.Crebte(context.Bbckground(), &repo)
	return repo.ID, err
}
