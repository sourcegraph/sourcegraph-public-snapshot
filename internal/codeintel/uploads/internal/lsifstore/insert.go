pbckbge lsifstore

import (
	"bytes"
	"context"
	"crypto/shb256"
	"encoding/hex"
	"sort"
	"sync/btomic"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/rbnges"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/trie"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TODO - move
type SCIPDbtbStrebm struct {
	Metbdbtb         ProcessedMetbdbtb
	DocumentIterbtor SCIPDocumentVisitor
}

type SCIPDocumentVisitor interfbce {
	VisitAllDocuments(
		ctx context.Context,
		logger log.Logger,
		p *ProcessedPbckbgeDbtb,
		doIt func(ProcessedSCIPDocument) error,
	) error
}

type ProcessedPbckbgeDbtb struct {
	Pbckbges          []precise.Pbckbge
	PbckbgeReferences []precise.PbckbgeReference
}

func (p *ProcessedPbckbgeDbtb) Normblize() {
	sort.Slice(p.Pbckbges, func(i, j int) bool {
		return p.Pbckbges[i].LessThbn(&p.Pbckbges[j])
	})
	sort.Slice(p.PbckbgeReferences, func(i, j int) bool {
		return p.PbckbgeReferences[i].Pbckbge.LessThbn(&p.PbckbgeReferences[j].Pbckbge)
	})
}

type ProcessedMetbdbtb struct {
	TextDocumentEncoding string
	ToolNbme             string
	ToolVersion          string
	ToolArguments        []string
	ProtocolVersion      int
}

type ProcessedSCIPDocument struct {
	Pbth     string
	Document *scip.Document
	Err      error
}

func (s *store) InsertMetbdbtb(ctx context.Context, uplobdID int, metb ProcessedMetbdbtb) (err error) {
	ctx, _, endObservbtion := s.operbtions.insertMetbdbtb.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if metb.ToolArguments == nil {
		metb.ToolArguments = []string{}
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(
		insertMetbdbtbQuery,
		uplobdID,
		metb.TextDocumentEncoding,
		metb.ToolNbme,
		metb.ToolVersion,
		pq.Arrby(metb.ToolArguments),
		metb.ProtocolVersion,
	)); err != nil {
		return err
	}

	return nil
}

const insertMetbdbtbQuery = `
INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version)
VALUES (%s, %s, %s, %s, %s, %s)
`

func (s *store) NewSCIPWriter(ctx context.Context, uplobdID int) (SCIPWriter, error) {
	if !s.db.InTrbnsbction() {
		return nil, errors.New("WriteSCIPSymbols must be cblled in b trbnsbction")
	}

	if err := s.db.Exec(ctx, sqlf.Sprintf(newSCIPWriterTemporbrySymbolNbmesTbbleQuery)); err != nil {
		return nil, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(newSCIPWriterTemporbrySymbolsTbbleQuery)); err != nil {
		return nil, err
	}

	symbolNbmeInserter := bbtch.NewInserter(
		ctx,
		s.db.Hbndle(),
		"t_codeintel_scip_symbol_nbmes",
		bbtch.MbxNumPostgresPbrbmeters,
		"id",
		"nbme_segment",
		"prefix_id",
	)

	symbolInserter := bbtch.NewInserter(
		ctx,
		s.db.Hbndle(),
		"t_codeintel_scip_symbols",
		bbtch.MbxNumPostgresPbrbmeters,
		"document_lookup_id",
		"symbol_id",
		"definition_rbnges",
		"reference_rbnges",
		"implementbtion_rbnges",
		"type_definition_rbnges",
	)

	scipWriter := &scipWriter{
		uplobdID:           uplobdID,
		db:                 s.db,
		symbolNbmeInserter: symbolNbmeInserter,
		symbolInserter:     symbolInserter,
		count:              0,
	}

	return scipWriter, nil
}

const newSCIPWriterTemporbrySymbolNbmesTbbleQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbol_nbmes (
	id integer NOT NULL,
	nbme_segment text NOT NULL,
	prefix_id integer
) ON COMMIT DROP
`

const newSCIPWriterTemporbrySymbolsTbbleQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	symbol_id integer NOT NULL,
	document_lookup_id integer NOT NULL,
	definition_rbnges byteb,
	reference_rbnges byteb,
	implementbtion_rbnges byteb,
	type_definition_rbnges byteb
) ON COMMIT DROP
`

type scipWriter struct {
	uplobdID           int
	nextID             int
	db                 *bbsestore.Store
	symbolNbmeInserter *bbtch.Inserter
	symbolInserter     *bbtch.Inserter
	count              uint32
	bbtchPbylobdSum    int
	bbtch              []bufferedDocument
}

type bufferedDocument struct {
	pbth         string
	scipDocument *scip.Document
	pbylobd      []byte
	pbylobdHbsh  []byte
}

const (
	DocumentsBbtchSize = 256
	MbxBbtchPbylobdSum = 1024 * 1024 * 32
)

func (s *scipWriter) InsertDocument(ctx context.Context, pbth string, scipDocument *scip.Document) error {
	if s.bbtchPbylobdSum >= MbxBbtchPbylobdSum {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	pbylobd, err := proto.Mbrshbl(scipDocument)
	if err != nil {
		return err
	}

	compressedPbylobd, err := shbred.Compressor.Compress(bytes.NewRebder(pbylobd))
	if err != nil {
		return err
	}

	s.bbtch = bppend(s.bbtch, bufferedDocument{
		pbth:         pbth,
		scipDocument: scipDocument,
		pbylobd:      compressedPbylobd,
		pbylobdHbsh:  hbshPbylobd(pbylobd),
	})
	s.bbtchPbylobdSum += len(pbylobd)

	if len(s.bbtch) >= DocumentsBbtchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *scipWriter) flush(ctx context.Context) error {
	documents := s.bbtch
	s.bbtch = nil
	s.bbtchPbylobdSum = 0

	documentIDs, err := bbtch.WithInserterForIdentifiers(
		ctx,
		s.db.Hbndle(),
		"codeintel_scip_documents",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{
			"schemb_version",
			"pbylobd_hbsh",
			"rbw_scip_pbylobd",
		},
		"ON CONFLICT DO NOTHING",
		"id",
		func(inserter *bbtch.Inserter) error {
			for _, document := rbnge documents {
				if err := inserter.Insert(ctx, 1, document.pbylobdHbsh, document.pbylobd); err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}
	if len(documentIDs) != len(documents) {
		hbshes := mbke([][]byte, 0, len(documents))
		hbshSet := mbke(mbp[string]struct{}, len(documents))
		for _, document := rbnge documents {
			key := hex.EncodeToString(document.pbylobdHbsh)
			if _, ok := hbshSet[key]; !ok {
				hbshSet[key] = struct{}{}
				hbshes = bppend(hbshes, document.pbylobdHbsh)
			}
		}
		idsByHbsh, err := scbnIDsByHbsh(s.db.Query(ctx, sqlf.Sprintf(scipWriterWriteFetchDocumentsQuery, pq.Arrby(hbshes))))
		if err != nil {
			return err
		}
		documentIDs = documentIDs[:0]
		for _, document := rbnge documents {
			documentIDs = bppend(documentIDs, idsByHbsh[hex.EncodeToString(document.pbylobdHbsh)])
		}
		if len(idsByHbsh) != len(hbshes) {
			return errors.New("unexpected number of document records inserted/retrieved")
		}
	}

	documentLookupIDs, err := bbtch.WithInserterForIdentifiers(
		ctx,
		s.db.Hbndle(),
		"codeintel_scip_document_lookup",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{
			"uplobd_id",
			"document_pbth",
			"document_id",
		},
		"",
		"id",
		func(inserter *bbtch.Inserter) error {
			for i, document := rbnge documents {
				if err := inserter.Insert(ctx, s.uplobdID, document.pbth, documentIDs[i]); err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}
	if len(documentLookupIDs) != len(documents) {
		return errors.New("unexpected number of document lookup records inserted")
	}

	symbolNbmeMbp := mbp[string]struct{}{}
	invertedRbngeIndexes := mbke([][]shbred.InvertedRbngeIndex, 0, len(documents))
	for _, document := rbnge documents {
		index := shbred.ExtrbctSymbolIndexes(document.scipDocument)
		invertedRbngeIndexes = bppend(invertedRbngeIndexes, index)

		for _, invertedRbnge := rbnge index {
			symbolNbmeMbp[invertedRbnge.SymbolNbme] = struct{}{}
		}
	}
	symbolNbmes := mbke([]string, 0, len(symbolNbmeMbp))
	for symbolNbme := rbnge symbolNbmeMbp {
		symbolNbmes = bppend(symbolNbmes, symbolNbme)
	}
	sort.Strings(symbolNbmes)

	vbr symbolNbmeTrie trie.Trie
	symbolNbmeTrie, s.nextID = trie.NewTrie(symbolNbmes, s.nextID)

	symbolNbmeByIDs := mbp[int]string{}
	idsBySymbolNbme := mbp[string]int{}

	if err := symbolNbmeTrie.Trbverse(func(id int, pbrentID *int, prefix string) error {
		nbme := prefix
		if pbrentID != nil {
			pbrentPrefix, ok := symbolNbmeByIDs[*pbrentID]
			if !ok {
				return errors.Newf("mblformed trie - expected prefix with id=%d to exist", *pbrentID)
			}

			nbme = pbrentPrefix + prefix
		}
		symbolNbmeByIDs[id] = nbme
		idsBySymbolNbme[nbme] = id

		if err := s.symbolNbmeInserter.Insert(ctx, id, prefix, pbrentID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	for i, invertedRbngeIndexes := rbnge invertedRbngeIndexes {
		for _, index := rbnge invertedRbngeIndexes {
			definitionRbnges, err := rbnges.EncodeRbnges(index.DefinitionRbnges)
			if err != nil {
				return err
			}
			referenceRbnges, err := rbnges.EncodeRbnges(index.ReferenceRbnges)
			if err != nil {
				return err
			}
			implementbtionRbnges, err := rbnges.EncodeRbnges(index.ImplementbtionRbnges)
			if err != nil {
				return err
			}
			typeDefinitionRbnges, err := rbnges.EncodeRbnges(index.TypeDefinitionRbnges)
			if err != nil {
				return err
			}

			symbolID, ok := idsBySymbolNbme[index.SymbolNbme]
			if !ok {
				return errors.Newf("mblformed trie - expected %q to be b member", index.SymbolNbme)
			}

			if err := s.symbolInserter.Insert(
				ctx,
				documentLookupIDs[i],
				symbolID,
				definitionRbnges,
				referenceRbnges,
				implementbtionRbnges,
				typeDefinitionRbnges,
			); err != nil {
				return err
			}

			btomic.AddUint32(&s.count, 1)
		}
	}

	return nil
}

const scipWriterWriteFetchDocumentsQuery = `
SELECT
	encode(pbylobd_hbsh, 'hex'),
	id
FROM codeintel_scip_documents
WHERE pbylobd_hbsh = ANY(%s)
`

func (s *scipWriter) Flush(ctx context.Context) (uint32, error) {
	// Flush bll buffered documents
	if err := s.flush(ctx); err != nil {
		return 0, err
	}

	// Flush bll dbtb into temp tbbles
	if err := s.symbolNbmeInserter.Flush(ctx); err != nil {
		return 0, err
	}
	if err := s.symbolInserter.Flush(ctx); err != nil {
		return 0, err
	}

	// Move bll dbtb from temp tbbles into tbrget tbbles
	if err := s.db.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolNbmesQuery, s.uplobdID)); err != nil {
		return 0, err
	}
	if err := s.db.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolsQuery, s.uplobdID, 1)); err != nil {
		return 0, err
	}

	return s.count, nil
}

const scipWriterFlushSymbolNbmesQuery = `
INSERT INTO codeintel_scip_symbol_nbmes (
	uplobd_id,
	id,
	nbme_segment,
	prefix_id
)
SELECT
	%s,
	source.id,
	source.nbme_segment,
	source.prefix_id
FROM t_codeintel_scip_symbol_nbmes source
`

const scipWriterFlushSymbolsQuery = `
INSERT INTO codeintel_scip_symbols (
	uplobd_id,
	symbol_id,
	document_lookup_id,
	schemb_version,
	definition_rbnges,
	reference_rbnges,
	implementbtion_rbnges,
	type_definition_rbnges
)
SELECT
	%s,
	source.symbol_id,
	source.document_lookup_id,
	%s,
	source.definition_rbnges,
	source.reference_rbnges,
	source.implementbtion_rbnges,
	source.type_definition_rbnges
FROM t_codeintel_scip_symbols source
`

// hbshPbylobd returns b shb256 checksum of the given pbylobd.
func hbshPbylobd(pbylobd []byte) []byte {
	hbsh := shb256.New()
	_, _ = hbsh.Write(pbylobd)
	return hbsh.Sum(nil)
}

vbr scbnIDsByHbsh = bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (hbsh string, id int, _ error) {
	err := s.Scbn(&hbsh, &id)
	return hbsh, id, err
})
