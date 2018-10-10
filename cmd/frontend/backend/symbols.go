package backend

import (
	"context"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	symbolsclient "github.com/sourcegraph/sourcegraph/pkg/symbols"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/xlang"
)

// Symbols backend.
var Symbols = &symbols{}

type symbols struct{}

// List returns symbol in a repository from language servers.
//
// Use the (lspext.WorkspaceSymbolParams).Symbol field to resolve symbols given a global ID. This is how Go symbol
// URLs (e.g., from godoc.org) are resolved.
func (symbols) List(ctx context.Context, repo api.RepoURI, commitID api.CommitID, mode string, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	if Mocks.Symbols.List != nil {
		return Mocks.Symbols.List(ctx, repo, commitID, mode, params)
	}

	var symbols []lsp.SymbolInformation
	rootURI := lsp.DocumentURI("git://" + string(repo) + "?" + string(commitID))
	err := xlang.UnsafeOneShotClientRequest(ctx, mode, rootURI, "workspace/symbol", params, &symbols)
	return symbols, err
}

// ListTags returns symbols in a repository from ctags.
func (symbols) ListTags(ctx context.Context, args protocol.SearchArgs) ([]protocol.Symbol, error) {
	result, err := symbolsclient.DefaultClient.Search(ctx, args)
	if result == nil {
		return nil, err
	}
	return result.Symbols, err
}

// MockSymbols is used by tests to mock Symbols backend methods.
type MockSymbols struct {
	List func(ctx context.Context, repo api.RepoURI, commitID api.CommitID, mode string, params lspext.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error)
}
