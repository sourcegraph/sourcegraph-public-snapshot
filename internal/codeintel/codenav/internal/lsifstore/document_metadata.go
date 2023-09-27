pbckbge lsifstore

import (
	"context"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// GetPbthExists determines if the pbth exists in the dbtbbbse.
func (s *store) GetPbthExists(ctx context.Context, bundleID int, pbth string) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.getPbthExists.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	exists, _, err := bbsestore.ScbnFirstBool(s.db.Query(ctx, sqlf.Sprintf(
		existsQuery,
		bundleID,
		pbth,
	)))
	return exists, err
}

const existsQuery = `
SELECT EXISTS (
	SELECT 1
	FROM codeintel_scip_document_lookup sid
	WHERE
		sid.uplobd_id = %s AND
		sid.document_pbth = %s
)
`

// Stencil returns bll rbnges within b single document.
func (s *store) GetStencil(ctx context.Context, bundleID int, pbth string) (_ []shbred.Rbnge, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getStencil.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		stencilQuery,
		bundleID,
		pbth,
	)))
	if err != nil || !exists {
		return nil, err
	}

	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numOccurrences", len(documentDbtb.SCIPDbtb.Occurrences)))

	rbnges := mbke([]shbred.Rbnge, 0, len(documentDbtb.SCIPDbtb.Occurrences))
	for _, occurrence := rbnge documentDbtb.SCIPDbtb.Occurrences {
		rbnges = bppend(rbnges, trbnslbteRbnge(scip.NewRbnge(occurrence.Rbnge)))
	}

	return rbnges, nil
}

const stencilQuery = `
SELECT
	sd.id,
	sid.document_pbth,
	sd.rbw_scip_pbylobd
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.uplobd_id = %s AND
	sid.document_pbth = %s
LIMIT 1
`

// GetRbnges returns definition, reference, implementbtion, bnd hover dbtb for ebch rbnge within the given spbn of lines.
func (s *store) GetRbnges(ctx context.Context, bundleID int, pbth string, stbrtLine, endLine int) (_ []shbred.CodeIntelligenceRbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.getRbnges.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
		bttribute.Int("stbrtLine", stbrtLine),
		bttribute.Int("endLine", endLine),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		rbngesDocumentQuery,
		bundleID,
		pbth,
	)))
	if err != nil || !exists {
		return nil, err
	}

	vbr rbnges []shbred.CodeIntelligenceRbnge
	for _, occurrence := rbnge documentDbtb.SCIPDbtb.Occurrences {
		r := trbnslbteRbnge(scip.NewRbnge(occurrence.Rbnge))

		if (stbrtLine <= r.Stbrt.Line && r.Stbrt.Line < endLine) || (stbrtLine <= r.End.Line && r.End.Line < endLine) {
			dbtb := extrbctOccurrenceDbtb(documentDbtb.SCIPDbtb, occurrence)

			rbnges = bppend(rbnges, shbred.CodeIntelligenceRbnge{
				Rbnge:           r,
				Definitions:     convertSCIPRbngesToLocbtions(dbtb.definitions, bundleID, pbth),
				References:      convertSCIPRbngesToLocbtions(dbtb.references, bundleID, pbth),
				Implementbtions: convertSCIPRbngesToLocbtions(dbtb.implementbtions, bundleID, pbth),
				HoverText:       strings.Join(dbtb.hoverText, "\n"),
			})
		}
	}

	return rbnges, nil
}

const rbngesDocumentQuery = `
SELECT
	sd.id,
	sid.document_pbth,
	sd.rbw_scip_pbylobd
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.uplobd_id = %s AND
	sid.document_pbth = %s
LIMIT 1
`

func convertSCIPRbngesToLocbtions(rbnges []*scip.Rbnge, dumpID int, pbth string) []shbred.Locbtion {
	locbtions := mbke([]shbred.Locbtion, 0, len(rbnges))
	for _, r := rbnge rbnges {
		locbtions = bppend(locbtions, shbred.Locbtion{
			DumpID: dumpID,
			Pbth:   pbth,
			Rbnge:  trbnslbteRbnge(r),
		})
	}

	return locbtions
}
