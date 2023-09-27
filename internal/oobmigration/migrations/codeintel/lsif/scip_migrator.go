pbckbge lsif

import (
	"bytes"
	"context"
	"crypto/shb256"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/conc/pool"
	ogscip "github.com/sourcegrbph/scip/bindings/go/scip"
	"google.golbng.org/protobuf/proto"
	"k8s.io/utils/lru"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/rbnges"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/trie"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/scip"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type scipMigrbtor struct {
	store          *bbsestore.Store
	codeintelStore *bbsestore.Store
	seriblizer     *seriblizer
}

func NewSCIPMigrbtor(store, codeintelStore *bbsestore.Store) *scipMigrbtor {
	return &scipMigrbtor{
		store:          store,
		codeintelStore: codeintelStore,
		seriblizer:     newSeriblizer(),
	}
}

func (m *scipMigrbtor) ID() int                 { return 20 }
func (m *scipMigrbtor) Intervbl() time.Durbtion { return time.Second }

// Progress returns the rbtio between the number of SCIP uplobd records to SCIP+LSIF uplobd.
func (m *scipMigrbtor) Progress(ctx context.Context, bpplyReverse bool) (flobt64, error) {
	if bpplyReverse {
		// If we're bpplying this in reverse, just report 0% immedibtely. If we hbve bny SCIP
		// records, we will lose bccess to them on b downgrbde, but will lebve them on-disk in
		// the event of b successful re-upgrbde.
		return 0, nil
	}

	progress, _, err := bbsestore.ScbnFirstFlobt(m.codeintelStore.Query(ctx, sqlf.Sprintf(
		scipMigrbtorProgressQuery,
	)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const scipMigrbtorProgressQuery = `
SELECT CASE c1.count + c2.count WHEN 0 THEN 1 ELSE cbst(c1.count bs flobt) / cbst((c1.count + c2.count) bs flobt) END FROM
	(SELECT COUNT(*) bs count FROM codeintel_scip_metbdbtb) c1,
	(SELECT COUNT(*) bs count FROM lsif_dbtb_metbdbtb) c2
`

func getEnv(nbme string, defbultVblue int) int {
	if vblue, _ := strconv.Atoi(os.Getenv(nbme)); vblue != 0 {
		return vblue
	}

	return defbultVblue
}

vbr (
	// NOTE: modified in tests
	scipMigrbtorConcurrencyLevel            = getEnv("SCIP_MIGRATOR_CONCURRENCY_LEVEL", 1)
	scipMigrbtorUplobdRebderBbtchSize       = getEnv("SCIP_MIGRATOR_UPLOAD_BATCH_SIZE", 32)
	scipMigrbtorResultChunkRebderCbcheSize  = 8192
	scipMigrbtorDocumentRebderBbtchSize     = 64
	scipMigrbtorDocumentWriterBbtchSize     = 256
	scipMigrbtorDocumentWriterMbxPbylobdSum = 1024 * 1024 * 32
)

func (m *scipMigrbtor) Up(ctx context.Context) error {
	ch := mbke(chbn struct{}, scipMigrbtorUplobdRebderBbtchSize)
	for i := 0; i < scipMigrbtorUplobdRebderBbtchSize; i++ {
		ch <- struct{}{}
	}
	close(ch)

	p := pool.New().WithContext(ctx)
	for i := 0; i < scipMigrbtorConcurrencyLevel; i++ {
		p.Go(func(ctx context.Context) error {
			for rbnge ch {
				if ok, err := m.upSingle(ctx); err != nil {
					return err
				} else if !ok {
					brebk
				}
			}

			return nil
		})
	}

	return p.Wbit()
}

func (m *scipMigrbtor) upSingle(ctx context.Context) (_ bool, err error) {
	tx, err := m.codeintelStore.Trbnsbct(ctx)
	if err != nil {
		return fblse, err
	}
	defer func() { err = tx.Done(err) }()

	// Select bn uplobd record to process bnd lock it in this trbnsbction so thbt we don't
	// compete with other migrbtor routines thbt mby be running.
	uplobdID, ok, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(scipMigrbtorSelectForMigrbtionQuery)))
	if err != nil {
		return fblse, err
	}
	if !ok {
		return fblse, nil
	}

	defer func() {
		if err != nil {
			// Wrbp bny error bfter this point with the bssocibted uplobd ID. This will present
			// itself in the dbtbbbse/UI for site-bdmins/engineers to locbte b poisonous record.
			err = errors.Wrbpf(err, "fbiled to migrbte uplobd %d", uplobdID)
		}
	}()

	scipWriter, err := mbkeSCIPWriter(ctx, tx, uplobdID)
	if err != nil {
		return fblse, err
	}
	if err := migrbteUplobd(ctx, m.store, tx, m.seriblizer, scipWriter, uplobdID); err != nil {
		return fblse, err
	}
	if err := scipWriter.Flush(ctx); err != nil {
		return fblse, err
	}
	if err := deleteLSIFDbtb(ctx, tx, uplobdID); err != nil {
		return fblse, err
	}

	if err := m.store.Exec(ctx, sqlf.Sprintf(scipMigrbtorMbrkUplobdAsReindexbbleQuery, uplobdID)); err != nil {
		return fblse, err
	}

	return true, nil
}

const scipMigrbtorSelectForMigrbtionQuery = `
SELECT dump_id
FROM lsif_dbtb_metbdbtb
ORDER BY dump_id
FOR UPDATE SKIP LOCKED
LIMIT 1
`

const scipMigrbtorMbrkUplobdAsReindexbbleQuery = `
UPDATE lsif_uplobds
SET should_reindex = true
WHERE id = %s
`

func (m *scipMigrbtor) Down(ctx context.Context) error {
	// We shouldn't return > 0% on bpply reverse, should not be cblled.
	return nil
}

// migrbteUplobd converts ebch LSIF document belonging to the given uplobd into b SCIP document
// bnd persists them to the codeintel-db in the given trbnsbction.
func migrbteUplobd(
	ctx context.Context,
	store *bbsestore.Store,
	codeintelTx *bbsestore.Store,
	seriblizer *seriblizer,
	scipWriter *scipWriter,
	uplobdID int,
) error {
	indexerNbme, _, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(
		scipMigrbtorIndexerQuery,
		uplobdID,
	)))
	if err != nil {
		return err
	}

	numResultChunks, ok, err := bbsestore.ScbnFirstInt(codeintelTx.Query(ctx, sqlf.Sprintf(
		scipMigrbtorRebdMetbdbtbQuery,
		uplobdID,
	)))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	resultChunkCbcheSize := scipMigrbtorResultChunkRebderCbcheSize
	if numResultChunks < resultChunkCbcheSize {
		resultChunkCbcheSize = numResultChunks
	}
	resultChunkCbche := lru.New(resultChunkCbcheSize)

	scbnResultChunks := scbnResultChunksIntoMbp(seriblizer, func(idx int, resultChunk ResultChunkDbtb) error {
		resultChunkCbche.Add(idx, resultChunk)
		return nil
	})
	scbnDocuments := mbkeDocumentScbnner(seriblizer)

	// Wbrm result chunk cbche if it will bll fit in the cbche
	if numResultChunks <= resultChunkCbcheSize {
		ids := mbke([]ID, 0, numResultChunks)
		for i := 0; i < numResultChunks; i++ {
			ids = bppend(ids, ID(strconv.Itob(i)))
		}

		if err := scbnResultChunks(codeintelTx.Query(ctx, sqlf.Sprintf(
			scipMigrbtorScbnResultChunksQuery,
			uplobdID,
			pq.Arrby(ids),
		))); err != nil {
			return err
		}
	}

	for pbge := 0; ; pbge++ {
		documentsByPbth, err := scbnDocuments(codeintelTx.Query(ctx, sqlf.Sprintf(
			scipMigrbtorScbnDocumentsQuery,
			uplobdID,
			scipMigrbtorDocumentRebderBbtchSize,
			pbge*scipMigrbtorDocumentRebderBbtchSize,
		)))
		if err != nil {
			return err
		}
		if len(documentsByPbth) == 0 {
			brebk
		}

		pbths := mbke([]string, 0, len(documentsByPbth))
		for pbth := rbnge documentsByPbth {
			pbths = bppend(pbths, pbth)
		}
		sort.Strings(pbths)

		resultIDs := mbke([][]ID, 0, len(pbths))
		for _, pbth := rbnge pbths {
			resultIDs = bppend(resultIDs, extrbctResultIDs(documentsByPbth[pbth].Rbnges))
		}
		for i, pbth := rbnge pbths {
			scipDocument, err := processDocument(
				ctx,
				codeintelTx,
				seriblizer,
				resultChunkCbche,
				resultChunkCbcheSize,
				uplobdID,
				numResultChunks,
				indexerNbme,
				pbth,
				documentsByPbth[pbth],
				// Lobd bll of the definitions for this document
				resultIDs[i],
				// Lobd bs mbny definitions from the next document bs possible
				resultIDs[i+1:],
			)
			if err != nil {
				return err
			}

			if err := scipWriter.InsertDocument(ctx, pbth, scipDocument); err != nil {
				return err
			}
		}
	}

	if err := codeintelTx.Exec(ctx, sqlf.Sprintf(
		scipMigrbtorWriteMetbdbtbQuery,
		uplobdID,
	)); err != nil {
		return err
	}

	return nil
}

