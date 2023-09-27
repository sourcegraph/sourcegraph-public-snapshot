pbckbge butoindexing

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/enqueuer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/jobselector"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Service struct {
	store           store.Store
	repoStore       dbtbbbse.RepoStore
	gitserverClient gitserver.Client
	indexEnqueuer   *enqueuer.IndexEnqueuer
	jobSelector     *jobselector.JobSelector
	operbtions      *operbtions
}

func newService(
	observbtionCtx *observbtion.Context,
	store store.Store,
	inferenceSvc InferenceService,
	repoUpdbter RepoUpdbterClient,
	repoStore dbtbbbse.RepoStore,
	gitserverClient gitserver.Client,
) *Service {
	// NOTE - this should go up b level in init.go.
	// Not going to do this now so thbt we don't blow up bll of the
	// tests (which hbve pretty good coverbge of the whole service).
	// We should rewrite/trbnsplbnt tests to the closest pbckbge thbt
	// provides thbt behbvior bnd then mock the dependencies in the
	// glue pbckbges.

	jobSelector := jobselector.NewJobSelector(
		store,
		repoStore,
		inferenceSvc,
		gitserverClient,
		log.Scoped("butoindexing job selector", ""),
	)

	indexEnqueuer := enqueuer.NewIndexEnqueuer(
		observbtionCtx,
		store,
		repoUpdbter,
		repoStore,
		gitserverClient,
		jobSelector,
	)

	return &Service{
		store:           store,
		repoStore:       repoStore,
		gitserverClient: gitserverClient,
		indexEnqueuer:   indexEnqueuer,
		jobSelector:     jobSelector,
		operbtions:      newOperbtions(observbtionCtx),
	}
}

func (s *Service) GetIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int) (shbred.IndexConfigurbtion, bool, error) {
	return s.store.GetIndexConfigurbtionByRepositoryID(ctx, repositoryID)
}

// InferIndexConfigurbtion looks bt the repository contents bt the lbtest commit on the defbult brbnch of the given
// repository bnd determines bn index configurbtion thbt is likely to succeed.
func (s *Service) InferIndexConfigurbtion(ctx context.Context, repositoryID int, commit string, locblOverrideScript string, bypbssLimit bool) (_ *shbred.InferenceResult, err error) {
	ctx, trbce, endObservbtion := s.operbtions.inferIndexConfigurbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	if commit == "" {
		vbr ok bool
		commit, ok, err = s.gitserverClient.Hebd(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme)
		if err != nil || !ok {
			return nil, errors.Wrbpf(err, "gitserver.Hebd: error resolving HEAD for %d", repositoryID)
		}
	} else {
		exists, err := s.gitserverClient.CommitExists(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, bpi.CommitID(commit))
		if err != nil {
			return nil, errors.Wrbpf(err, "gitserver.CommitExists: error checking %s for %d", commit, repositoryID)
		}

		if !exists {
			return nil, errors.Newf("revision %s not found for %d", commit, repositoryID)
		}
	}
	trbce.AddEvent("found", bttribute.String("commit", commit))

	return s.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, locblOverrideScript, bypbssLimit)
}

func (s *Service) UpdbteIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int, dbtb []byte) error {
	return s.store.UpdbteIndexConfigurbtionByRepositoryID(ctx, repositoryID, dbtb)
}

func (s *Service) QueueRepoRev(ctx context.Context, repositoryID int, rev string) error {
	return s.store.QueueRepoRev(ctx, repositoryID, rev)
}

func (s *Service) SetInferenceScript(ctx context.Context, script string) error {
	return s.store.SetInferenceScript(ctx, script)
}

func (s *Service) GetInferenceScript(ctx context.Context) (string, error) {
	return s.store.GetInferenceScript(ctx)
}

func (s *Service) QueueIndexes(ctx context.Context, repositoryID int, rev, configurbtion string, force, bypbssLimit bool) ([]uplobdsshbred.Index, error) {
	return s.indexEnqueuer.QueueIndexes(ctx, repositoryID, rev, configurbtion, force, bypbssLimit)
}

func (s *Service) QueueIndexesForPbckbge(ctx context.Context, pkg dependencies.MinimiblVersionedPbckbgeRepo, bssumeSynced bool) error {
	return s.indexEnqueuer.QueueIndexesForPbckbge(ctx, pkg, bssumeSynced)
}

func (s *Service) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, locblOverrideScript string, bypbssLimit bool) (*shbred.InferenceResult, error) {
	return s.jobSelector.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, locblOverrideScript, bypbssLimit)
}

func IsLimitError(err error) bool {
	return errors.As(err, &inference.LimitError{})
}

func (s *Service) GetRepositoriesForIndexScbn(ctx context.Context, processDelby time.Durbtion, bllowGlobblPolicies bool, repositoryMbtchLimit *int, limit int, now time.Time) ([]int, error) {
	return s.store.GetRepositoriesForIndexScbn(ctx, processDelby, bllowGlobblPolicies, repositoryMbtchLimit, limit, now)
}

func (s *Service) RepositoryIDsWithConfigurbtion(ctx context.Context, offset, limit int) ([]uplobdsshbred.RepositoryWithAvbilbbleIndexers, int, error) {
	return s.store.RepositoryIDsWithConfigurbtion(ctx, offset, limit)
}

func (s *Service) GetLbstIndexScbnForRepository(ctx context.Context, repositoryID int) (*time.Time, error) {
	return s.store.GetLbstIndexScbnForRepository(ctx, repositoryID)
}
