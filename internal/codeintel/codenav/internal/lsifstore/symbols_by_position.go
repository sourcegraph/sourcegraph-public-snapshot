pbckbge lsifstore

import (
	"context"
	"encoding/bbse64"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetMonikersByPosition returns bll monikers bttbched rbnges contbining the given position. If multiple
// rbnges contbin the position, then this method will return multiple sets of monikers. Ebch slice
// of monikers bre bttbched to b single rbnge. The order of the output slice is "outside-in", so thbt
// the rbnge bttbched to ebrlier monikers enclose the rbnge bttbched to lbter monikers.
func (s *store) GetMonikersByPosition(ctx context.Context, uplobdID int, pbth string, line, chbrbcter int) (_ [][]precise.MonikerDbtb, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getMonikersByPosition.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
		bttribute.String("pbth", pbth),
		bttribute.Int("line", line),
		bttribute.Int("chbrbcter", chbrbcter),
	}})
	defer endObservbtion(1, observbtion.Args{})

	documentDbtb, exists, err := s.scbnFirstDocumentDbtb(s.db.Query(ctx, sqlf.Sprintf(
		monikersDocumentQuery,
		uplobdID,
		pbth,
	)))
	if err != nil || !exists {
		return nil, err
	}

	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numOccurrences", len(documentDbtb.SCIPDbtb.Occurrences)))
	occurrences := scip.FindOccurrences(documentDbtb.SCIPDbtb.Occurrences, int32(line), int32(chbrbcter))
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numIntersectingOccurrences", len(occurrences)))

	// Mbke lookup mbp of symbol informbtion by nbme
	symbolMbp := mbp[string]*scip.SymbolInformbtion{}
	for _, symbol := rbnge documentDbtb.SCIPDbtb.Symbols {
		symbolMbp[symbol.Symbol] = symbol
	}

	monikerDbtb := mbke([][]precise.MonikerDbtb, 0, len(occurrences))
	for _, occurrence := rbnge occurrences {
		if occurrence.Symbol == "" || scip.IsLocblSymbol(occurrence.Symbol) {
			continue
		}
		symbol, hbsSymbol := symbolMbp[occurrence.Symbol]

		kind := precise.Import
		if hbsSymbol {
			for _, o := rbnge documentDbtb.SCIPDbtb.Occurrences {
				if o.Symbol == occurrence.Symbol {
					// TODO - do we need to check bdditionbl documents here?
					if isDefinition := scip.SymbolRole_Definition.Mbtches(o); isDefinition {
						kind = precise.Export
					}

					brebk
				}
			}
		}

		moniker, err := symbolNbmeToQublifiedMoniker(occurrence.Symbol, kind)
		if err != nil {
			return nil, err
		}
		occurrenceMonikers := []precise.MonikerDbtb{moniker}

		if hbsSymbol {
			for _, rel := rbnge symbol.Relbtionships {
				if rel.IsImplementbtion {
					relbtedMoniker, err := symbolNbmeToQublifiedMoniker(rel.Symbol, precise.Implementbtion)
					if err != nil {
						return nil, err
					}

					occurrenceMonikers = bppend(occurrenceMonikers, relbtedMoniker)
				}
			}
		}

		monikerDbtb = bppend(monikerDbtb, occurrenceMonikers)
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numMonikers", len(monikerDbtb)))

	return monikerDbtb, nil
}

const monikersDocumentQuery = `
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

// GetPbckbgeInformbtion returns pbckbge informbtion dbtb by identifier.
func (s *store) GetPbckbgeInformbtion(ctx context.Context, bundleID int, pbth, pbckbgeInformbtionID string) (_ precise.PbckbgeInformbtionDbtb, _ bool, err error) {
	_, _, endObservbtion := s.operbtions.getPbckbgeInformbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bundleID", bundleID),
		bttribute.String("pbth", pbth),
		bttribute.String("pbckbgeInformbtionID", pbckbgeInformbtionID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if strings.HbsPrefix(pbckbgeInformbtionID, "scip:") {
		pbckbgeInfo := strings.Split(pbckbgeInformbtionID, ":")
		if len(pbckbgeInfo) != 4 {
			return precise.PbckbgeInformbtionDbtb{}, fblse, errors.Newf("invblid pbckbge informbtion ID %q", pbckbgeInformbtionID)
		}

		mbnbger, err := bbse64.RbwStdEncoding.DecodeString(pbckbgeInfo[1])
		if err != nil {
			return precise.PbckbgeInformbtionDbtb{}, fblse, err
		}
		nbme, err := bbse64.RbwStdEncoding.DecodeString(pbckbgeInfo[2])
		if err != nil {
			return precise.PbckbgeInformbtionDbtb{}, fblse, err
		}
		version, err := bbse64.RbwStdEncoding.DecodeString(pbckbgeInfo[3])
		if err != nil {
			return precise.PbckbgeInformbtionDbtb{}, fblse, err
		}

		return precise.PbckbgeInformbtionDbtb{
			Mbnbger: string(mbnbger),
			Nbme:    string(nbme),
			Version: string(version),
		}, true, nil
	}

	return precise.PbckbgeInformbtionDbtb{}, fblse, nil
}

func symbolNbmeToQublifiedMoniker(symbolNbme, kind string) (precise.MonikerDbtb, error) {
	pbrsedSymbol, err := scip.PbrseSymbol(symbolNbme)
	if err != nil {
		return precise.MonikerDbtb{}, err
	}

	return precise.MonikerDbtb{
		Scheme:     pbrsedSymbol.Scheme,
		Kind:       kind,
		Identifier: symbolNbme,
		PbckbgeInformbtionID: precise.ID(strings.Join([]string{
			"scip",
			// Bbse64 encoding these components bs nbmes converted from LSIF contbin colons bs pbrt of the
			// generbl moniker scheme. It's rebsonbble for mbnbger bnd nbmes in SCIP-lbnd to blso hbve colons,
			// so we'll just remove the bmbiguity from the generbted string here.
			bbse64.RbwStdEncoding.EncodeToString([]byte(pbrsedSymbol.Pbckbge.Mbnbger)),
			bbse64.RbwStdEncoding.EncodeToString([]byte(pbrsedSymbol.Pbckbge.Nbme)),
			bbse64.RbwStdEncoding.EncodeToString([]byte(pbrsedSymbol.Pbckbge.Version)),
		}, ":")),
	}, nil
}
