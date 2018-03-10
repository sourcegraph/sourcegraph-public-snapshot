package graphqlbackend

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/neelance/parallel"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

var mockSearchSymbols func(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int) (res []*symbolResolver, common *searchResultsCommon, err error)

// searchSymbols searches the given repos in parallel for symbols matching the given search query
// it can be used for both search suggestions and search results
//
// May return partial results and an error
func searchSymbols(ctx context.Context, args *repoSearchArgs, query searchquery.Query, limit int) (res []*symbolResolver, common *searchResultsCommon, err error) {
	if mockSearchSymbols != nil {
		return mockSearchSymbols(ctx, args, query, limit)
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Search symbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	if args.query.Pattern == "" {
		return nil, nil, nil
	}

	ctx, cancelAll := context.WithCancel(ctx)
	defer cancelAll()

	common = &searchResultsCommon{}
	var (
		run = parallel.NewRun(20)
		mu  sync.Mutex
	)
	for _, repoRevs := range args.repos {
		repoRevs := repoRevs
		if ctx.Err() != nil {
			break
		}
		if len(repoRevs.revspecs()) == 0 {
			continue
		}
		run.Acquire()
		goroutine.Go(func() {
			defer run.Release()
			repoSymbols, repoErr := searchSymbolsInRepo(ctx, repoRevs, args.query, query, limit)
			mu.Lock()
			defer mu.Unlock()
			limitHit := len(res) > limit
			repoErr = handleRepoSearchResult(common, *repoRevs, false, limitHit, repoErr)
			if repoErr != nil {
				run.Error(repoErr)
			} else {
				common.searched = append(common.searched, repoRevs.repo.URI)
			}
			if repoSymbols != nil {
				res = append(res, repoSymbols...)
				if limitHit {
					cancelAll()
				}
			}
		})
	}
	err = run.Wait()

	if len(res) > limit {
		common.limitHit = true
		res = res[:limit]
	}
	return res, common, err
}

func searchSymbolsInRepo(ctx context.Context, repoRevs *repositoryRevisions, patternInfo *patternInfo, query searchquery.Query, limit int) (res []*symbolResolver, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Search symbols in repo")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("repo", string(repoRevs.repo.URI))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	inputRev := repoRevs.revspecs()[0]
	span.SetTag("rev", inputRev)
	// Do not trigger a repo-updater lookup (e.g.,
	// backend.Repos.{GitserverRepoInfo,ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	vcsrepo := backend.Repos.VCS(repoRevs.gitserverRepo)
	commitID, err := vcsrepo.ResolveRevision(ctx, inputRev, nil)
	if err != nil {
		return nil, err
	}
	span.SetTag("commit", string(commitID))

	var excludePattern string
	if patternInfo.ExcludePattern != nil {
		excludePattern = *patternInfo.ExcludePattern
	}
	symbols, err := backend.Symbols.ListTags(ctx, protocol.SearchArgs{
		Repo:            repoRevs.repo.URI,
		CommitID:        commitID,
		Query:           patternInfo.Pattern,
		IsCaseSensitive: patternInfo.IsCaseSensitive,
		IsRegExp:        patternInfo.IsRegExp,
		IncludePatterns: patternInfo.IncludePatterns,
		ExcludePattern:  excludePattern,
		First:           limit,
	})
	baseURI, uriParseErr := uri.Parse("git://" + string(repoRevs.repo.URI) + "?" + string(commitID))
	if uriParseErr != nil {
		return nil, uriParseErr
	}
	symbolResolvers := make([]*symbolResolver, 0, len(symbols))
	for _, symbol := range symbols {
		commit := &gitCommitResolver{
			repo: &repositoryResolver{repo: repoRevs.repo},
			oid:  gitObjectID(commitID),
			// NOTE: Not all fields are set, for performance.
		}
		if inputRev != "" {
			commit.inputRev = &inputRev
		}
		symbolResolvers = append(symbolResolvers, toSymbolResolver(symbolToLSPSymbolInformation(symbol, baseURI), strings.ToLower(symbol.Language), commit))
	}
	return symbolResolvers, err
}

// symbolToLSPSymbolInformation converts a symbols service Symbol struct to an LSP SymbolInformation
// baseURI is the git://repo?rev base URI for the symbol that is extended with the file path
func symbolToLSPSymbolInformation(s protocol.Symbol, baseURI *uri.URI) lsp.SymbolInformation {
	ch := ctagsSymbolCharacter(s)
	return lsp.SymbolInformation{
		Name:          s.Name + s.Signature,
		ContainerName: s.Parent,
		Kind:          ctagsKindToLSPSymbolKind(s.Kind),
		Location: lsp.Location{
			URI: lsp.DocumentURI(baseURI.WithFilePath(s.Path).String()),
			Range: lsp.Range{
				Start: lsp.Position{Line: s.Line - 1, Character: ch},
				End:   lsp.Position{Line: s.Line - 1, Character: ch + len(s.Name)},
			},
		},
	}
}

// ctagsSymbolCharacter only outputs the line number, not the character (or range). Use the regexp it provides to
// guess the character.
func ctagsSymbolCharacter(s protocol.Symbol) int {
	if s.Pattern == "" {
		return 0
	}
	pattern := strings.TrimPrefix(s.Pattern, "/^")
	i := strings.Index(pattern, s.Name)
	if i >= 0 {
		return i
	}
	return 0
}

func ctagsKindToLSPSymbolKind(kind string) lsp.SymbolKind {
	// Ctags kinds are determined by the parser and do not (in general) match LSP symbol kinds.
	switch kind {
	case "file":
		return lsp.SKFile
	case "module":
		return lsp.SKModule
	case "namespace":
		return lsp.SKNamespace
	case "package", "subprogspec":
		return lsp.SKPackage
	case "class", "type", "service", "typedef", "union", "section", "subtype", "component":
		return lsp.SKClass
	case "method":
		return lsp.SKMethod
	case "property":
		return lsp.SKProperty
	case "field", "member", "anonMember":
		return lsp.SKField
	case "constructor":
		return lsp.SKConstructor
	case "enum", "enumerator":
		return lsp.SKEnum
	case "interface":
		return lsp.SKInterface
	case "function", "func", "subroutine", "macro", "subprogram", "procedure", "command":
		return lsp.SKFunction
	case "variable", "var", "functionVar", "define", "alias":
		return lsp.SKVariable
	case "constant", "const":
		return lsp.SKConstant
	case "string", "message", "heredoc":
		return lsp.SKString
	case "number":
		return lsp.SKNumber
	case "bool", "boolean":
		return lsp.SKBoolean
	case "array":
		return lsp.SKArray
	case "object", "literal", "map":
		return lsp.SKObject
	case "key", "label", "target", "selector", "id", "tag":
		return lsp.SKKey
	case "null":
		return lsp.SKNull
	case "enum member", "enumConstant":
		return lsp.SKEnumMember
	case "struct":
		return lsp.SKStruct
	case "event":
		return lsp.SKEvent
	case "operator":
		return lsp.SKOperator
	case "type parameter":
		return lsp.SKTypeParameter
	}
	log.Printf("Unknown ctags kind: %q", kind)
	return 0
}
