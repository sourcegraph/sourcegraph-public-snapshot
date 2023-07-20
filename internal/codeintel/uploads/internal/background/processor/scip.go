package processor

import (
	"bytes"
	"context"
	"io"
	"sort"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/pathexistence"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type firstPassResult struct {
	metadata              *scip.Metadata
	externalSymbolsByName map[string]*scip.SymbolInformation
	relativePaths         []string
	documentCountByPath   map[string]int
}

func aggregateExternalSymbolsAndPaths(indexReader *gzipReadSeeker) (firstPassResult, error) {
	var metadata *scip.Metadata
	var paths []string
	externalSymbolsByName := make(map[string]*scip.SymbolInformation, 1024)
	documentCountByPath := make(map[string]int, 1)
	indexVisitor := scip.IndexVisitor{
		VisitMetadata: func(m *scip.Metadata) {
			metadata = m
		},
		// Assumption: Post-processing of documents is much more expensive than
		// pure deserialization, so we don't optimize the visitation here to support
		// only deserializing the RelativePath and skipping other fields.
		VisitDocument: func(d *scip.Document) {
			paths = append(paths, d.RelativePath)
			documentCountByPath[d.RelativePath] = documentCountByPath[d.RelativePath] + 1
		},
		VisitExternalSymbol: func(s *scip.SymbolInformation) {
			externalSymbolsByName[s.Symbol] = s
		},
	}
	if err := indexVisitor.ParseStreaming(indexReader); err != nil {
		return firstPassResult{}, err
	}
	if err := indexReader.seekToStart(); err != nil {
		return firstPassResult{}, err
	}
	return firstPassResult{metadata, externalSymbolsByName, paths, documentCountByPath}, nil
}

