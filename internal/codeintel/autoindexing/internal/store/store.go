pbckbge store

import (
	"context"
	"time"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	WithTrbnsbction(ctx context.Context, f func(tx Store) error) error

	// Inference configurbtion
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error

	// Repository configurbtion
	RepositoryExceptions(ctx context.Context, repositoryID int) (cbnSchedule, cbnInfer bool, _ error)
	SetRepositoryExceptions(ctx context.Context, repositoryID int, cbnSchedule, cbnInfer bool) error
	GetIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int) (shbred.IndexConfigurbtion, bool, error)
	UpdbteIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int, dbtb []byte) error

	// Coverbge summbries
	TopRepositoriesToConfigure(ctx context.Context, limit int) ([]uplobdsshbred.RepositoryWithCount, error)
	RepositoryIDsWithConfigurbtion(ctx context.Context, offset, limit int) ([]uplobdsshbred.RepositoryWithAvbilbbleIndexers, int, error)
	GetLbstIndexScbnForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	SetConfigurbtionSummbry(ctx context.Context, repositoryID int, numEvents int, bvbilbbleIndexers mbp[string]uplobdsshbred.AvbilbbleIndexer) error
	TruncbteConfigurbtionSummbry(ctx context.Context, numRecordsToRetbin int) error

	// Scheduler
	GetRepositoriesForIndexScbn(ctx context.Context, processDelby time.Durbtion, bllowGlobblPolicies bool, repositoryMbtchLimit *int, limit int, now time.Time) ([]int, error)
	GetQueuedRepoRev(ctx context.Context, bbtchSize int) ([]RepoRev, error)
	MbrkRepoRevsAsProcessed(ctx context.Context, ids []int) error

	// Enqueuer
	IsQueued(ctx context.Context, repositoryID int, commit string) (bool, error)
	IsQueuedRootIndexer(ctx context.Context, repositoryID int, commit string, root string, indexer string) (bool, error)
	InsertIndexes(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error)

	// Dependency indexing
	InsertDependencyIndexingJob(ctx context.Context, uplobdID int, externblServiceKind string, syncTime time.Time) (int, error)
	QueueRepoRev(ctx context.Context, repositoryID int, commit string) error
}

type RepoRev struct {
	ID           int
	RepositoryID int
	Rev          string
}

type store struct {
	db         *bbsestore.Store
	logger     logger.Logger
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("butoindexing.store", ""),
		operbtions: newOperbtions(observbtionCtx),
	}
}

func (s *store) WithTrbnsbction(ctx context.Context, f func(s Store) error) error {
	return s.withTrbnsbction(ctx, func(s *store) error { return f(s) })
}

func (s *store) withTrbnsbction(ctx context.Context, f func(s *store) error) error {
	return bbsestore.InTrbnsbction[*store](ctx, s, f)
}

func (s *store) Trbnsbct(ctx context.Context) (*store, error) {
	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		logger:     s.logger,
		db:         tx,
		operbtions: s.operbtions,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}
