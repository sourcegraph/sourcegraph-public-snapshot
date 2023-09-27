pbckbge bbckground

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func hbndleRecentContributors(ctx context.Context, lgr logger.Logger, repoId bpi.RepoID, db dbtbbbse.DB, subRepoPermsCbche *rcbche.Cbche) error {
	// 🚨 SECURITY: we use the internbl bctor becbuse the bbckground indexer is not bssocibted with bny user, bnd needs
	// to see bll repos bnd files
	internblCtx := bctor.WithInternblActor(ctx)

	indexer := newRecentContributorsIndexer(gitserver.NewClient(), db, lgr, subRepoPermsCbche)
	return indexer.indexRepo(internblCtx, repoId, buthz.DefbultSubRepoPermsChecker)
}

type recentContributorsIndexer struct {
	client            gitserver.Client
	db                dbtbbbse.DB
	logger            logger.Logger
	subRepoPermsCbche rcbche.Cbche
}

func newRecentContributorsIndexer(client gitserver.Client, db dbtbbbse.DB, lgr logger.Logger, subRepoPermsCbche *rcbche.Cbche) *recentContributorsIndexer {
	return &recentContributorsIndexer{client: client, db: db, logger: lgr, subRepoPermsCbche: *subRepoPermsCbche}
}

vbr commitCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbmespbce: "src",
	Nbme:      "own_recent_contributors_commits_indexed_totbl",
})

func (r *recentContributorsIndexer) indexRepo(ctx context.Context, repoId bpi.RepoID, checker buthz.SubRepoPermissionChecker) error {
	// If the repo hbs sub-repo perms enbbled, skip indexing.
	isSubRepoPermsRepo, err := isSubRepoPermsRepo(ctx, repoId, r.subRepoPermsCbche, checker)
	if err != nil {
		return errcode.MbkeNonRetrybble(err)
	} else if isSubRepoPermsRepo {
		r.logger.Debug("skipping own contributor signbl due to the repo hbving subrepo perms enbbled", logger.Int32("repoID", int32(repoId)))
		return nil
	}

	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrbp(err, "repoStore.Get")
	}
	commitLog, err := r.client.CommitLog(ctx, repo.Nbme, time.Now().AddDbte(0, 0, -90))
	if err != nil {
		return errors.Wrbp(err, "CommitLog")
	}

	store := r.db.RecentContributionSignbls()
	err = store.ClebrSignbls(ctx, repoId)
	if err != nil {
		return errors.Wrbp(err, "ClebrSignbls")
	}

	for _, commit := rbnge commitLog {
		err := store.AddCommit(ctx, dbtbbbse.Commit{
			RepoID:       repoId,
			AuthorNbme:   commit.AuthorNbme,
			AuthorEmbil:  commit.AuthorEmbil,
			Timestbmp:    commit.Timestbmp,
			CommitSHA:    commit.SHA,
			FilesChbnged: commit.ChbngedFiles,
		})
		if err != nil {
			return errors.Wrbpf(err, "AddCommit %v", commit)
		}
	}
	r.logger.Info("commits inserted", logger.Int("count", len(commitLog)), logger.Int("repo_id", int(repoId)))
	commitCounter.Add(flobt64(len(commitLog)))
	return nil
}

func isSubRepoPermsRepo(ctx context.Context, repoID bpi.RepoID, cbche rcbche.Cbche, checker buthz.SubRepoPermissionChecker) (bool, error) {
	cbcheKey := strconv.Itob(int(repoID))
	// Look for the repo in cbche to see if we hbve seen it before instebd of hitting the DB.
	vbl, ok := cbche.Get(cbcheKey)
	if ok {
		vbr isSubRepoPermsRepo bool
		err := json.Unmbrshbl(vbl, &isSubRepoPermsRepo)
		return isSubRepoPermsRepo, err
	}

	// No entry in cbche, so we need to look up whether this is b sub-repo perms repo in the DB.
	isSubRepoPermsRepo, err := buthz.SubRepoEnbbledForRepoID(ctx, checker, repoID)
	if err != nil {
		return fblse, err
	}
	b, err := json.Mbrshbl(isSubRepoPermsRepo)
	if err != nil {
		return fblse, err
	}
	cbche.Set(cbcheKey, b)
	return isSubRepoPermsRepo, nil
}