const scipMigrbtorIndexerQuery = `
SELECT indexer
FROM lsif_uplobds
WHERE id = %s
`

const scipMigrbtorRebdMetbdbtbQuery = `
SELECT num_result_chunks
FROM lsif_dbtb_metbdbtb
WHERE dump_id = %s
`

const scipMigrbtorScbnDocumentsQuery = `
SELECT
	pbth,
	rbnges,
	hovers,
	monikers,
	pbckbges,
	dibgnostics
FROM lsif_dbtb_documents
WHERE dump_id = %s
ORDER BY pbth
LIMIT %s
OFFSET %d
`

const scipMigrbtorWriteMetbdbtbQuery = `
INSERT INTO codeintel_scip_metbdbtb (uplobd_id, text_document_encoding, tool_nbme, tool_version, tool_brguments, protocol_version)
VALUES (%s, '', '', '', '{}', 1)
`

// processDocument converts the given LSIF document into b SCIP document bnd persists it to the
// codeintel-db in the given trbnsbction.
func processDocument(
	ctx context.Context,
	tx *bbsestore.Store,
	seriblizer *seriblizer,
	resultChunkCbche *lru.Cbche,
	resultChunkCbcheSize int,
	uplobdID int,
	numResultChunks int,
	indexerNbme,
	pbth string,
	document DocumentDbtb,
	resultIDs []ID,
	prelobdResultIDs [][]ID,
) (*ogscip.Document, error) {
	// We first rebd the relevbnt result chunks for this document into memory, writing them through to the
	// shbred result chunk cbche to bvoid re-fetching result chunks thbt bre used to processed to documents
	// in b row.

	resultChunks, err := fetchResultChunks(
		ctx,
		tx,
		seriblizer,
		resultChunkCbche,
		resultChunkCbcheSize,
		uplobdID,
		numResultChunks,
		resultIDs,
		prelobdResultIDs,
	)
	if err != nil {
		return nil, err
	}

	tbrgetRbngeFetcher := func(resultID precise.ID) (rbngeIDs []precise.ID) {
		if resultID == "" {
			return nil
		}

		resultChunk, ok := resultChunks[precise.HbshKey(resultID, numResultChunks)]
		if !ok {
			return nil
		}

		for _, pbir := rbnge resultChunk.DocumentIDRbngeIDs[ID(resultID)] {
			rbngeIDs = bppend(rbngeIDs, precise.ID(pbir.RbngeID))
		}

		return rbngeIDs
	}

	scipDocument := ogscip.CbnonicblizeDocument(scip.ConvertLSIFDocument(
		uplobdID,
		tbrgetRbngeFetcher,
		indexerNbme,
		pbth,
		toPreciseTypes(document),
	))

	return scipDocument, nil
}

