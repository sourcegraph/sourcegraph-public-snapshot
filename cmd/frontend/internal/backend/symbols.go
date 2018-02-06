package backend

import (
	"context"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	symbolsclient "sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// Symbols backend.
var Symbols = &symbols{}

type symbols struct{}

// List resolves a symbol in a repository.
//
// Use the (lspext.WorkspaceSymbolParams).Symbol field to resolve symbols given a global ID. This is how Go symbol
// URLs (e.g., from godoc.org) are resolved.
func (symbols) List(ctx context.Context, repo api.RepoURI, commitID api.CommitID, mode string, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	if Mocks.Symbols.List != nil {
		return Mocks.Symbols.List(ctx, repo, commitID, mode, params)
	}

	if mode == "tags" {
		return (symbols{}).listTags(ctx, repo, commitID, params)
	}

	var symbols []lsp.SymbolInformation
	rootURI := lsp.DocumentURI("git://" + string(repo) + "?" + string(commitID))
	err := xlang.UnsafeOneShotClientRequest(ctx, mode, rootURI, "workspace/symbol", params, &symbols)
	return symbols, err
}

func (symbols) listTags(ctx context.Context, repo api.RepoURI, commitID api.CommitID, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	result, err := symbolsclient.DefaultClient.Search(ctx, protocol.SearchArgs{
		Repo:     repo,
		CommitID: commitID,
		Query:    params.Query,
		First:    params.Limit,
	})
	if err != nil {
		return nil, err
	}

	baseURI, err := uri.Parse("git://" + string(repo) + "?" + string(commitID))
	if err != nil {
		return nil, err
	}

	symbols := make([]lsp.SymbolInformation, 0, len(result.Symbols))
	for _, s := range result.Symbols {
		if s.FileLimited {
			continue
		}
		ch := ctagsSymbolCharacter(s)
		symbols = append(symbols, lsp.SymbolInformation{
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
		})
	}
	return symbols, nil
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
	case "variable", "var", "functionVar", "define":
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
	case "unknown", "":
		return 0
	}
	// log.Printf("Unknown ctags kind: %q", kind)
	return 0 // unknown
}

// MockSymbols is used by tests to mock Symbols backend methods.
type MockSymbols struct {
	List func(ctx context.Context, repo api.RepoURI, commitID api.CommitID, mode string, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error)
}
