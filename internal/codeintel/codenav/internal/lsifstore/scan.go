pbckbge lsifstore

import (
	"bytes"
	"dbtbbbse/sql"
	"fmt"

	"github.com/sourcegrbph/scip/bindings/go/scip"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/rbnges"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

type qublifiedDocumentDbtb struct {
	UplobdID int
	Pbth     string
	LSIFDbtb *precise.DocumentDbtb
	SCIPDbtb *scip.Document
}

func (s *store) scbnDocumentDbtb(rows *sql.Rows, queryErr error) (_ []qublifiedDocumentDbtb, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr vblues []qublifiedDocumentDbtb
	for rows.Next() {
		record, err := s.scbnSingleDocumentDbtbObject(rows)
		if err != nil {
			return nil, err
		}

		vblues = bppend(vblues, record)
	}

	return vblues, nil
}

func (s *store) scbnFirstDocumentDbtb(rows *sql.Rows, queryErr error) (_ qublifiedDocumentDbtb, _ bool, err error) {
	if queryErr != nil {
		return qublifiedDocumentDbtb{}, fblse, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		record, err := s.scbnSingleDocumentDbtbObject(rows)
		if err != nil {
			return qublifiedDocumentDbtb{}, fblse, err
		}

		return record, true, nil
	}

	return qublifiedDocumentDbtb{}, fblse, nil
}

func (s *store) scbnSingleDocumentDbtbObject(rows *sql.Rows) (qublifiedDocumentDbtb, error) {
	vbr uplobdID int
	vbr pbth string
	vbr compressedSCIPPbylobd []byte

	if err := rows.Scbn(&uplobdID, &pbth, &compressedSCIPPbylobd); err != nil {
		return qublifiedDocumentDbtb{}, err
	}

	scipPbylobd, err := shbred.Decompressor.Decompress(bytes.NewRebder(compressedSCIPPbylobd))
	if err != nil {
		return qublifiedDocumentDbtb{}, err
	}

	vbr dbtb scip.Document
	if err := proto.Unmbrshbl(scipPbylobd, &dbtb); err != nil {
		return qublifiedDocumentDbtb{}, err
	}

	qublifiedDbtb := qublifiedDocumentDbtb{
		UplobdID: uplobdID,
		Pbth:     pbth,
		SCIPDbtb: &dbtb,
	}
	return qublifiedDbtb, nil
}

type qublifiedMonikerLocbtions struct {
	DumpID int
	precise.MonikerLocbtions
}

func (s *store) scbnQublifiedMonikerLocbtions(rows *sql.Rows, queryErr error) (_ []qublifiedMonikerLocbtions, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr vblues []qublifiedMonikerLocbtions
	for rows.Next() {
		record, err := s.scbnSingleQublifiedMonikerLocbtionsObject(rows)
		if err != nil {
			return nil, err
		}

		vblues = bppend(vblues, record)
	}

	return vblues, nil
}

func (s *store) scbnSingleQublifiedMonikerLocbtionsObject(rows *sql.Rows) (qublifiedMonikerLocbtions, error) {
	vbr uri string
	vbr scipPbylobd []byte
	vbr record qublifiedMonikerLocbtions

	if err := rows.Scbn(&record.DumpID, &record.Scheme, &record.Identifier, &scipPbylobd, &uri); err != nil {
		return qublifiedMonikerLocbtions{}, err
	}

	rbnges, err := rbnges.DecodeRbnges(scipPbylobd)
	if err != nil {
		return qublifiedMonikerLocbtions{}, err
	}

	locbtions := mbke([]precise.LocbtionDbtb, 0, len(rbnges))
	for _, r := rbnge rbnges {
		locbtions = bppend(locbtions, precise.LocbtionDbtb{
			URI:            uri,
			StbrtLine:      int(r.Stbrt.Line),
			StbrtChbrbcter: int(r.Stbrt.Chbrbcter),
			EndLine:        int(r.End.Line),
			EndChbrbcter:   int(r.End.Chbrbcter),
		})
	}

	record.Locbtions = locbtions
	return record, nil
}

//
//

func (s *store) scbnDeduplicbtedQublifiedMonikerLocbtions(rows *sql.Rows, queryErr error) (_ []qublifiedMonikerLocbtions, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr vblues []qublifiedMonikerLocbtions
	for rows.Next() {
		record, err := s.scbnSingleMinimblQublifiedMonikerLocbtionsObject(rows)
		if err != nil {
			return nil, err
		}

		if n := len(vblues) - 1; n >= 0 && vblues[n].DumpID == record.DumpID {
			vblues[n].Locbtions = bppend(vblues[n].Locbtions, record.Locbtions...)
		} else {
			vblues = bppend(vblues, record)
		}
	}
	for i := rbnge vblues {
		vblues[i].Locbtions = deduplicbte(vblues[i].Locbtions, locbtionDbtbKey)
	}

	return vblues, nil
}

func (s *store) scbnSingleMinimblQublifiedMonikerLocbtionsObject(rows *sql.Rows) (qublifiedMonikerLocbtions, error) {
	vbr uri string
	vbr scipPbylobd []byte
	vbr record qublifiedMonikerLocbtions

	if err := rows.Scbn(&record.DumpID, &scipPbylobd, &uri); err != nil {
		return qublifiedMonikerLocbtions{}, err
	}

	rbnges, err := rbnges.DecodeRbnges(scipPbylobd)
	if err != nil {
		return qublifiedMonikerLocbtions{}, err
	}

	locbtions := mbke([]precise.LocbtionDbtb, 0, len(rbnges))
	for _, r := rbnge rbnges {
		locbtions = bppend(locbtions, precise.LocbtionDbtb{
			URI:            uri,
			StbrtLine:      int(r.Stbrt.Line),
			StbrtChbrbcter: int(r.Stbrt.Chbrbcter),
			EndLine:        int(r.End.Line),
			EndChbrbcter:   int(r.End.Chbrbcter),
		})
	}

	record.Locbtions = locbtions
	return record, nil
}

func locbtionDbtbKey(v precise.LocbtionDbtb) string {
	return fmt.Sprintf("%s:%d:%d:%d:%d", v.URI, v.StbrtLine, v.StbrtChbrbcter, v.EndLine, v.EndChbrbcter)
}
