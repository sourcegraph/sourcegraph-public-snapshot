package background

import (
	"context"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/pathexistence"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// correlate reads the content of the given reader as a SCIP index object. The index is processed in
// the background, and processed documents are emitted on a channel to be persisted to the database.
//
// **NOTE TO CONSUMERS OF THIS FUNCTION**:
// As a side-effect of processing documents, a symbol map is built to determine which symbols should
// be advertised as part of our cross-index/cross-repository metadata. Consumers must expect to consume
// the set of processed documents *before* accessing the package or package reference channels - they
// will not be written to until the documents channel has been closed. Consumers should process both
// package and package reference channels concurrently.
func correlate(
	ctx context.Context,
	r io.Reader,
	root string,
	getChildren pathexistence.GetChildrenFunc,
) (lsifstore.ProcessedSCIPData, error) {
	index, err := readIndex(r)
	if err != nil {
		return lsifstore.ProcessedSCIPData{}, err
	}

	ignorePaths, err := ignorePaths(ctx, index.Documents, root, getChildren)
	if err != nil {
		return lsifstore.ProcessedSCIPData{}, err
	}

	var (
		documents             = make(chan lsifstore.ProcessedSCIPDocument)
		packages              = make(chan precise.Package)
		packageReferences     = make(chan precise.PackageReference)
		externalSymbolsByName = readExternalSymbols(index)
	)

	go func() {
		defer close(documents)

		packageSet := map[precise.Package]bool{}
		for _, document := range index.Documents {
			if _, ok := ignorePaths[document.RelativePath]; ok {
				continue
			}

			select {
			case documents <- processDocument(document, externalSymbolsByName):
			case <-ctx.Done():
				return
			}

			// While processing this document, stash the unique packages of each symbol name
			// that is associated with an occurrence. If the occurrence is a definition, mark
			// that package as being one that we define (rather than simply reference).

			for _, occurrence := range document.Occurrences {
				if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
					continue
				}

				if pkg, ok := packageFromSymbol(occurrence.Symbol); ok {
					isDefinition := scip.SymbolRole_Definition.Matches(occurrence)
					packageSet[pkg] = packageSet[pkg] || isDefinition
				}
			}
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
		TextDocumentEncoding: index.Metadata.TextDocumentEncoding.String(),
		ToolName:             index.Metadata.ToolInfo.Name,
		ToolVersion:          index.Metadata.ToolInfo.Version,
		ToolArguments:        index.Metadata.ToolInfo.Arguments,
		ProtocolVersion:      int(index.Metadata.Version),
	}

	return lsifstore.ProcessedSCIPData{
		Metadata:          metadata,
		Documents:         documents,
		Packages:          packages,
		PackageReferences: packageReferences,
	}, nil
}

// readIndex unmarshals a SCIP index from the given reader.
func readIndex(r io.Reader) (*scip.Index, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var index scip.Index
	if err := proto.Unmarshal(content, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// ignorePaths returns a set consisting of the relative paths of documents in the give
// slice that are not resolvable via Git.
func ignorePaths(ctx context.Context, documents []*scip.Document, root string, getChildren pathexistence.GetChildrenFunc) (map[string]struct{}, error) {
	paths := make([]string, 0, len(documents))
	for _, document := range documents {
		paths = append(paths, document.RelativePath)
	}

	checker, err := pathexistence.NewExistenceChecker(ctx, root, paths, getChildren)
	if err != nil {
		return nil, err
	}

	ignorePathMap := map[string]struct{}{}
	for _, document := range documents {
		if !checker.Exists(document.RelativePath) {
			ignorePathMap[document.RelativePath] = struct{}{}
		}
	}

	return ignorePathMap, nil
}

// readExternalSymbols inverts the external symbols from the given index into a map keyed by name.
func readExternalSymbols(index *scip.Index) map[string]*scip.SymbolInformation {
	externalSymbolsByName := make(map[string]*scip.SymbolInformation, len(index.ExternalSymbols))
	for _, symbol := range index.ExternalSymbols {
		externalSymbolsByName[symbol.Symbol] = symbol
	}

	return externalSymbolsByName
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
