package httpapi

import (
	"context"
	"encoding/json"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	xlspext "sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"

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

	hasXDefinitionAndXPackages bool
	hasCrossRepoHover          bool
	mode                       string
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

	switch {
	case method == "initialize":
		var init xlspext.ClientProxyInitializeParams
		if err := json.Unmarshal(*params.(*json.RawMessage), &init); err != nil {
			return err
		}
		c.mode = init.InitializationOptions.Mode
		if c.mode == "" {
			// DEPRECATED: Use old Mode field if the new one is not set.
			c.mode = init.Mode
		}
		_, c.hasXDefinitionAndXPackages = xlang.HasXDefinitionAndXPackages[c.mode]
		_, c.hasCrossRepoHover = xlang.HasCrossRepoHover[c.mode]
		return c.Client.Call(ctx, method, params, result, opt...)
	case !c.hasXDefinitionAndXPackages:
		break
	case method == "textDocument/definition":
		span.SetTag("LocationAbsent", "true")
		return c.jumpToDefCrossRepo(ctx, params, result, opt...)
	case method == "textDocument/hover":
		return c.hoverCrossRepo(ctx, params, result, opt...)
	}
	return c.Client.Call(ctx, method, params, result, opt...)
}

func (c *xclient) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	return c.Client.Notify(ctx, method, params, opt...)
}

func (c *xclient) Close() error {
	return c.Client.Close()
}

func (c *xclient) xdefQuery(ctx context.Context, syms []lspext.SymbolLocationInformation, includeHover bool) (map[string][]lsp.SymbolInformation, error) {
	span := opentracing.SpanFromContext(ctx)

	symInfos := make(map[string][]lsp.SymbolInformation)
	// For each symbol in the xdefinition-result-derived query, compute the symbol information for that symbol
	for _, sym := range syms {

		var rootPaths []string
		// If we can extract the repository URL from the symbol metadata, do so
		if repoURL := xlang.SymbolRepoURL(sym.Symbol); repoURL != "" {
			span.LogEvent("extracted repo directly from symbol metadata")

			repoInfo, err := vcsurl.Parse(repoURL)
			if err != nil {
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			repoURI := string(repoInfo.RepoHost) + "/" + repoInfo.FullName
			// SECURITY NOTE: The LSP proxy DOES NOT check permissions, so this line is a necessary
			// security check
			repo, err := backend.Repos.GetByURI(ctx, repoURI)
			if err != nil {
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
			if err != nil {
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
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
				return nil, errors.Wrap(err, "getting repo by package db query")
			}
			span.LogEvent("listed repository packages")

			for _, pkg := range pkgs {
				repo, err := backend.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: pkg.RepoID})
				if err != nil {
					return nil, errors.Wrap(err, "fetch repo for package")
				}
				var commit string
				if repo.IndexedRevision != nil {
					commit = *repo.IndexedRevision
				} else {
					rev, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID})
					if err != nil {
						return nil, errors.Wrap(err, "resolve revision for package repo")
					}
					commit = rev.CommitID
				}
				// TODO: store VCS type in *sourcegraph.Repo object.
				rootPaths = append(rootPaths, "git://"+repo.URI+"?"+commit)
			}
			span.LogEvent("resolved rootPaths")
		}

		// Issue a workspace/symbol for each repository that provides a definition for the symbol
		for _, rootPath := range rootPaths {
			params := &lspext.WorkspaceSymbolParams{Symbol: sym.Symbol, Limit: 10}
			var repoSymInfos []lsp.SymbolInformation
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootPath, "workspace/symbol", params, &repoSymInfos); err != nil {
				return nil, errors.Wrap(err, "resolving symbol to location")
			}
			symInfos[rootPath] = repoSymInfos
		}
		span.LogEvent("done issuing workspace/symbol requests")
	}
	return symInfos, nil
}

func (c *xclient) jumpToDefCrossRepo(ctx context.Context, params, result interface{}, opt ...jsonrpc2.CallOption) (err error) {
	// Issue xdefinition request
	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return err
	}
	locs := make([]lsp.Location, 0, len(syms))

	var nolocSyms []lspext.SymbolLocationInformation
	for _, sym := range syms {
		// If a concrete location is already present, just use that
		if sym.Location != (lsp.Location{}) {
			locs = append(locs, sym.Location)
		} else {
			nolocSyms = append(nolocSyms, sym)
		}
	}

	symInfos, err := c.xdefQuery(ctx, nolocSyms, false)
	if err != nil {
		return err
	}
	for _, repoSymInfos := range symInfos {
		for _, symInfo := range repoSymInfos {
			locs = append(locs, symInfo.Location)
		}
	}
	locBytes, err := json.Marshal(locs)
	if err != nil {
		return errors.Wrap(err, "marshaling locations")
	}
	return json.Unmarshal(locBytes, result)
}

func (c *xclient) hoverCrossRepo(ctx context.Context, params, result interface{}, opt ...jsonrpc2.CallOption) (err error) {
	// Note: we can't parallelize the hover and xdefinition requests
	// without breaking the request cancellation logic used by LSP
	// proxy

	// xdefinition request
	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return errors.Wrap(err, "hoverCrossRepo: textDocument/xdefinition error")
	}

	// hover request
	var hover lsp.Hover
	if err := c.Client.Call(ctx, "textDocument/hover", params, &hover, opt...); err != nil {
		return errors.Wrap(err, "hoverCrossRepo: textDocument/hover error")
	}

	foundLoc := false
	for _, sym := range syms {
		if sym.Location != (lsp.Location{}) {
			foundLoc = true
			break
		}
	}
	if foundLoc { // return local hover if local definition found
		h, err := json.Marshal(hover)
		if err != nil {
			return err
		}
		return json.Unmarshal(h, &result)
	}

	symInfos, err := c.xdefQuery(ctx, syms, true)
	if err != nil {
		return err
	}
	var crossHov lsp.Hover
	crossHov.Range = hover.Range
Outer: // display first hover found
	for rootPath, repoSymInfos := range symInfos {
		for _, symInfo := range repoSymInfos {
			pos := symInfo.Location.Range.Start
			pos.Character++
			p := lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: symInfo.Location.URI},
				Position:     pos,
			}
			var xhov lsp.Hover
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootPath, "textDocument/hover", p, &xhov); err != nil {
				return errors.Wrap(err, "hoverCrossRepo: external textDocument/hover error")
			}
			if len(xhov.Contents) > 0 {
				crossHov.Contents = xhov.Contents
				break Outer
			}
		}
	}
	if len(crossHov.Contents) == 0 { // fall back to local hover contents
		crossHov.Contents = hover.Contents
	}
	h, err := json.Marshal(crossHov)
	if err != nil {
		return errors.Wrap(err, "marshaling crossHov")
	}
	return json.Unmarshal(h, result)
}
