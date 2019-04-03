package symbols

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"gopkg.in/inconshreveable/log15.v2"
)

type ComputeSymbolsArgs struct {
	CommitID        api.CommitID
	RepoName        api.RepoName
	Query           *string
	First           *int32
	IncludePatterns *[]string
	ExcludePattern  *string
	IsRegExp        *bool
	IsCaseSensitive *bool
}

func (a ComputeSymbolsArgs) toSearchArgs() protocol.SearchArgs {
	args := protocol.SearchArgs{
		Repo:     a.RepoName,
		CommitID: a.CommitID,
		First:    LimitOrDefault(a.First) + 1,
	}
	if a.Query != nil {
		args.Query = *a.Query
	}
	if a.IncludePatterns != nil {
		args.IncludePatterns = *a.IncludePatterns
	}
	if a.ExcludePattern != nil {
		args.ExcludePattern = *a.ExcludePattern
	}
	if a.IsRegExp != nil {
		args.IsRegExp = *a.IsRegExp
	}
	if a.IsCaseSensitive != nil {
		args.IsCaseSensitive = *a.IsCaseSensitive
	}
	return args
}

func ComputeSymbols(ctx context.Context, args ComputeSymbolsArgs) (res []*Symbol, err error) {
	ctx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.New("processing symbols is taking longer than expected. Try again in a while")
		}
	}()

	baseURI, err := gituri.Parse("git://" + string(args.RepoName) + "?" + string(args.CommitID))
	if err != nil {
		return nil, err
	}
	if baseURI == nil {
		return
	}

	symbols, err := backend.Symbols.ListTags(ctx, args.toSearchArgs())
	out := make([]*Symbol, len(symbols))
	for i, symbol := range symbols {
		srange := symbolRange(symbol)

		uri := baseURI.WithFilePath(symbol.Path)

		out[i] = &Symbol{
			Symbol:   symbol,
			Range:    &srange,
			Language: strings.ToLower(symbol.Language),
			CommitID: args.CommitID,
			RepoName: args.RepoName,
			URI:      uri,
		}
	}

	return out, err
}

func LimitOrDefault(first *int32) int {
	if first == nil {
		return 100
	}
	return int(*first)
}

type Symbol struct {
	Symbol   protocol.Symbol
	Language string
	Range    *lsp.Range

	CommitID api.CommitID
	RepoName api.RepoName
	URI      *gituri.URI
}

func symbolRange(s protocol.Symbol) lsp.Range {
	ch := CtagsSymbolCharacter(s)
	return lsp.Range{
		Start: lsp.Position{Line: s.Line - 1, Character: ch},
		End:   lsp.Position{Line: s.Line - 1, Character: ch + len(s.Name)},
	}
}

func (r *Symbol) Kind() string {
	return strings.ToUpper(CtagsKindToLSPSymbolKind(r.Symbol.Kind).String())
}

// CtagsSymbolCharacter only outputs the line number, not the character (or range). Use the regexp it provides to
// guess the character.
func CtagsSymbolCharacter(s protocol.Symbol) int {
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

func CtagsKindToLSPSymbolKind(kind string) lsp.SymbolKind {
	// Ctags kinds are determined by the parser and do not (in general) match LSP symbol kinds.
	switch kind {
	case "file":
		return lsp.SKFile
	case "module":
		return lsp.SKModule
	case "namespace":
		return lsp.SKNamespace
	case "package", "packageName", "subprogspec":
		return lsp.SKPackage
	case "class", "type", "service", "typedef", "union", "section", "subtype", "component":
		return lsp.SKClass
	case "method", "methodSpec":
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
	case "function", "func", "subroutine", "macro", "subprogram", "procedure", "command", "singletonMethod":
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
	case "type parameter", "annotation":
		return lsp.SKTypeParameter
	}
	log15.Debug("Unknown ctags kind", "kind", kind)
	return 0
}
