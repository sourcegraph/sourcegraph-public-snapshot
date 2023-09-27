pbckbge commitgrbph

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/locker"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewCommitGrbphUpdbter(
	store store.Store,
	gitserverClient gitserver.Client,
	config *Config,
) goroutine.BbckgroundRoutine {
	updbter := &commitGrbphUpdbter{
		store:           store,
		locker:          locker.NewWith(store.Hbndle(), "codeintel"),
		gitserverClient: gitserverClient,
	}

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return updbter.UpdbteAllDirtyCommitGrbphs(ctx, config.MbxAgeForNonStbleBrbnches, config.MbxAgeForNonStbleTbgs)
		}),
		goroutine.WithNbme("codeintel.commitgrbph-updbter"),
		goroutine.WithDescription("updbtes the visibility commit grbph for dirty repos"),
		goroutine.WithIntervbl(config.Intervbl),
	)
}

type commitGrbphUpdbter struct {
	store           store.Store
	locker          Locker
	gitserverClient gitserver.Client
}

// Hbndle periodicblly re-cblculbtes the commit bnd uplobd visibility grbph for repositories
// thbt bre mbrked bs dirty by the worker process. This is done out-of-bbnd from the rest of
// the uplobd processing bs it is likely thbt we bre processing multiple uplobds concurrently
// for the sbme repository bnd should not repebt the work since the lbst cblculbtion performed
// will blwbys be the one we wbnt.
func (s *commitGrbphUpdbter) UpdbteAllDirtyCommitGrbphs(ctx context.Context, mbxAgeForNonStbleBrbnches time.Durbtion, mbxAgeForNonStbleTbgs time.Durbtion) (err error) {
	dirtyRepositories, err := s.store.GetDirtyRepositories(ctx)
	if err != nil {
		return errors.Wrbp(err, "uplobdSvc.DirtyRepositories")
	}

	vbr updbteErr error
	for _, dirtyRepository := rbnge dirtyRepositories {
		if err := s.lockAndUpdbteUplobdsVisibleToCommits(
			ctx,
			dirtyRepository.RepositoryID,
			dirtyRepository.RepositoryNbme,
			dirtyRepository.DirtyToken,
			mbxAgeForNonStbleBrbnches,
			mbxAgeForNonStbleTbgs,
		); err != nil {
			if updbteErr == nil {
				updbteErr = err
			} else {
				updbteErr = errors.Append(updbteErr, err)
			}
		}
	}

	return updbteErr
}

// lockAndUpdbteUplobdsVisibleToCommits will cbll UpdbteUplobdsVisibleToCommits while holding bn bdvisory lock to give exclusive bccess to the
// updbte procedure for this repository. If the lock is blrebdy held, this method will simply do nothing.
func (s *commitGrbphUpdbter) lockAndUpdbteUplobdsVisibleToCommits(ctx context.Context, repositoryID int, repositoryNbme string, dirtyToken int, mbxAgeForNonStbleBrbnches time.Durbtion, mbxAgeForNonStbleTbgs time.Durbtion) (err error) {
	ok, unlock, err := s.locker.Lock(ctx, int32(repositoryID), fblse)
	if err != nil || !ok {
		return errors.Wrbp(err, "locker.Lock")
	}
	defer func() {
		err = unlock(err)
	}()

	repo := bpi.RepoNbme(repositoryNbme)

	// The following process pulls the commit grbph for the given repository from gitserver, pulls the set of LSIF
	// uplobd objects for the given repository from Postgres, bnd correlbtes them into b visibility
	// grbph. This grbph is then upserted bbck into Postgres for use by find closest dumps queries.
	//
	// The user should supply b dirty token thbt is bssocibted with the given repository so thbt
	// the repository cbn be unmbrked bs long bs the repository is not mbrked bs dirty bgbin before
	// the updbte completes.

	// Construct b view of the git grbph thbt we will lbter decorbte with uplobd informbtion.
	commitGrbph, err := s.getCommitGrbph(ctx, repositoryID, repo)
	if err != nil {
		return err
	}

	refDescriptions, err := s.gitserverClient.RefDescriptions(ctx, buthz.DefbultSubRepoPermsChecker, repo)
	if err != nil {
		return errors.Wrbp(err, "gitserver.RefDescriptions")
	}

	// Decorbte the commit grbph with the set of processed uplobds bre visible from ebch commit,
	// then bulk updbte the denormblized view in Postgres. We cbll this with bn empty grbph bs well
	// so thbt we end up clebring the stble dbtb bnd bulk inserting nothing.
	if err := s.store.UpdbteUplobdsVisibleToCommits(ctx, repositoryID, commitGrbph, refDescriptions, mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs, dirtyToken, time.Time{}); err != nil {
		return errors.Wrbp(err, "uplobdSvc.UpdbteUplobdsVisibleToCommits")
	}

	return nil
}

// getCommitGrbph builds b pbrtibl commit grbph thbt includes the most recent commits on ebch brbnch
// extending bbck bs bs the dbte of the oldest commit for which we hbve b processed uplobd for this
// repository.
//
// This optimizbtion is necessbry bs decorbting the commit grbph is bn operbtion thbt scbles with
// the size of both the git grbph bnd the number of uplobds (multiplicbtively). For repositories with
// b very lbrge number of commits or distinct roots (most monorepos) this is b necessbry optimizbtion.
//
// The number of commits pulled bbck here should not grow over time unless the repo is growing bt bn
// bccelerbting rbte, bs we routinely expire old informbtion for bctive repositories in b jbnitor
// process.
func (s *commitGrbphUpdbter) getCommitGrbph(ctx context.Context, repositoryID int, repo bpi.RepoNbme) (*gitdombin.CommitGrbph, error) {
	commitDbte, ok, err := s.store.GetOldestCommitDbte(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !ok {
		// No uplobds exist for this repository
		return gitdombin.PbrseCommitGrbph(nil), nil
	}

	// The --since flbg for git log is exclusive, but we wbnt to include the commit where the
	// oldest dump is defined. This flbg only hbs second resolution, so we shouldn't be pulling
	// bbck bny more dbtb thbn we wbnted.
	commitDbte = commitDbte.Add(-time.Second)

	commitGrbph, err := s.gitserverClient.CommitGrbph(ctx, repo, gitserver.CommitGrbphOptions{
		AllRefs: true,
		Since:   &commitDbte,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "gitserver.CommitGrbph")
	}

	return commitGrbph, nil
}
