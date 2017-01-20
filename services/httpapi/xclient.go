package httpapi

import (
	"context"
	"encoding/json"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
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

	hasXDefinition bool
	mode           string
}

// hasXDefinition is the hardcoded list of languages that provide
// textDocument/xdefinition.  We cannot rely on the value returned
// from the LSP proxy, because that does not pass through the value of
// the initialize result.
var hasXDefinition = map[string]struct{}{
	"go":         struct{}{},
	"typescript": struct{}{},
	"php":        struct{}{},
}

// Call transparently wraps xlang.Client.Call *except* for `textDocument/definition` if the language
// server is a textDocument/xdefinition provider. In that case, this method invokes
// `textDocument/xdefinition` instead. If the result contains a non-zero `Location` field, then that
// is returned to the client as if it came from `textDocument/definition`. If the location is zero,
// then that means the definition did not exist locally. The method will locate the definition in an
// external repository and return that to the client as if it came from a single
// `textDocument/definition` call.
//
// SECURITY NOTE: Call also verifies permissions for cross-repo jumps. Any changes to this method
// should preserve this property.
func (c *xclient) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "xclient.Call")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("Method", method)

	if method == "initialize" {
		var init xlang.ClientProxyInitializeParams
		if err := json.Unmarshal(*params.(*json.RawMessage), &init); err != nil {
			return err
		}
		c.mode = init.Mode
		_, c.hasXDefinition = hasXDefinition[c.mode]
		return c.Client.Call(ctx, method, params, result, opt...)
	} else if method != "textDocument/definition" || !c.hasXDefinition {
		return c.Client.Call(ctx, method, params, result, opt...)
	}

	span.SetTag("LocationAbsent", "true")

	// Issue xdefinition request
	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return err
	}
	locs := make([]lsp.Location, 0, len(syms))
	// For each symbol in the xdefinition result, compute the location(s) for that symbol
	for _, sym := range syms {
		// If a concrete location is already present, just use that
		if sym.Location != (lsp.Location{}) {
			locs = append(locs, sym.Location)
			continue
		}

		var rootPaths []string
		// If we can extract the repository URL from the symbol metadata, do so
		if repoURL := xlang.SymbolRepoURL(sym.Symbol); repoURL != "" {
			span.LogEvent("extracted repo directly from symbol metadata")

			repoInfo, err := vcsurl.Parse(repoURL)
			if err != nil {
				return errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			repoURI := string(repoInfo.RepoHost) + "/" + repoInfo.FullName
			// SECURITY NOTE: The LSP proxy DOES NOT check permissions, so this line is a necessary
			// security check
			repo, err := backend.Repos.GetByURI(ctx, repoURI)
			if err != nil {
				return errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
			if err != nil {
				return errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			rootPaths = append(rootPaths, string(repoInfo.VCS)+"://"+repoURI+"?"+rev.CommitID)
		} else { // if we can't extract the repository URL directly, we have to consult the pkgs database
			pkgDescriptor, ok := xlang.SymbolPackageDescriptor(sym.Symbol, c.mode)
			if !ok {
				continue
			}

			span.LogEvent("cross-repo jump to def")
			pkgs, err := backend.Pkgs.ListPackages(ctx, &sourcegraph.ListPackagesOp{PkgQuery: pkgDescriptor, Lang: c.mode, Limit: 1})
			if err != nil {
				return errors.Wrap(err, "getting repo by package db query")
			}
			span.LogEvent("listed repository packages")

			for _, pkg := range pkgs {
				repo, err := backend.Repos.Get(ctx, &sourcegraph.RepoSpec{pkg.RepoID})
				if err != nil {
					return errors.Wrap(err, "fetch repo for package")
				}
				rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
				if err != nil {
					return errors.Wrap(err, "resolve revision for package repo")
				}
				// TODO: store VCS type in *sourcegraph.Repo object.
				rootPaths = append(rootPaths, "git://"+repo.URI+"?"+rev.CommitID)
			}
			span.LogEvent("resolved rootPaths")
		}

		// Issue a workspace/symbol for each repository that provides a definition for the symbol
		for _, rootPath := range rootPaths {
			params := &lspext.WorkspaceSymbolParams{Symbol: sym.Symbol, Limit: 10}
			var syms []lsp.SymbolInformation
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootPath, "workspace/symbol", params, &syms); err != nil {
				return errors.Wrap(err, "resolving symbol to location")
			}
			for _, sym := range syms {
				locs = append(locs, sym.Location)
			}
		}
		span.LogEvent("done issuing workspace/symbol requests")
	}
	locBytes, err := json.Marshal(locs)
	if err != nil {
		return errors.Wrap(err, "marshaling locations")
	}
	return json.Unmarshal(locBytes, result)
}

func (c *xclient) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	return c.Client.Notify(ctx, method, params, opt...)
}

func (c *xclient) Close() error {
	return c.Client.Close()
}