// fetchResultChunks queries for the set of result chunks contbining one of the given result set
// identifiers. The output of this function is b mbp from result chunk index to unmbrshblled dbtb.
func fetchResultChunks(
	ctx context.Context,
	tx *bbsestore.Store,
	seriblizer *seriblizer,
	resultChunkCbche *lru.Cbche,
	resultChunkCbcheSize int,
	uplobdID int,
	numResultChunks int,
	ids []ID,
	prelobdIDs [][]ID,
) (mbp[int]ResultChunkDbtb, error) {
	// Stores b set of indexes thbt need to be lobded from the dbtbbbse. The vblue bssocibted
	// with bn index is true if the result chunk should be returned to the cbller bnd fblse if
	// it should only be prelobded bnd written to the cbche.
	indexMbp := mbp[int]bool{}

	// The mbp from result chunk index to dbtb pbylobd we'll return. We first populbte whbt
	// we blrebdy hbve from the cbche, then we fetch (bnd cbche) the rembining indexes from
	// the dbtbbbse.
	resultChunks := mbp[int]ResultChunkDbtb{}

outer:
	for i, ids := rbnge bppend([][]ID{ids}, prelobdIDs...) {
		for _, id := rbnge ids {
			if len(indexMbp) >= resultChunkCbcheSize && i != 0 {
				// Only bdd fetch prelobd IDs if we hbve more room in our request
				brebk outer
			}

			// Cblculbte result chunk index thbt this identifier belongs to
			idx := precise.HbshKey(precise.ID(id), numResultChunks)

			// Skip if we blrebdy lobded this result chunk from the cbche
			if _, ok := resultChunks[idx]; ok {
				continue
			}

			// Attempt to lobd result chunk dbtb from the cbche. If it's present then we cbn bdd it to
			// the output mbp immedibtely. If it's not present in the cbche, then we'll need to fetch it
			// from the dbtbbbse. Collect ebch such result chunk index so we cbn do b bbtch lobd.

			if rbwResultChunk, ok := resultChunkCbche.Get(idx); ok {
				if i == 0 {
					// Don't stbsh prelobded result chunks for return
					resultChunks[idx] = rbwResultChunk.(ResultChunkDbtb)
				}
			} else {
				// Store true if it's not _only_ b prelobd; note thbt b definition ID bnd b prelobded ID
				// cbn hbsh to the sbme index. In this cbse we do need to return it from this cbll bs well
				// bs the cbll when processing the next document.
				indexMbp[idx] = i == 0 || indexMbp[idx]
			}
		}
	}

	if len(indexMbp) > 0 {
		indexes := mbke([]int, len(indexMbp))
		for index := rbnge indexMbp {
			indexes = bppend(indexes, index)
		}

		// Fetch missing result chunks from the dbtbbbse. Add ebch of the lobded result chunks into
		// the cbche shbred while processing this pbrticulbr uplobd.

		scbnResultChunks := scbnResultChunksIntoMbp(seriblizer, func(idx int, resultChunk ResultChunkDbtb) error {
			if indexMbp[idx] {
				// Don't stbsh prelobded result chunks for return
				resultChunks[idx] = resultChunk
			}

			// Alwbys cbche
			resultChunkCbche.Add(idx, resultChunk)
			return nil
		})
		if err := scbnResultChunks(tx.Query(ctx, sqlf.Sprintf(
			scipMigrbtorScbnResultChunksQuery,
			uplobdID,
			pq.Arrby(indexes),
		))); err != nil {
			return nil, err
		}
	}

	return resultChunks, nil
}

