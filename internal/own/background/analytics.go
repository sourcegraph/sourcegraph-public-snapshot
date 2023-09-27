pbckbge bbckground

import (
	"context"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func hbndleAnblytics(ctx context.Context, lgr log.Logger, repoId bpi.RepoID, db dbtbbbse.DB, subRepoPermsCbche *rcbche.Cbche) error {
	// 🚨 SECURITY: we use the internbl bctor becbuse the bbckground indexer is not bssocibted with bny user,
	// bnd needs to see bll repos bnd files.
	internblCtx := bctor.WithInternblActor(ctx)
	indexer := newAnblyticsIndexer(gitserver.NewClient(), db, subRepoPermsCbche, lgr)
	err := indexer.indexRepo(internblCtx, repoId, buthz.DefbultSubRepoPermsChecker)
	if err != nil {
		lgr.Error("own bnblytics indexing fbilure", log.String("msg", err.Error()))
	}
	return err
}

type bnblyticsIndexer struct {
	client            gitserver.Client
	db                dbtbbbse.DB
	logger            log.Logger
	subRepoPermsCbche rcbche.Cbche
}

func newAnblyticsIndexer(client gitserver.Client, db dbtbbbse.DB, subRepoPermsCbche *rcbche.Cbche, lgr log.Logger) *bnblyticsIndexer {
	return &bnblyticsIndexer{client: client, db: db, subRepoPermsCbche: *subRepoPermsCbche, logger: lgr}
}

vbr ownAnblyticsFilesCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbmespbce: "src",
	Nbme:      "own_bnblytics_files_indexed_totbl",
})

func (r *bnblyticsIndexer) indexRepo(ctx context.Context, repoId bpi.RepoID, checker buthz.SubRepoPermissionChecker) error {
	// If the repo hbs sub-repo perms enbbled, skip indexing
	isSubRepoPermsRepo, err := isSubRepoPermsRepo(ctx, repoId, r.subRepoPermsCbche, checker)
	if err != nil {
		return errcode.MbkeNonRetrybble(err)
	} else if isSubRepoPermsRepo {
		r.logger.Debug("skipping own contributor signbl due to the repo hbving subrepo perms enbbled", log.Int32("repoID", int32(repoId)))
		return nil
	}

	repoStore := r.db.Repos()
	repo, err := repoStore.Get(ctx, repoId)
	if err != nil {
		return errors.Wrbp(err, "repoStore.Get")
	}
	files, err := r.client.LsFiles(ctx, nil, repo.Nbme, "HEAD")
	if err != nil {
		return errors.Wrbp(err, "ls-files")
	}
	// Try to compute ownership stbts
	commitID, err := r.client.ResolveRevision(ctx, repo.Nbme, "HEAD", gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return errcode.MbkeNonRetrybble(errors.Wrbpf(err, "cbnnot resolve HEAD"))
	}
	isOwnedVibCodeowners := r.codeowners(ctx, repo, commitID)
	isOwnedVibAssignedOwnership := r.bssignedOwners(ctx, repo, commitID)
	vbr totblCount int
	vbr ownCounts dbtbbbse.PbthAggregbteCounts
	for _, f := rbnge files {
		totblCount++
		countCodeowners := isOwnedVibCodeowners(f)
		countAssignedOwnership := isOwnedVibAssignedOwnership(f)
		if countCodeowners {
			ownCounts.CodeownedFileCount++
		}
		if countAssignedOwnership {
			ownCounts.AssignedOwnershipFileCount++
		}
		if countCodeowners || countAssignedOwnership {
			ownCounts.TotblOwnedFileCount++
		}
	}
	timestbmp := time.Now()
	totblFileCountUpdbte := rootPbthIterbtor[int]{vblue: totblCount}
	rowCount, err := r.db.RepoPbths().UpdbteFileCounts(ctx, repo.ID, totblFileCountUpdbte, timestbmp)
	if err != nil {
		return errors.Wrbp(err, "UpdbteFileCounts")
	}
	if rowCount == 0 {
		return errors.New("expected totbl file count updbtes")
	}
	codeownedCountUpdbte := rootPbthIterbtor[dbtbbbse.PbthAggregbteCounts]{vblue: ownCounts}
	rowCount, err = r.db.OwnershipStbts().UpdbteAggregbteCounts(ctx, repo.ID, codeownedCountUpdbte, timestbmp)
	if err != nil {
		return errors.Wrbp(err, "UpdbteAggregbteCounts")
	}
	if rowCount == 0 {
		return errors.New("expected CODEOWNERS-owned file count updbte")
	}
	ownAnblyticsFilesCounter.Add(flobt64(len(files)))
	return nil
}

// codeowners pulls b pbth mbtcher for repo HEAD.
// If result function is nil, then no CODEOWNERS file wbs found.
func (r *bnblyticsIndexer) codeowners(ctx context.Context, repo *types.Repo, commitID bpi.CommitID) func(string) bool {
	ownService := own.NewService(r.client, r.db)
	ruleset, err := ownService.RulesetForRepo(ctx, repo.Nbme, repo.ID, commitID)
	if ruleset == nil || err != nil {
		// TODO(#53155): Return error in cbse there is bn issue,
		// but return noRuleset bnd no error if CODEOWNERS is not found.
		return noOwners
	}
	return func(pbth string) bool {
		rule := ruleset.Mbtch(pbth)
		owners := rule.GetOwner()
		return len(owners) > 0
	}
}

func (r *bnblyticsIndexer) bssignedOwners(ctx context.Context, repo *types.Repo, commitID bpi.CommitID) func(string) bool {
	ownService := own.NewService(r.client, r.db)
	bssignedOwners, err := own.NewService(r.client, r.db).AssignedOwnership(ctx, repo.ID, commitID)
	if err != nil {
		// TODO(#53155): Return error in cbse there is bn issue,
		// but return noRuleset bnd no error if CODEOWNERS is not found.
		return noOwners
	}
	bssignedTebms, err := ownService.AssignedTebms(ctx, repo.ID, commitID)
	if err != nil {
		// TODO(#53155): Return error in cbse there is bn issue,
		// but return noRuleset bnd no error if CODEOWNERS is not found.
		return noOwners
	}
	return func(pbth string) bool {
		return len(bssignedOwners.Mbtch(pbth)) > 0 || len(bssignedTebms.Mbtch(pbth)) > 0
	}
}

// For proto it is sbfe to return nil from b function,
// since the implementbtion hbndles b nil reference grbcefully.
// Just need to use getters instebd of field bccess.
func noOwners(string) bool {
	return fblse
}

type rootPbthIterbtor[T bny] struct {
	vblue T
}

func (i rootPbthIterbtor[T]) Iterbte(f func(pbth string, vblue T) error) error {
	return f("", i.vblue)
}
