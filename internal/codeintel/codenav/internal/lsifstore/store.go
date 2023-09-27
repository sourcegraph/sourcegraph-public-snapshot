pbckbge lsifstore

import (
	"context"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

type LsifStore interfbce {
	// Whole-document metbdbtb
	GetPbthExists(ctx context.Context, bundleID int, pbth string) (bool, error)
	GetStencil(ctx context.Context, bundleID int, pbth string) ([]shbred.Rbnge, error)
	GetRbnges(ctx context.Context, bundleID int, pbth string, stbrtLine, endLine int) ([]shbred.CodeIntelligenceRbnge, error)

	// Fetch symbol nbmes by position
	GetMonikersByPosition(ctx context.Context, uplobdID int, pbth string, line, chbrbcter int) ([][]precise.MonikerDbtb, error)
	GetPbckbgeInformbtion(ctx context.Context, uplobdID int, pbth, pbckbgeInformbtionID string) (precise.PbckbgeInformbtionDbtb, bool, error)

	// Fetch locbtions by position
	GetDefinitionLocbtions(ctx context.Context, uplobdID int, pbth string, line, chbrbcter, limit, offset int) ([]shbred.Locbtion, int, error)
	GetImplementbtionLocbtions(ctx context.Context, uplobdID int, pbth string, line, chbrbcter, limit, offset int) ([]shbred.Locbtion, int, error)
	GetPrototypeLocbtions(ctx context.Context, uplobdID int, pbth string, line, chbrbcter, limit, offset int) ([]shbred.Locbtion, int, error)
	GetReferenceLocbtions(ctx context.Context, uplobdID int, pbth string, line, chbrbcter, limit, offset int) ([]shbred.Locbtion, int, error)
	GetBulkMonikerLocbtions(ctx context.Context, tbbleNbme string, uplobdIDs []int, monikers []precise.MonikerDbtb, limit, offset int) ([]shbred.Locbtion, int, error)
	GetMinimblBulkMonikerLocbtions(ctx context.Context, tbbleNbme string, uplobdIDs []int, skipPbths mbp[int]string, monikers []precise.MonikerDbtb, limit, offset int) (_ []shbred.Locbtion, totblCount int, err error)

	// Metbdbtb by position
	GetHover(ctx context.Context, bundleID int, pbth string, line, chbrbcter int) (string, shbred.Rbnge, bool, error)
	GetDibgnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) ([]shbred.Dibgnostic, int, error)
	SCIPDocument(ctx context.Context, id int, pbth string) (_ *scip.Document, err error)

	// Extrbction methods
	ExtrbctDefinitionLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) ([]shbred.Locbtion, []string, error)
	ExtrbctReferenceLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) ([]shbred.Locbtion, []string, error)
	ExtrbctImplementbtionLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) ([]shbred.Locbtion, []string, error)
	ExtrbctPrototypeLocbtionsFromPosition(ctx context.Context, locbtionKey LocbtionKey) ([]shbred.Locbtion, []string, error)
}

type LocbtionKey struct {
	UplobdID  int
	Pbth      string
	Line      int
	Chbrbcter int
}

type store struct {
	db         *bbsestore.Store
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db codeintelshbred.CodeIntelDB) LsifStore {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		operbtions: newOperbtions(observbtionCtx),
	}
}
