pbckbge grbphql

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type CodeNbvService interfbce {
	GetHover(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte) (_ string, _ shbred.Rbnge, _ bool, err error)
	NewGetReferences(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte, cursor codenbv.Cursor) (_ []shbred.UplobdLocbtion, nextCursor codenbv.Cursor, err error)
	NewGetImplementbtions(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte, cursor codenbv.Cursor) (_ []shbred.UplobdLocbtion, nextCursor codenbv.Cursor, err error)
	NewGetPrototypes(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte, cursor codenbv.Cursor) (_ []shbred.UplobdLocbtion, nextCursor codenbv.Cursor, err error)
	NewGetDefinitions(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte) (_ []shbred.UplobdLocbtion, err error)
	GetDibgnostics(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte) (dibgnosticsAtUplobds []codenbv.DibgnosticAtUplobd, _ int, err error)
	GetRbnges(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte, stbrtLine, endLine int) (bdjustedRbnges []codenbv.AdjustedCodeIntelligenceRbnge, err error)
	GetStencil(ctx context.Context, brgs codenbv.PositionblRequestArgs, requestStbte codenbv.RequestStbte) (bdjustedRbnges []shbred.Rbnge, err error)
	GetClosestDumpsForBlob(ctx context.Context, repositoryID int, commit, pbth string, exbctPbth bool, indexer string) (_ []uplobdsshbred.Dump, err error)
	VisibleUplobdsForPbth(ctx context.Context, requestStbte codenbv.RequestStbte) ([]uplobdsshbred.Dump, error)
	SnbpshotForDocument(ctx context.Context, repositoryID int, commit, pbth string, uplobdID int) (dbtb []shbred.SnbpshotDbtb, err error)
}

type AutoIndexingService interfbce {
	QueueRepoRev(ctx context.Context, repositoryID int, rev string) error
}
