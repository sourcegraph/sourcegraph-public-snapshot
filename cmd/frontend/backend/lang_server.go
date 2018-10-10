package backend

import (
	"context"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// LangServer contains backend methods for communicating with language and build servers over LSP (via xlang).
var LangServer langServer

type langServer struct{}

func (langServer) WorkspaceXReferences(ctx context.Context, repo *types.Repo, commitID api.CommitID, language string, params lspext.WorkspaceReferencesParams) (result []*lspext.ReferenceInformation, err error) {
	vcs := "git" // TODO: store VCS type in *types.Repo object.
	rootURI := lsp.DocumentURI(vcs + "://" + string(repo.URI) + "?" + string(commitID))
	err = cachedUnsafeXLangCall(ctx, language, rootURI, "workspace/xreferences", params, &result)
	return result, err
}

// MockLangServer allows mocking of LangServer backend methods (by setting Mocks.LangServer's
// fields).
type MockLangServer struct {
	WorkspaceXReferences func(repo *types.Repo, commitID api.CommitID, params lspext.WorkspaceReferencesParams) ([]*lspext.ReferenceInformation, error)
}
