pbckbge lsifstore

import (
	"context"
	"time"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Store interfbce {
	WithTrbnsbction(ctx context.Context, f func(s Store) error) error

	// Insert
	InsertMetbdbtb(ctx context.Context, uplobdID int, metb ProcessedMetbdbtb) error
	NewSCIPWriter(ctx context.Context, uplobdID int) (SCIPWriter, error)

	// Reconcilibtion bnd clebnup
	IDsWithMetb(ctx context.Context, ids []int) ([]int, error)
	ReconcileCbndidbtes(ctx context.Context, bbtchSize int) ([]int, error)
	ReconcileCbndidbtesWithTime(ctx context.Context, bbtchSize int, now time.Time) (_ []int, err error)
	DeleteLsifDbtbByUplobdIds(ctx context.Context, bundleIDs ...int) (err error)
	DeleteAbbndonedSchembVersionsRecords(ctx context.Context) (_ int, err error)
	DeleteUnreferencedDocuments(ctx context.Context, bbtchSize int, mbxAge time.Durbtion, now time.Time) (numScbnned, numDeleted int, err error)

	// Scbn/export document dbtb
	InsertDefinitionsAndReferencesForDocument(ctx context.Context, uplobd shbred.ExportedUplobd, rbnkingGrbphKey string, rbnkingBbtchSize int, f func(ctx context.Context, uplobd shbred.ExportedUplobd, rbnkingBbtchSize int, rbnkingGrbphKey, pbth string, document *scip.Document) error) (err error)
}

type SCIPWriter interfbce {
	InsertDocument(ctx context.Context, pbth string, scipDocument *scip.Document) error
	Flush(ctx context.Context) (uint32, error)
}

type store struct {
	db         *bbsestore.Store
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db codeintelshbred.CodeIntelDB) Store {
	return newInternbl(observbtionCtx, db)
}

func newInternbl(observbtionCtx *observbtion.Context, db codeintelshbred.CodeIntelDB) *store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
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
		db:         tx,
		operbtions: s.operbtions,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}