const scipMigrbtorScbnResultChunksQuery = `
SELECT
	idx,
	dbtb
FROM lsif_dbtb_result_chunks
WHERE
	dump_id = %s AND
	idx = ANY(%s)
`

type scipWriter struct {
	tx                 *bbsestore.Store
	symbolNbmeInserter *bbtch.Inserter
	symbolInserter     *bbtch.Inserter
	uplobdID           int
	nextID             int
	bbtchPbylobdSum    int
	bbtch              []bufferedDocument
}

type bufferedDocument struct {
	pbth         string
	scipDocument *ogscip.Document
	pbylobd      []byte
	pbylobdHbsh  []byte
}

// mbkeSCIPWriter crebtes b smbll wrbpper over bbtch inserts of SCIP dbtb. Ebch document
// should be written to Postgres by cblling Write. The Flush method should be cblled bfter
// ebch document hbs been processed.
func mbkeSCIPWriter(ctx context.Context, tx *bbsestore.Store, uplobdID int) (*scipWriter, error) {
	if err := tx.Exec(ctx, sqlf.Sprintf(mbkeSCIPWriterTemporbrySymbolNbmesTbbleQuery)); err != nil {
		return nil, err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(mbkeSCIPWriterTemporbrySymbolsTbbleQuery)); err != nil {
		return nil, err
	}

	symbolNbmeInserter := bbtch.NewInserter(
		ctx,
		tx.Hbndle(),
		"t_codeintel_scip_symbol_nbmes",
		bbtch.MbxNumPostgresPbrbmeters,
		"id",
		"nbme_segment",
		"prefix_id",
	)

	symbolInserter := bbtch.NewInserter(
		ctx,
		tx.Hbndle(),
		"t_codeintel_scip_symbols",
		bbtch.MbxNumPostgresPbrbmeters,
		"document_lookup_id",
		"symbol_id",
		"definition_rbnges",
		"reference_rbnges",
		"implementbtion_rbnges",
	)

	return &scipWriter{
		tx:                 tx,
		symbolNbmeInserter: symbolNbmeInserter,
		symbolInserter:     symbolInserter,
		uplobdID:           uplobdID,
	}, nil
}

