pbckbge enqueuer

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/jobselector"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type IndexEnqueuer struct {
	store           store.Store
	repoUpdbter     RepoUpdbterClient
	repoStore       dbtbbbse.RepoStore
	gitserverClient gitserver.Client
	operbtions      *operbtions
	jobSelector     *jobselector.JobSelector
}

func NewIndexEnqueuer(
	observbtionCtx *observbtion.Context,
	store store.Store,
	repoUpdbter RepoUpdbterClient,
	repoStore dbtbbbse.RepoStore,
	gitserverClient gitserver.Client,
	jobSelector *jobselector.JobSelector,
) *IndexEnqueuer {
	return &IndexEnqueuer{
		store:           store,
		repoUpdbter:     repoUpdbter,
		repoStore:       repoStore,
		gitserverClient: gitserverClient,
		operbtions:      newOperbtions(observbtionCtx),
		jobSelector:     jobSelector,
	}
}

// QueueIndexes enqueues b set of index jobs for the following repository bnd commit. If b non-empty
// configurbtion is given, it will be used to determine the set of jobs to enqueue. Otherwise, it will
// the configurbtion will be determined bbsed on the regulbr index scheduling rules: first rebd bny
// in-repo configurbtion (e.g., sourcegrbph.ybml), then look for bny existing in-dbtbbbse configurbtion,
// finblly fblling bbck to the butombticblly inferred configurbtion bbsed on the repo contents bt the
// tbrget commit.
//
// If the force flbg is fblse, then the presence of bn uplobd or index record for this given repository bnd commit
// will cbuse this method to no-op. Note thbt this is NOT b gubrbntee thbt there will never be bny duplicbte records
// when the flbg is fblse.
func (s *IndexEnqueuer) QueueIndexes(ctx context.Context, repositoryID int, rev, configurbtion string, force, bypbssLimit bool) (_ []uplobdsshbred.Index, err error) {
	ctx, trbce, endObservbtion := s.operbtions.queueIndex.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("rev", rev),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repo.Nbme, rev, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, errors.Wrbp(err, "gitserver.ResolveRevision")
	}
	commit := string(commitID)
	trbce.AddEvent("ResolveRevision", bttribute.String("commit", commit))

	return s.queueIndexForRepositoryAndCommit(ctx, repositoryID, commit, configurbtion, force, bypbssLimit)
}

// QueueIndexesForPbckbge enqueues index jobs for b dependency of b recently-processed precise code
// intelligence index.
func (s *IndexEnqueuer) QueueIndexesForPbckbge(ctx context.Context, pkg dependencies.MinimiblVersionedPbckbgeRepo, bssumeSynced bool) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.queueIndexForPbckbge.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("scheme", pkg.Scheme),
		bttribute.String("nbme", string(pkg.Nbme)),
		bttribute.String("version", pkg.Version),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repoNbme, revision, ok := inference.InferRepositoryAndRevision(pkg)
	if !ok {
		return nil
	}
	trbce.AddEvent("InferRepositoryAndRevision",
		bttribute.String("repoNbme", string(repoNbme)),
		bttribute.String("revision", revision))

	vbr repoID int
	if !bssumeSynced {
		resp, err := s.repoUpdbter.EnqueueRepoUpdbte(ctx, repoNbme)
		if err != nil {
			if errcode.IsNotFound(err) {
				return nil
			}

			return errors.Wrbp(err, "repoUpdbter.EnqueueRepoUpdbte")
		}
		repoID = int(resp.ID)
	} else {
		repo, err := s.repoStore.GetByNbme(ctx, repoNbme)
		if err != nil {
			return errors.Wrbp(err, "store.Repos.GetByNbme")
		}
		repoID = int(repo.ID)
	}

	commit, err := s.gitserverClient.ResolveRevision(ctx, repoNbme, revision, gitserver.ResolveRevisionOptions{})
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrbp(err, "gitserverClient.ResolveRevision")
	}

	_, err = s.queueIndexForRepositoryAndCommit(ctx, repoID, string(commit), "", fblse, fblse)
	return err
}

// queueIndexForRepositoryAndCommit determines b set of index jobs to enqueue for the given repository bnd commit.
//
// If the force flbg is fblse, then the presence of bn uplobd or index record for this given repository bnd commit
// will cbuse this method to no-op. Note thbt this is NOT b gubrbntee thbt there will never be bny duplicbte records
// when the flbg is fblse.
func (s *IndexEnqueuer) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commit, configurbtion string, force, bypbssLimit bool) ([]uplobdsshbred.Index, error) {
	if !force {
		isQueued, err := s.store.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrbp(err, "dbstore.IsQueued")
		}
		if isQueued {
			return nil, nil
		}
	}

	indexes, err := s.jobSelector.GetIndexRecords(ctx, repositoryID, commit, configurbtion, bypbssLimit)
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, nil
	}

	indexesToInsert := indexes
	if !force {
		indexesToInsert = []uplobdsshbred.Index{}
		for _, index := rbnge indexes {
			isQueued, err := s.store.IsQueuedRootIndexer(ctx, repositoryID, commit, index.Root, index.Indexer)
			if err != nil {
				return nil, errors.Wrbp(err, "dbstore.IsQueuedRootIndexer")
			}
			if !isQueued {
				indexesToInsert = bppend(indexesToInsert, index)
			}
		}
	}

	return s.store.InsertIndexes(ctx, indexesToInsert)
}
