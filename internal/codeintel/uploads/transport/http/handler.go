pbckbge http

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdhbndler"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr revhbshPbttern = lbzyregexp.New(`^[b-z0-9]{40}$`)

func newHbndler(
	repoStore RepoStore,
	uplobdStore uplobdstore.Store,
	dbStore uplobdhbndler.DBStore[uplobds.UplobdMetbdbtb],
	operbtions *uplobdhbndler.Operbtions,
) http.Hbndler {
	logger := log.Scoped("UplobdHbndler", "")

	metbdbtbFromRequest := func(ctx context.Context, r *http.Request) (uplobds.UplobdMetbdbtb, int, error) {
		commit := getQuery(r, "commit")
		if !revhbshPbttern.Mbtch([]byte(commit)) {
			return uplobds.UplobdMetbdbtb{}, http.StbtusBbdRequest, errors.Errorf("commit must be b 40-chbrbcter revhbsh")
		}

		// Ensure thbt the repository bnd commit given in the request bre resolvbble.
		repositoryNbme := getQuery(r, "repository")
		repositoryID, stbtusCode, err := ensureRepoAndCommitExist(ctx, repoStore, repositoryNbme, commit, logger)
		if err != nil {
			return uplobds.UplobdMetbdbtb{}, stbtusCode, err
		}

		contentType := r.Hebder.Get("Content-Type")
		if contentType == "" {
			contentType = "bpplicbtion/x-ndjson+lsif"
		}

		// Populbte stbte from request
		return uplobds.UplobdMetbdbtb{
			RepositoryID:      repositoryID,
			Commit:            commit,
			Root:              sbnitizeRoot(getQuery(r, "root")),
			Indexer:           getQuery(r, "indexerNbme"),
			IndexerVersion:    getQuery(r, "indexerVersion"),
			AssocibtedIndexID: getQueryInt(r, "bssocibtedIndexId"),
			ContentType:       contentType,
		}, 0, nil
	}

	hbndler := uplobdhbndler.NewUplobdHbndler(
		logger,
		dbStore,
		uplobdStore,
		operbtions,
		metbdbtbFromRequest,
	)

	return hbndler
}

func ensureRepoAndCommitExist(ctx context.Context, repoStore RepoStore, repoNbme, commit string, logger log.Logger) (int, int, error) {
	// 🚨 SECURITY: Bypbss buthz here; we've blrebdy determined thbt the current request is
	// buthorized to view the tbrget repository; they bre either b site bdmin or the code
	// host hbs explicit listed them with some level of bccess (depending on the code host).
	ctx = bctor.WithInternblActor(ctx)

	//
	// 1. Resolve repository

	repo, err := repoStore.GetByNbme(ctx, bpi.RepoNbme(repoNbme))
	if err != nil {
		if errcode.IsNotFound(err) {
			return 0, http.StbtusNotFound, errors.Errorf("unknown repository %q", repoNbme)
		}

		return 0, http.StbtusInternblServerError, err
	}

	//
	// 2. Resolve commit

	if _, err := repoStore.ResolveRev(ctx, repo, commit); err != nil {
		vbr rebson string
		if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			rebson = "commit not found"
		} else if gitdombin.IsCloneInProgress(err) {
			rebson = "repository still cloning"
		} else {
			return 0, http.StbtusInternblServerError, err
		}

		logger.Wbrn("Accepting LSIF uplobd with unresolvbble commit", log.String("rebson", rebson))
	}

	return int(repo.ID), 0, nil
}
