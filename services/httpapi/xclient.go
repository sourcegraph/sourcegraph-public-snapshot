package httpapi

import (
	"context"
	"encoding/json"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	vcsurl "github.com/sourcegraph/go-vcsurl"
	"github.com/sourcegraph/jsonrpc2"
)

// xclient is an LSP client that transparently wraps xlang.Client,
// except that it translates textDocument/definition requests into a
// series of requests that computes the cross-repo jump-to-definition
// result.
type xclient struct {
	*xlang.Client

	mode string
}

func (c *xclient) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	if method != "textDocument/definition" {
		if method == "initialize" {
			var init xlang.ClientProxyInitializeParams
			json.Unmarshal(*params.(*json.RawMessage), &init)
			c.mode = init.Mode
		}
		return c.Client.Call(ctx, method, params, result, opt...)
	}

	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return err
	}
	locs := make([]lsp.Location, 0, len(syms))
	for _, sym := range syms {
		if sym.Location != (lsp.Location{}) {
			locs = append(locs, sym.Location)
			continue
		}

		var rootPaths []string
		if repoURL := extractRepoURL(sym.Symbol); repoURL != "" {
			repoInfo, err := vcsurl.Parse(repoURL)
			if err != nil {
				return err
			}
			repoURI := string(repoInfo.RepoHost) + "/" + repoInfo.FullName
			repo, err := backend.Repos.GetByURI(ctx, repoURI)
			if err != nil {
				return err
			}
			rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
			if err != nil {
				return err
			}
			rootPaths = append(rootPaths, string(repoInfo.VCS)+"://"+repoURI+"?"+rev.CommitID)
		} else {
			subSelector, exists := subSelectors[c.mode]
			if !exists {
				locs = append(locs, sym.Location)
				continue
			}

			pkgs, err := backend.Pkgs.ListPackages(ctx, &sourcegraph.ListPackagesOp{PkgQuery: subSelector(sym.Symbol), Lang: c.mode, Limit: 100})
			if err != nil {
				return err
			}

			for _, pkg := range pkgs {
				repo, err := backend.Repos.Get(ctx, &sourcegraph.RepoSpec{pkg.RepoID})
				if err != nil {
					return err
				}
				rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
				if err != nil {
					return err
				}
				rootPaths = append(rootPaths, "git://"+repo.URI+"?"+rev.CommitID)
			}
		}

		for _, rootPath := range rootPaths {
			params := &lspext.WorkspaceSymbolParams{Symbol: sym.Symbol, Limit: 10}
			var syms []lsp.SymbolInformation
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootPath, "workspace/symbol", params, &syms); err != nil {
				return err
			}
			for _, sym := range syms {
				locs = append(locs, sym.Location)
			}
		}
	}
	locBytes, err := json.Marshal(locs)
	if err != nil {
		return err
	}
	return json.Unmarshal(locBytes, result)
}

func (c *xclient) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	return c.Client.Notify(ctx, method, params, opt...)
}

func (c *xclient) Close() error {
	return c.Client.Close()
}

// TODO(beyang): copy-pasted from services/backend/defs_refs.go
var subSelectors = map[string]func(lspext.SymbolDescriptor) map[string]interface{}{
	"go": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"package": symbol["package"],
		}
	},
	"php": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		if _, ok := symbol["package"]; !ok {
			// package can be missing if the symbol did not belong to a package, e.g. a project without
			// a composer.json file. In this case, there are no external references to this symbol.
			return nil
		}
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
	"typescript": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
}

// extractRepoURL returns the repository URL extracted from the
// package metadata at the JSON path
// `symDescriptor.package.repoURL`. If that does not exist, it returns
// the empty string.
func extractRepoURL(symDescriptor lspext.SymbolDescriptor) string {
	pkgData := symDescriptor["package"]
	if pkgData, ok := pkgData.(map[string]interface{}); ok {
		repoURL := pkgData["repoURL"]
		if repoURL, ok := repoURL.(string); ok {
			return repoURL
		}
	}
	return ""
}