// correlateSCIP reads the content of the given reader as a SCIP index object. The index is processed in
// the background, and processed documents are emitted on a channel to be persisted to the database.
//
// **NOTE TO CONSUMERS OF THIS FUNCTION** (see `readPackageAndPackageReferences` for a concrete impl):
//
// As a side-effect of processing documents, a symbol map is built to determine which symbols should
// be advertised as part of our cross-index/cross-repository metadata. Consumers must expect to consume
// the set of processed documents *before* accessing the package or package reference channels - they
// will not be written to until the documents channel has been closed. Consumers should process both
// package and package reference channels concurrently.
func correlateSCIP(
	ctx context.Context,
	logger log.Logger,
	indexReader gzipReadSeeker,
	root string,
	getChildren pathexistence.GetChildrenFunc,
) (lsifstore.ProcessedSCIPData, error) {
	indexSummary, err := aggregateExternalSymbolsAndPaths(&indexReader)
	if err != nil {
		return lsifstore.ProcessedSCIPData{}, err
	}

	ignorePaths, err := ignorePaths(ctx, indexSummary.relativePaths, root, getChildren)
	if err != nil {
		return lsifstore.ProcessedSCIPData{}, err
	}

	var (
		documents             = make(chan lsifstore.ProcessedSCIPDocument)
		packages              = make(chan precise.Package)
		packageReferences     = make(chan precise.PackageReference)
		externalSymbolsByName = indexSummary.externalSymbolsByName
	)

	go func() {
		defer close(documents)

		repeatedDocumentsByPath := make(map[string][]*scip.Document, 1)
		packageSet := map[precise.Package]bool{}

		secondPassVisitor := scip.IndexVisitor{VisitDocument: func(currentDocument *scip.Document) {
			path := currentDocument.RelativePath
			if ignorePaths.Has(path) {
				return
			}
			document := currentDocument
			// Most indexers should emit a single SCIP Document for a single file.
			// In the rare case of an indexer bug, if the indexer emits multiple
			// Documents for a single path, we will aggregate those Documents in
			// memory so that we can merge+normalize before processing them further.
			if docCount := indexSummary.documentCountByPath[path]; docCount > 1 {
				samePathDocs := append(repeatedDocumentsByPath[path], document)
				repeatedDocumentsByPath[path] = samePathDocs
				if len(samePathDocs) != docCount {
					// The document will be processed later when all other Documents
					// with the same path are seen.
					return
				}
				flattenedDoc := scip.FlattenDocuments(samePathDocs)
				delete(repeatedDocumentsByPath, path)
				if len(flattenedDoc) != 1 {
					logger.Warn("FlattenDocuments should return a single Document as input slice contains Documents"+
						" with the same RelativePath",
						log.String("path", path),
						log.Int("obtainedCount", len(flattenedDoc)))
					return
				}
				document = flattenedDoc[0]
			}

			select {
			case documents <- processDocument(document, externalSymbolsByName):
			case <-ctx.Done():
				return
			}

			// While processing this document, stash the unique packages of each symbol name
			// in the document. If there is an occurrence that defines that symbol, mark that
			// package as being one that we define (rather than simply reference).

			for _, symbol := range document.Symbols {
				if pkg, ok := packageFromSymbol(symbol.Symbol); ok {
					// no-op if key exists; add false if key is absent
					packageSet[pkg] = packageSet[pkg] || false
				}

				for _, relationship := range symbol.Relationships {
					if pkg, ok := packageFromSymbol(relationship.Symbol); ok {
						// no-op if key exists; add false if key is absent
						packageSet[pkg] = packageSet[pkg] || false
					}
				}
			}

			for _, occurrence := range document.Occurrences {
				if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
					continue
				}

				if pkg, ok := packageFromSymbol(occurrence.Symbol); ok {
					if isDefinition := scip.SymbolRole_Definition.Matches(occurrence); isDefinition {
						packageSet[pkg] = true
					} else {
						// no-op if key exists; add false if key is absent
						packageSet[pkg] = packageSet[pkg] || false
					}
				}
			}
		},
		}
		if err := secondPassVisitor.ParseStreaming(&indexReader); err != nil {
			logger.Warn("error on second pass over SCIP index; should've hit it in the first pass",
				log.Error(err))
		}

		go func() {
			defer close(packages)
			defer close(packageReferences)

			// Now that we've populated our index-global packages map, separate them into ones that
			// we define and ones that we simply reference. The closing of the documents channel at
			// the end of this function will signal that these lists have been populated.

			for pkg, hasDefinition := range packageSet {
				if hasDefinition {
					select {
					case packages <- pkg:
					case <-ctx.Done():
						return
					}
				} else {
					select {
					case packageReferences <- precise.PackageReference{Package: pkg}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}()

	metadata := lsifstore.ProcessedMetadata{
		TextDocumentEncoding: indexSummary.metadata.TextDocumentEncoding.String(),
		ToolName:             indexSummary.metadata.ToolInfo.Name,
		ToolVersion:          indexSummary.metadata.ToolInfo.Version,
		ToolArguments:        indexSummary.metadata.ToolInfo.Arguments,
		ProtocolVersion:      int(indexSummary.metadata.Version),
	}

	return lsifstore.ProcessedSCIPData{
		Metadata:          metadata,
		Documents:         documents,
		Packages:          packages,
		PackageReferences: packageReferences,
	}, nil
}

// readPackageAndPackageReferences reads content from the package and package reference channels of
// the output of `correlateSCIP` and returns them as slices categorized by type. See the implementations
// notes on that function for details.
func readPackageAndPackageReferences(
	ctx context.Context,
	correlatedSCIPData lsifstore.ProcessedSCIPData,
) (packages []precise.Package, packageReferences []precise.PackageReference, _ error) {
	// Perform the following loop while both the package and package reference channels are
	// open. Since the producer of both channels are from the same thread, we have to be able
	// to read from either channel as values are being produced. Once one channel closes, we
	// switch to reading from the still open channel until it is closed as well.

loop:
	for {
		select {
		case pkg, ok := <-correlatedSCIPData.Packages:
			if !ok {
				break loop
			}
			packages = append(packages, pkg)

		case packageReference, ok := <-correlatedSCIPData.PackageReferences:
			if !ok {
				break loop
			}
			packageReferences = append(packageReferences, packageReference)

		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}

	// Drain both channels in case anything is left
	for pkg := range correlatedSCIPData.Packages {
		packages = append(packages, pkg)
	}
	for packageReference := range correlatedSCIPData.PackageReferences {
		packageReferences = append(packageReferences, packageReference)
	}

	// Sort prior to return to get deterministic output
	sort.Slice(packages, func(i, j int) bool {
		return comparePackages(packages[i], packages[j])
	})
	sort.Slice(packageReferences, func(i, j int) bool {
		return comparePackages(packageReferences[i].Package, packageReferences[j].Package)
	})

	return packages, packageReferences, nil
}

// readIndex unmarshals a SCIP index from the given reader. The given reader is in practice
// a gzip deflate layer. We pass the _uncompressed_ size of the reader's payload, which we
// store at upload time, as a hint to the total buffer size we'll be returning. If this value
// is undersized, the standard slice resizing behavior (symmetric to io.ReadAll) is used.
func readIndex(r io.Reader, n int64) (*scip.Index, error) {
	content, err := readAllWithSizeHint(r, n)
	if err != nil {
		return nil, err
	}

	var index scip.Index
	if err := proto.Unmarshal(content, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// Copied from io.ReadAll, but uses the given initial size for the buffer to
// attempt to reduce temporary slice allocations during large reads. If the
// given size is zero, then this function has the same behavior as io.ReadAll.
func readAllWithSizeHint(r io.Reader, n int64) ([]byte, error) {
	if n == 0 {
		return io.ReadAll(r)
	}

	buf := bytes.NewBuffer(make([]byte, 0, n))
	_, err := io.Copy(buf, r)
	return buf.Bytes(), err
}

// ignorePaths returns a set consisting of the relative paths of documents in the give
// slice that are not resolvable via Git.
func ignorePaths(ctx context.Context, documentRelativePaths []string, root string, getChildren pathexistence.GetChildrenFunc) (collections.Set[string], error) {
	checker, err := pathexistence.NewExistenceChecker(ctx, root, documentRelativePaths, getChildren)
	if err != nil {
		return nil, err
	}

	ignorePathSet := collections.NewSet[string]()
	for _, documentRelativePath := range documentRelativePaths {
		if !checker.Exists(documentRelativePath) {
			ignorePathSet.Add(documentRelativePath)
		}
	}

	return ignorePathSet, nil
}

// processDocument canonicalizes and serializes the given document for persistence.
func processDocument(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) lsifstore.ProcessedSCIPDocument {
	// Stash path here as canonicalization removes it
	path := document.RelativePath
	canonicalizeDocument(document, externalSymbolsByName)

	return lsifstore.ProcessedSCIPDocument{
		Path:     path,
		Document: document,
	}
}

// canonicalizeDocument ensures that the fields of the given document are ordered in a
// deterministic manner (when it would not otherwise affect the data semantics). This pass
// has a two-fold benefit:
//
// (1) equivalent document payloads will share a canonical form, so they will hash to the
// same value when being inserted into the codeintel-db, and
// (2) consumers of canonical-form documents can rely on order of fields for quicker access,
// such as binary search through symbol names or occurrence ranges.
func canonicalizeDocument(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) {
	// We store the relative path outside of the document payload so that renames do not
	// necessarily invalidate the document payload. When returning a SCIP document to the
	// consumer of a codeintel API, we reconstruct this relative path.
	document.RelativePath = ""

	// Denormalize external symbols into each referencing document
	injectExternalSymbols(document, externalSymbolsByName)

	// Order the remaining fields deterministically
	_ = scip.CanonicalizeDocument(document)
}

// injectExternalSymbols adds symbol information objects from the external symbols into the document
// if there is an occurrence that references the external symbol name and no local symbol information
// exists.
func injectExternalSymbols(document *scip.Document, externalSymbolsByName map[string]*scip.SymbolInformation) {
	// Build set of existing definitions
	definitionsSet := make(map[string]struct{}, len(document.Symbols))
	for _, symbol := range document.Symbols {
		definitionsSet[symbol.Symbol] = struct{}{}
	}

	// Build a set of occurrence and symbol relationship references
	referencesSet := make(map[string]struct{}, len(document.Symbols))
	for _, symbol := range document.Symbols {
		for _, relationship := range symbol.Relationships {
			referencesSet[relationship.Symbol] = struct{}{}
		}
	}
	for _, occurrence := range document.Occurrences {
		if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
			continue
		}

		referencesSet[occurrence.Symbol] = struct{}{}
	}

	// Add any references that do not have an associated definition
	for len(referencesSet) > 0 {
		// Collect unreferenced symbol names for new symbols. This can happen if we have
		// a set of external symbols that reference each other. The references set acts
		// as the frontier of our search.
		newReferencesSet := map[string]struct{}{}

		for symbolName := range referencesSet {
			if _, ok := definitionsSet[symbolName]; ok {
				continue
			}
			definitionsSet[symbolName] = struct{}{}

			symbol, ok := externalSymbolsByName[symbolName]
			if !ok {
				continue
			}

			// Add new definition for referenced symbol
			document.Symbols = append(document.Symbols, symbol)

			// Populate new frontier
			for _, relationship := range symbol.Relationships {
				newReferencesSet[relationship.Symbol] = struct{}{}
			}
		}

		// Continue resolving references while we added new symbols
		referencesSet = newReferencesSet
	}
}

// packageFromSymbol parses the given symbol name and returns its package scheme, name, and version.
// If the symbol name could not be parsed, a false-valued flag is returned.
func packageFromSymbol(symbolName string) (precise.Package, bool) {
	symbol, err := scip.ParseSymbol(symbolName)
	if err != nil {
		return precise.Package{}, false
	}
	if symbol.Package == nil {
		return precise.Package{}, false
	}

	pkg := precise.Package{
		Scheme:  symbol.Scheme,
		Manager: symbol.Package.Manager,
		Name:    symbol.Package.Name,
		Version: symbol.Package.Version,
	}
	return pkg, true
}

// writeSCIPData transactionally writes the given correlated SCIP data into the given store targeting
// the codeintel-db.
func writeSCIPData(
	ctx context.Context,
	lsifStore lsifstore.Store,
	upload shared.Upload,
	correlatedSCIPData lsifstore.ProcessedSCIPData,
	trace observation.TraceLogger,
) (err error) {
	return lsifStore.WithTransaction(ctx, func(tx lsifstore.Store) error {
		if err := tx.InsertMetadata(ctx, upload.ID, correlatedSCIPData.Metadata); err != nil {
			return err
		}

		scipWriter, err := tx.NewSCIPWriter(ctx, upload.ID)
		if err != nil {
			return err
		}

		var numDocuments uint32
		for document := range correlatedSCIPData.Documents {
			if err := scipWriter.InsertDocument(ctx, document.Path, document.Document); err != nil {
				return err
			}

			numDocuments += 1
		}
		trace.AddEvent("TODO Domain Owner", attribute.Int64("numDocuments", int64(numDocuments)))

		count, err := scipWriter.Flush(ctx)
		if err != nil {
			return err
		}
		trace.AddEvent("TODO Domain Owner", attribute.Int64("numSymbols", int64(count)))

		return nil
	})
}

// comparePackages returns true if pi sorts lower than pj.
func comparePackages(pi, pj precise.Package) bool {
	if pi.Scheme == pj.Scheme {
		if pi.Manager == pj.Manager {
			if pi.Name == pj.Name {
				return pi.Version < pj.Version
			}

			return pi.Name < pj.Name
		}

		return pi.Manager < pj.Manager
	}

	return pi.Scheme < pj.Scheme
}