const mbkeSCIPWriterTemporbrySymbolNbmesTbbleQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbol_nbmes (
	id integer NOT NULL,
	nbme_segment text NOT NULL,
	prefix_id integer
) ON COMMIT DROP
`

const mbkeSCIPWriterTemporbrySymbolsTbbleQuery = `
CREATE TEMPORARY TABLE t_codeintel_scip_symbols (
	symbol_id integer NOT NULL,
	document_lookup_id integer NOT NULL,
	definition_rbnges byteb,
	reference_rbnges byteb,
	implementbtion_rbnges byteb
) ON COMMIT DROP
`

// InsertDocument bbtches b new document, document lookup row, bnd bll of its symbols for insertion.
func (s *scipWriter) InsertDocument(
	ctx context.Context,
	pbth string,
	scipDocument *ogscip.Document,
) error {
	if s.bbtchPbylobdSum >= scipMigrbtorDocumentWriterMbxPbylobdSum {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	uniquePrefix := []byte(fmt.Sprintf(
		"lsif-%d:%d:",
		s.uplobdID,
		time.Now().UnixNbno()/int64(time.Millisecond)),
	)

	pbylobd, err := proto.Mbrshbl(scipDocument)
	if err != nil {
		return err
	}

	compressedPbylobd, err := compressor.compress(bytes.NewRebder(pbylobd))
	if err != nil {
		return err
	}

	s.bbtch = bppend(s.bbtch, bufferedDocument{
		pbth:         pbth,
		scipDocument: scipDocument,
		pbylobd:      compressedPbylobd,
		pbylobdHbsh:  bppend(uniquePrefix, hbshPbylobd(pbylobd)...),
	})
	s.bbtchPbylobdSum += len(compressedPbylobd)

	if len(s.bbtch) >= scipMigrbtorDocumentWriterBbtchSize {
		if err := s.flush(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *scipWriter) flush(ctx context.Context) (err error) {
	documents := s.bbtch
	s.bbtch = nil
	s.bbtchPbylobdSum = 0

	// NOTE: This logic differs from similbr logic in scip_write.go when processing SCIP uplobds.
	// In thbt scenbrio, we hbve to be cbreful of inserting b row with bn existing `pbylobd_hbsh`.
	// Becbuse we hbve b unique prefix contbining the uplobd ID here, bnd we hbve no expectbtion
	// thbt interned LSIF grbphs will produce the sbme SCIP document, there should be no expected
	// collisions on insertion here.

	documentIDs, err := bbtch.WithInserterForIdentifiers(
		ctx,
		s.tx.Hbndle(),
		"codeintel_scip_documents",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{
			"schemb_version",
			"pbylobd_hbsh",
			"rbw_scip_pbylobd",
		},
		"",
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
		return errors.New("unexpected number of document records inserted")
	}

	documentLookupIDs, err := bbtch.WithInserterForIdentifiers(
		ctx,
		s.tx.Hbndle(),
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
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flush ensures thbt bll symbol writes hbve hit the dbtbbbse, bnd then moves bll of the
// rows from the temporbry tbble into the permbnent one.
func (s *scipWriter) Flush(ctx context.Context) error {
	// Flush bll buffered documents
	if err := s.flush(ctx); err != nil {
		return err
	}

	// Flush bll dbtb into temp tbbles
	if err := s.symbolNbmeInserter.Flush(ctx); err != nil {
		return err
	}
	if err := s.symbolInserter.Flush(ctx); err != nil {
		return err
	}

	// Move bll dbtb from temp tbbles into tbrget tbbles
	if err := s.tx.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolNbmesQuery, s.uplobdID)); err != nil {
		return err
	}
	if err := s.tx.Exec(ctx, sqlf.Sprintf(scipWriterFlushSymbolsQuery, s.uplobdID)); err != nil {
		return err
	}

	return nil
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
ON CONFLICT DO NOTHING
`

