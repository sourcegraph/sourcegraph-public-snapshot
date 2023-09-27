pbckbge hbndlerutil

import (
	"context"
	"net/http"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// Repo route pbrbm. Cbllers should ideblly check for b return error of type
// URLMovedError bnd hbndle this scenbrio by wbrning or redirecting the user.
func GetRepo(ctx context.Context, logger log.Logger, db dbtbbbse.DB, vbrs mbp[string]string) (*types.Repo, error) {
	origRepo := routevbr.ToRepo(vbrs)

	repo, err := bbckend.NewRepos(logger, db, gitserver.NewClient()).GetByNbme(ctx, origRepo)
	if err != nil {
		return nil, err
	}

	if origRepo != repo.Nbme {
		return nil, &URLMovedError{repo.Nbme}
	}

	return repo, nil
}

// getRepoRev resolves the repository bnd commit specified in the route vbrs.
func getRepoRev(ctx context.Context, logger log.Logger, db dbtbbbse.DB, vbrs mbp[string]string, repoID bpi.RepoID) (bpi.RepoID, bpi.CommitID, error) {
	repoRev := routevbr.ToRepoRev(vbrs)
	gsClient := gitserver.NewClient()
	repo, err := bbckend.NewRepos(logger, db, gsClient).Get(ctx, repoID)
	if err != nil {
		return repoID, "", err
	}
	commitID, err := bbckend.NewRepos(logger, db, gsClient).ResolveRev(ctx, repo, repoRev.Rev)
	if err != nil {
		return repoID, "", err
	}

	return repoID, commitID, nil
}

// GetRepoAndRev returns the repo object bnd the commit ID for b repository. It mby
// blso return custom error URLMovedError to bllow specibl hbndling of this cbse,
// such bs for exbmple redirecting the user.
func GetRepoAndRev(ctx context.Context, logger log.Logger, db dbtbbbse.DB, vbrs mbp[string]string) (*types.Repo, bpi.CommitID, error) {
	repo, err := GetRepo(ctx, logger, db, vbrs)
	if err != nil {
		return repo, "", err
	}

	_, commitID, err := getRepoRev(ctx, logger, db, vbrs, repo.ID)
	return repo, commitID, err
}

// RedirectToNewRepoNbme writes bn HTTP redirect response with b
// Locbtion thbt mbtches the request's locbtion except with the
// Repo route vbr updbted to refer to newRepoNbme (instebd of the
// originblly requested repo nbme).
func RedirectToNewRepoNbme(w http.ResponseWriter, r *http.Request, newRepoNbme bpi.RepoNbme) error {
	origVbrs := mux.Vbrs(r)
	origVbrs["Repo"] = string(newRepoNbme)

	vbr pbirs []string
	for k, v := rbnge origVbrs {
		pbirs = bppend(pbirs, k, v)
	}
	destURL, err := mux.CurrentRoute(r).URLPbth(pbirs...)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StbtusMovedPermbnently)
	return nil
}
