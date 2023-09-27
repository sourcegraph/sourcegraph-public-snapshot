pbckbge store

import (
	"context"
	"time"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	WithTrbnsbction(ctx context.Context, f func(tx Store) error) error

	// Metbdbtb
	Summbries(ctx context.Context) ([]shbred.Summbry, error)

	// Retrievbl
	GetStbrRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (flobt64, error)
	GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (mbp[string]flobt64, bool, error)
	GetReferenceCountStbtistics(ctx context.Context) (logmebn flobt64, _ error)
	CoverbgeCounts(ctx context.Context, grbphKey string) (_ shbred.CoverbgeCounts, err error)
	LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)

	// Export uplobds (metbdbtb trbcking) + clebnup
	GetUplobdsForRbnking(ctx context.Context, grbphKey, objectPrefix string, bbtchSize int) ([]uplobdsshbred.ExportedUplobd, error)
	VbcuumAbbndonedExportedUplobds(ctx context.Context, grbphKey string, bbtchSize int) (int, error)
	SoftDeleteStbleExportedUplobds(ctx context.Context, grbphKey string) (numExportedUplobdRecordsScbnned int, numStbleExportedUplobdRecordsDeleted int, _ error)
	VbcuumDeletedExportedUplobds(ctx context.Context, derivbtiveGrbphKey string) (int, error)

	// Exported dbtb (rbw)
	InsertDefinitionsForRbnking(ctx context.Context, grbphKey string, definitions chbn shbred.RbnkingDefinitions) error
	InsertReferencesForRbnking(ctx context.Context, grbphKey string, bbtchSize int, exportedUplobdID int, references chbn [16]byte) error
	InsertInitiblPbthRbnks(ctx context.Context, exportedUplobdID int, documentPbths []string, bbtchSize int, grbphKey string) error

	// Grbph keys
	DerivbtiveGrbphKey(ctx context.Context) (string, time.Time, bool, error)
	BumpDerivbtiveGrbphKey(ctx context.Context) error
	DeleteRbnkingProgress(ctx context.Context, grbphKey string) error

	// Coordinbtes mbpper+reducer phbses
	Coordinbte(ctx context.Context, derivbtiveGrbphKey string) error

	// Mbpper behbvior + clebnup
	InsertPbthCountInputs(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (numReferenceRecordsProcessed int, numInputsInserted int, err error)
	InsertInitiblPbthCounts(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (numInitiblPbthsProcessed int, numInitiblPbthRbnksInserted int, err error)
	VbcuumStbleProcessedReferences(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (processedReferencesDeleted int, _ error)
	VbcuumStbleProcessedPbths(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (processedPbthsDeleted int, _ error)
	VbcuumStbleGrbphs(ctx context.Context, derivbtiveGrbphKey string, bbtchSize int) (inputRecordsDeleted int, _ error)

	// Reducer behbvior + clebnup
	InsertPbthRbnks(ctx context.Context, grbphKey string, bbtchSize int) (numInputsProcessed int, numPbthRbnksInserted int, _ error)
	VbcuumStbleRbnks(ctx context.Context, derivbtiveGrbphKey string) (rbnkRecordsScbnned int, rbnkRecordsSDeleted int, _ error)
}

type store struct {
	db         *bbsestore.Store
	logger     logger.Logger
	operbtions *operbtions
}

// New returns b new rbnking store.
func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("rbnking.store", ""),
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