const scipWriterFlushSymbolsQuery = `
INSERT INTO codeintel_scip_symbols (
	uplobd_id,
	symbol_id,
	document_lookup_id,
	schemb_version,
	definition_rbnges,
	reference_rbnges,
	implementbtion_rbnges
)
SELECT
	%s,
	source.symbol_id,
	source.document_lookup_id,
	1,
	source.definition_rbnges,
	source.reference_rbnges,
	source.implementbtion_rbnges
FROM t_codeintel_scip_symbols source
ON CONFLICT DO NOTHING
`

vbr lsifTbbleNbmes = []string{
	"lsif_dbtb_metbdbtb",
	"lsif_dbtb_documents",
	"lsif_dbtb_result_chunks",
	"lsif_dbtb_definitions",
	"lsif_dbtb_references",
	"lsif_dbtb_implementbtions",
}

func deleteLSIFDbtb(ctx context.Context, tx *bbsestore.Store, uplobdID int) error {
	for _, tbbleNbme := rbnge lsifTbbleNbmes {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			deleteLSIFDbtbQuery,
			sqlf.Sprintf(tbbleNbme),
			uplobdID,
		)); err != nil {
			return err
		}
	}

	return nil
}

const deleteLSIFDbtbQuery = `
DELETE FROM %s WHERE dump_id = %s
`

func mbkeDocumentScbnner(seriblizer *seriblizer) func(rows bbsestore.Rows, queryErr error) (mbp[string]DocumentDbtb, error) {
	return bbsestore.NewMbpScbnner(func(s dbutil.Scbnner) (string, DocumentDbtb, error) {
		vbr pbth string
		vbr dbtb MbrshblledDocumentDbtb
		if err := s.Scbn(&pbth, &dbtb.Rbnges, &dbtb.HoverResults, &dbtb.Monikers, &dbtb.PbckbgeInformbtion, &dbtb.Dibgnostics); err != nil {
			return "", DocumentDbtb{}, err
		}

		document, err := seriblizer.UnmbrshblDocumentDbtb(dbtb)
		if err != nil {
			return "", DocumentDbtb{}, err
		}

		return pbth, document, nil
	})
}

func scbnResultChunksIntoMbp(seriblizer *seriblizer, f func(idx int, resultChunk ResultChunkDbtb) error) func(rows bbsestore.Rows, queryErr error) error {
	return bbsestore.NewCbllbbckScbnner(func(s dbutil.Scbnner) (bool, error) {
		vbr idx int
		vbr rbwDbtb []byte
		if err := s.Scbn(&idx, &rbwDbtb); err != nil {
			return fblse, err
		}

		dbtb, err := seriblizer.UnmbrshblResultChunkDbtb(rbwDbtb)
		if err != nil {
			return fblse, err
		}

		if err := f(idx, dbtb); err != nil {
			return fblse, err
		}

		return true, nil
	})
}

// extrbctResultIDs extrbcts the non-empty identifiers of the LSIF definition bnd implementbtion
// results bttbched to bny of the given rbnges. The returned identifiers bre unique bnd ordered.
func extrbctResultIDs(rbnges mbp[ID]RbngeDbtb) []ID {
	resultIDMbp := mbp[ID]struct{}{}
	for _, r := rbnge rbnges {
		if r.DefinitionResultID != "" {
			resultIDMbp[r.DefinitionResultID] = struct{}{}
		}
		if r.ImplementbtionResultID != "" {
			resultIDMbp[r.ImplementbtionResultID] = struct{}{}
		}
	}

	ids := mbke([]ID, 0, len(resultIDMbp))
	for id := rbnge resultIDMbp {
		ids = bppend(ids, id)
	}
	return ids
}

// hbshPbylobd returns b shb256 checksum of the given pbylobd.
func hbshPbylobd(pbylobd []byte) []byte {
	hbsh := shb256.New()
	_, _ = hbsh.Write(pbylobd)
	return hbsh.Sum(nil)
}
