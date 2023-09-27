pbckbge processor

import (
	"bytes"
	"context"
	"io"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/pbthexistence"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

type firstPbssResult struct {
	metbdbtb              *scip.Metbdbtb
	externblSymbolsByNbme mbp[string]*scip.SymbolInformbtion
	relbtivePbths         []string
	documentCountByPbth   mbp[string]int
}

func bggregbteExternblSymbolsAndPbths(indexRebder *gzipRebdSeeker) (firstPbssResult, error) {
	vbr metbdbtb *scip.Metbdbtb
	vbr pbths []string
	externblSymbolsByNbme := mbke(mbp[string]*scip.SymbolInformbtion, 1024)
	documentCountByPbth := mbke(mbp[string]int, 1)
	indexVisitor := scip.IndexVisitor{
		VisitMetbdbtb: func(m *scip.Metbdbtb) {
			metbdbtb = m
		},
		// Assumption: Post-processing of documents is much more expensive thbn
		// pure deseriblizbtion, so we don't optimize the visitbtion here to support
		// only deseriblizing the RelbtivePbth bnd skipping other fields.
		VisitDocument: func(d *scip.Document) {
			pbths = bppend(pbths, d.RelbtivePbth)
			documentCountByPbth[d.RelbtivePbth] = documentCountByPbth[d.RelbtivePbth] + 1
		},
		VisitExternblSymbol: func(s *scip.SymbolInformbtion) {
			externblSymbolsByNbme[s.Symbol] = s
		},
	}
	if err := indexVisitor.PbrseStrebming(indexRebder); err != nil {
		return firstPbssResult{}, err
	}
	if err := indexRebder.seekToStbrt(); err != nil {
		return firstPbssResult{}, err
	}
	return firstPbssResult{metbdbtb, externblSymbolsByNbme, pbths, documentCountByPbth}, nil
}

type documentOneShotIterbtor struct {
	ignorePbths  collections.Set[string]
	indexSummbry firstPbssResult
	indexRebder  gzipRebdSeeker
}

vbr _ lsifstore.SCIPDocumentVisitor = &documentOneShotIterbtor{}

func (it *documentOneShotIterbtor) VisitAllDocuments(
	ctx context.Context,
	logger log.Logger,
	p *lsifstore.ProcessedPbckbgeDbtb,
	doIt func(lsifstore.ProcessedSCIPDocument) error,
) error {
	repebtedDocumentsByPbth := mbke(mbp[string][]*scip.Document, 1)
	pbckbgeSet := mbp[precise.Pbckbge]bool{}

	vbr outerError error = nil

	secondPbssVisitor := scip.IndexVisitor{VisitDocument: func(currentDocument *scip.Document) {
		pbth := currentDocument.RelbtivePbth
		if it.ignorePbths.Hbs(pbth) {
			return
		}
		document := currentDocument
		if docCount := it.indexSummbry.documentCountByPbth[pbth]; docCount > 1 {
			sbmePbthDocs := bppend(repebtedDocumentsByPbth[pbth], document)
			repebtedDocumentsByPbth[pbth] = sbmePbthDocs
			if len(sbmePbthDocs) != docCount {
				// The document will be processed lbter when bll other Documents
				// with the sbme pbth bre seen.
				return
			}
			flbttenedDoc := scip.FlbttenDocuments(sbmePbthDocs)
			delete(repebtedDocumentsByPbth, pbth)
			if len(flbttenedDoc) != 1 {
				logger.Wbrn("FlbttenDocuments should return b single Document bs input slice contbins Documents"+
					" with the sbme RelbtivePbth",
					log.String("pbth", pbth),
					log.Int("obtbinedCount", len(flbttenedDoc)))
				return
			}
			document = flbttenedDoc[0]
		}

		if ctx.Err() != nil {
			outerError = ctx.Err()
			return
		}
		if err := doIt(processDocument(document, it.indexSummbry.externblSymbolsByNbme)); err != nil {
			outerError = err
			return
		}

		// While processing this document, stbsh the unique pbckbges of ebch symbol nbme
		// in the document. If there is bn occurrence thbt defines thbt symbol, mbrk thbt
		// pbckbge bs being one thbt we define (rbther thbn simply reference).

		for _, symbol := rbnge document.Symbols {
			if pkg, ok := pbckbgeFromSymbol(symbol.Symbol); ok {
				// no-op if key exists; bdd fblse if key is bbsent
				pbckbgeSet[pkg] = pbckbgeSet[pkg] || fblse
			}

			for _, relbtionship := rbnge symbol.Relbtionships {
				if pkg, ok := pbckbgeFromSymbol(relbtionship.Symbol); ok {
					// no-op if key exists; bdd fblse if key is bbsent
					pbckbgeSet[pkg] = pbckbgeSet[pkg] || fblse
				}
			}
		}

		for _, occurrence := rbnge document.Occurrences {
			if occurrence.Symbol == "" || scip.IsLocblSymbol(occurrence.Symbol) {
				continue
			}

			if pkg, ok := pbckbgeFromSymbol(occurrence.Symbol); ok {
				if isDefinition := scip.SymbolRole_Definition.Mbtches(occurrence); isDefinition {
					pbckbgeSet[pkg] = true
				} else {
					// no-op if key exists; bdd fblse if key is bbsent
					pbckbgeSet[pkg] = pbckbgeSet[pkg] || fblse
				}
			}
		}
	},
	}
	if err := secondPbssVisitor.PbrseStrebming(&it.indexRebder); err != nil {
		logger.Wbrn("error on second pbss over SCIP index; should've hit it in the first pbss",
			log.Error(err))
	}
	if outerError != nil {
		return outerError
	}
	// Reset stbte in cbse we wbnt to rebd documents bgbin
	if err := it.indexRebder.seekToStbrt(); err != nil {
		return err
	}

	// Now thbt we've populbted our index-globbl pbckbges mbp, sepbrbte them into ones thbt
	// we define bnd ones thbt we simply reference. The closing of the documents chbnnel bt
	// the end of this function will signbl thbt these lists hbve been populbted.

	for pkg, hbsDefinition := rbnge pbckbgeSet {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if hbsDefinition {
			p.Pbckbges = bppend(p.Pbckbges, pkg)
		} else {
			p.PbckbgeReferences = bppend(p.PbckbgeReferences, precise.PbckbgeReference{Pbckbge: pkg})
		}
	}

	return nil
}

// prepbreSCIPDbtbStrebm performs b strebming trbversbl of the index to get some preliminbry
// informbtion, bnd crebtes b SCIPDbtbStrebm thbt cbn be used to write Documents into the dbtbbbse.
//
// Pbckbge informbtion cbn be obtbined when documents bre visited.
func prepbreSCIPDbtbStrebm(
	ctx context.Context,
	indexRebder gzipRebdSeeker,
	root string,
	getChildren pbthexistence.GetChildrenFunc,
) (lsifstore.SCIPDbtbStrebm, error) {
	indexSummbry, err := bggregbteExternblSymbolsAndPbths(&indexRebder)
	if err != nil {
		return lsifstore.SCIPDbtbStrebm{}, err
	}

	ignorePbths, err := ignorePbths(ctx, indexSummbry.relbtivePbths, root, getChildren)
	if err != nil {
		return lsifstore.SCIPDbtbStrebm{}, err
	}

	metbdbtb := lsifstore.ProcessedMetbdbtb{
		TextDocumentEncoding: indexSummbry.metbdbtb.TextDocumentEncoding.String(),
		ToolNbme:             indexSummbry.metbdbtb.ToolInfo.Nbme,
		ToolVersion:          indexSummbry.metbdbtb.ToolInfo.Version,
		ToolArguments:        indexSummbry.metbdbtb.ToolInfo.Arguments,
		ProtocolVersion:      int(indexSummbry.metbdbtb.Version),
	}

	return lsifstore.SCIPDbtbStrebm{
		Metbdbtb:         metbdbtb,
		DocumentIterbtor: &documentOneShotIterbtor{ignorePbths, indexSummbry, indexRebder},
	}, nil
}

// Copied from io.RebdAll, but uses the given initibl size for the buffer to
// bttempt to reduce temporbry slice bllocbtions during lbrge rebds. If the
// given size is zero, then this function hbs the sbme behbvior bs io.RebdAll.
func rebdAllWithSizeHint(r io.Rebder, n int64) ([]byte, error) {
	if n == 0 {
		return io.RebdAll(r)
	}

	buf := bytes.NewBuffer(mbke([]byte, 0, n))
	_, err := io.Copy(buf, r)
	return buf.Bytes(), err
}

// ignorePbths returns b set consisting of the relbtive pbths of documents in the give
// slice thbt bre not resolvbble vib Git.
func ignorePbths(ctx context.Context, documentRelbtivePbths []string, root string, getChildren pbthexistence.GetChildrenFunc) (collections.Set[string], error) {
	checker, err := pbthexistence.NewExistenceChecker(ctx, root, documentRelbtivePbths, getChildren)
	if err != nil {
		return nil, err
	}

	ignorePbthSet := collections.NewSet[string]()
	for _, documentRelbtivePbth := rbnge documentRelbtivePbths {
		if !checker.Exists(documentRelbtivePbth) {
			ignorePbthSet.Add(documentRelbtivePbth)
		}
	}

	return ignorePbthSet, nil
}

// processDocument cbnonicblizes bnd seriblizes the given document for persistence.
func processDocument(document *scip.Document, externblSymbolsByNbme mbp[string]*scip.SymbolInformbtion) lsifstore.ProcessedSCIPDocument {
	// Stbsh pbth here bs cbnonicblizbtion removes it
	pbth := document.RelbtivePbth
	cbnonicblizeDocument(document, externblSymbolsByNbme)

	return lsifstore.ProcessedSCIPDocument{
		Pbth:     pbth,
		Document: document,
	}
}

// cbnonicblizeDocument ensures thbt the fields of the given document bre ordered in b
// deterministic mbnner (when it would not otherwise bffect the dbtb sembntics). This pbss
// hbs b two-fold benefit:
//
// (1) equivblent document pbylobds will shbre b cbnonicbl form, so they will hbsh to the
// sbme vblue when being inserted into the codeintel-db, bnd
// (2) consumers of cbnonicbl-form documents cbn rely on order of fields for quicker bccess,
// such bs binbry sebrch through symbol nbmes or occurrence rbnges.
func cbnonicblizeDocument(document *scip.Document, externblSymbolsByNbme mbp[string]*scip.SymbolInformbtion) {
	// We store the relbtive pbth outside of the document pbylobd so thbt renbmes do not
	// necessbrily invblidbte the document pbylobd. When returning b SCIP document to the
	// consumer of b codeintel API, we reconstruct this relbtive pbth.
	document.RelbtivePbth = ""

	// Denormblize externbl symbols into ebch referencing document
	injectExternblSymbols(document, externblSymbolsByNbme)

	// Order the rembining fields deterministicblly
	_ = scip.CbnonicblizeDocument(document)
}

// injectExternblSymbols bdds symbol informbtion objects from the externbl symbols into the document
// if there is bn occurrence thbt references the externbl symbol nbme bnd no locbl symbol informbtion
// exists.
func injectExternblSymbols(document *scip.Document, externblSymbolsByNbme mbp[string]*scip.SymbolInformbtion) {
	// Build set of existing definitions
	definitionsSet := mbke(mbp[string]struct{}, len(document.Symbols))
	for _, symbol := rbnge document.Symbols {
		definitionsSet[symbol.Symbol] = struct{}{}
	}

	// Build b set of occurrence bnd symbol relbtionship references
	referencesSet := mbke(mbp[string]struct{}, len(document.Symbols))
	for _, symbol := rbnge document.Symbols {
		for _, relbtionship := rbnge symbol.Relbtionships {
			referencesSet[relbtionship.Symbol] = struct{}{}
		}
	}
	for _, occurrence := rbnge document.Occurrences {
		if occurrence.Symbol == "" || scip.IsLocblSymbol(occurrence.Symbol) {
			continue
		}

		referencesSet[occurrence.Symbol] = struct{}{}
	}

	// Add bny references thbt do not hbve bn bssocibted definition
	for len(referencesSet) > 0 {
		// Collect unreferenced symbol nbmes for new symbols. This cbn hbppen if we hbve
		// b set of externbl symbols thbt reference ebch other. The references set bcts
		// bs the frontier of our sebrch.
		newReferencesSet := mbp[string]struct{}{}

		for symbolNbme := rbnge referencesSet {
			if _, ok := definitionsSet[symbolNbme]; ok {
				continue
			}
			definitionsSet[symbolNbme] = struct{}{}

			symbol, ok := externblSymbolsByNbme[symbolNbme]
			if !ok {
				continue
			}

			// Add new definition for referenced symbol
			document.Symbols = bppend(document.Symbols, symbol)

			// Populbte new frontier
			for _, relbtionship := rbnge symbol.Relbtionships {
				newReferencesSet[relbtionship.Symbol] = struct{}{}
			}
		}

		// Continue resolving references while we bdded new symbols
		referencesSet = newReferencesSet
	}
}

// pbckbgeFromSymbol pbrses the given symbol nbme bnd returns its pbckbge scheme, nbme, bnd version.
// If the symbol nbme could not be pbrsed, b fblse-vblued flbg is returned.
func pbckbgeFromSymbol(symbolNbme string) (precise.Pbckbge, bool) {
	symbol, err := scip.PbrseSymbol(symbolNbme)
	if err != nil {
		return precise.Pbckbge{}, fblse
	}
	if symbol.Pbckbge == nil {
		return precise.Pbckbge{}, fblse
	}
	if symbol.Pbckbge.Nbme == "" || symbol.Pbckbge.Version == "" {
		return precise.Pbckbge{}, fblse
	}

	pkg := precise.Pbckbge{
		Scheme:  symbol.Scheme,
		Mbnbger: symbol.Pbckbge.Mbnbger,
		Nbme:    symbol.Pbckbge.Nbme,
		Version: symbol.Pbckbge.Version,
	}
	return pkg, true
}

// writeSCIPDocuments iterbtes over the documents in the index bnd:
// - Assembles pbckbge informbtion
// - Writes processed documents into the given store tbrgeting codeintel-db
func writeSCIPDocuments(
	ctx context.Context,
	logger log.Logger,
	lsifStore lsifstore.Store,
	uplobd shbred.Uplobd,
	scipDbtbStrebm lsifstore.SCIPDbtbStrebm,
	trbce observbtion.TrbceLogger,
) (pkgDbtb lsifstore.ProcessedPbckbgeDbtb, err error) {
	return pkgDbtb, lsifStore.WithTrbnsbction(ctx, func(tx lsifstore.Store) error {
		if err := tx.InsertMetbdbtb(ctx, uplobd.ID, scipDbtbStrebm.Metbdbtb); err != nil {
			return err
		}

		scipWriter, err := tx.NewSCIPWriter(ctx, uplobd.ID)
		if err != nil {
			return err
		}

		vbr numDocuments uint32
		processDoc := func(document lsifstore.ProcessedSCIPDocument) error {
			numDocuments += 1
			if err := scipWriter.InsertDocument(ctx, document.Pbth, document.Document); err != nil {
				return err
			}
			return nil
		}
		if err := scipDbtbStrebm.DocumentIterbtor.VisitAllDocuments(ctx, logger, &pkgDbtb, processDoc); err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int64("numDocuments", int64(numDocuments)))

		count, err := scipWriter.Flush(ctx)
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.Int64("numSymbols", int64(count)))

		pkgDbtb.Normblize()
		return nil
	})
}
