package httpapi

import (
	"context"
	"encoding/json"
	"net/url"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	vcsurl "github.com/sourcegraph/go-vcsurl"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/xlang"
	xlang_lspext "github.com/sourcegraph/sourcegraph/xlang/lspext"
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

	// marshalResult takes an existing interface and marshals it into result
	// via JSON.
	marshalResult := func(v interface{}, err error) error {
		if err != nil {
			return err
		}
		b, err := json.Marshal(v)
		if err != nil {
			return errors.Wrap(err, "marshaling result")
		}
		return json.Unmarshal(b, result)
	}

	switch {
	case method == "initialize":
		var init xlang_lspext.ClientProxyInitializeParams
		if err := json.Unmarshal(*params.(*json.RawMessage), &init); err != nil {
			return err
		}
		c.mode = init.InitializationOptions.Mode
		if c.mode == "" {
			// DEPRECATED: Use old Mode field if the new one is not set.
			c.mode = init.Mode
		}
		var resultRaw json.RawMessage
		if err := c.Client.Call(ctx, method, params, &resultRaw, opt...); err != nil {
			return err
		}

		// We only care about the XDefinitionProvider. Right now it implies
		// the support of XPackages as well :'(
		var initResultSubset struct {
			Capabilities struct {
				// XDefinitionProvider indicates the server provides support for
				// textDocument/xdefinition. This is a Sourcegraph extension.
				XDefinitionProvider bool `json:"xdefinitionProvider,omitempty"`
			} `json:"capabilities,omitempty"`
		}
		if err := json.Unmarshal(resultRaw, &initResultSubset); err != nil {
			return err
		}

		c.hasXDefinitionAndXPackages = initResultSubset.Capabilities.XDefinitionProvider
		//_, c.hasXDefinitionAndXPackages = xlang.HasXDefinitionAndXPackages[c.mode]
		_, c.hasCrossRepoHover = xlang.HasCrossRepoHover[c.mode]

		return json.Unmarshal(resultRaw, result)

	case !c.hasXDefinitionAndXPackages:
		break
	case method == "textDocument/definition":
		span.SetTag("LocationAbsent", "true")
		return marshalResult(c.jumpToDefCrossRepo(ctx, params, opt...))
	case method == "textDocument/hover" && c.hasCrossRepoHover:
		return marshalResult(c.hoverCrossRepo(ctx, params, opt...))

	case method == "xsymbol/hover":
		// Federation. This will only run on sourcegraph.com
		var syms []lspext.SymbolLocationInformation
		if err := json.Unmarshal(*params.(*json.RawMessage), &syms); err != nil {
			return err
		}
		return marshalResult(c.symbolHover(ctx, syms))

	case method == "xsymbol/definition":
		// Federation. This will only run on sourcegraph.com
		var syms []lspext.SymbolLocationInformation
		if err := json.Unmarshal(*params.(*json.RawMessage), &syms); err != nil {
			return err
		}
		return marshalResult(c.symbolDefinition(ctx, syms))
	}
	return c.Client.Call(ctx, method, params, result, opt...)
}

func (c *xclient) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	return c.Client.Notify(ctx, method, params, opt...)
}

func (c *xclient) Close() error {
	return c.Client.Close()
}

func (c *xclient) xdefQuery(ctx context.Context, syms []lspext.SymbolLocationInformation) (map[lsp.DocumentURI][]lsp.SymbolInformation, error) {
	span := opentracing.SpanFromContext(ctx)

	symInfos := make(map[lsp.DocumentURI][]lsp.SymbolInformation)
	// For each symbol in the xdefinition-result-derived query, compute the symbol information for that symbol
	for _, sym := range syms {

		var rootURIs []lsp.DocumentURI
		// If we can extract the repository URL from the symbol metadata, do so
		if repoURL := xlang.SymbolRepoURL(sym.Symbol); repoURL != "" {
			span.LogFields(otlog.String("event", "extracted repo directly from symbol metadata"),
				otlog.String("repoURL", repoURL))

			repoInfo, err := vcsurl.Parse(repoURL)
			if err != nil {
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			repoName := api.RepoName(string(repoInfo.RepoHost) + "/" + repoInfo.FullName)

			// We issue a workspace/symbols on the URL, so ensure we have the repo / it exists.
			repo, err := backend.Repos.GetByName(ctx, repoName)
			if err != nil {
				span.LogFields(otlog.Error(err))
				if _, isSeeOther := err.(backend.ErrRepoSeeOther); isSeeOther || errcode.IsNotFound(err) {
					span.LogFields(otlog.String("event", "ignoring not found error"))
					continue
				}
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			rev, err := backend.Repos.ResolveRev(ctx, repo, "")
			if err != nil {
				span.LogFields(otlog.Error(err))
				if vcs.IsRepoNotExist(err) {
					span.LogFields(otlog.String("event", "ignoring not found error"))
					continue
				}
				return nil, errors.Wrap(err, "extract repo URL from symbol metadata")
			}
			rootURIs = append(rootURIs, lsp.DocumentURI(string(repoInfo.VCS)+"://"+string(repoName)+"?"+string(rev)))
		} else { // if we can't extract the repository URL directly, we have to consult the pkgs database
			pkgDescriptor, ok := xlang.SymbolPackageDescriptor(sym.Symbol, c.mode)
			if !ok {
				continue
			}

			span.LogFields(otlog.String("event", "cross-repo jump to def"))
			pkgs, err := db.Pkgs.ListPackages(ctx, &api.ListPackagesOp{PkgQuery: pkgDescriptor, Lang: c.mode, Limit: 1})
			if err != nil {
				return nil, errors.Wrap(err, "getting repo by package db query")
			}
			span.LogFields(otlog.String("event", "listed repository packages"))
			for _, pkg := range pkgs {
				repo, err := backend.Repos.Get(ctx, pkg.RepoID)
				if err != nil {
					return nil, errors.Wrap(err, "fetch repo for package")
				}
				var commit api.CommitID
				if repo.IndexedRevision != nil {
					commit = *repo.IndexedRevision
				} else {
					var err error
					commit, err = backend.Repos.ResolveRev(ctx, repo, "")
					if err != nil {
						return nil, errors.Wrap(err, "resolve revision for package repo")
					}
				}
				rootURIs = append(rootURIs, lsp.DocumentURI("git://"+string(repo.Name)+"?"+string(commit)))
			}
			span.LogFields(otlog.String("event", "resolved rootURIs"))
		}

		// Issue a workspace/symbol for each repository that provides a definition for the symbol
		for _, rootURI := range rootURIs {
			params := &lspext.WorkspaceSymbolParams{Symbol: sym.Symbol, Limit: 10}
			var repoSymInfos []lsp.SymbolInformation
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootURI, "workspace/symbol", params, &repoSymInfos); err != nil {
				return nil, errors.Wrap(err, "resolving symbol to location")
			}
			symInfos[rootURI] = repoSymInfos
		}
		span.LogFields(otlog.String("event", "done issuing workspace/symbol requests"))
	}
	return symInfos, nil
}

func (c *xclient) jumpToDefCrossRepo(ctx context.Context, params interface{}, opt ...jsonrpc2.CallOption) ([]lsp.Location, error) {
	// Issue xdefinition request
	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return nil, err
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

	symLocs, err := c.symbolDefinition(ctx, nolocSyms)
	if err != nil {
		return nil, err
	}
	locs = append(locs, symLocs...)

	// Failed to find the definition locally, try symbolDefinition on Sourcegraph.com
	// which may have indexed the OSS repo used.
	if len(locs) == 0 && len(nolocSyms) > 0 && conf.JumpToDefOSSIndexEnabled() {
		// HACK we need a valid rootURI, even though we are doing symbol queries.
		rootURI := lsp.DocumentURI("git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06")
		err := xlang.RemoteOneShotClientRequest(ctx, sourcegraphDotComBaseURL, c.mode, rootURI, "xsymbol/definition", nolocSyms, &locs)
		if err != nil {
			return nil, err
		}
		return locs, nil
	}

	return locs, nil
}

func (c *xclient) symbolDefinition(ctx context.Context, syms []lspext.SymbolLocationInformation) ([]lsp.Location, error) {
	symInfos, err := c.xdefQuery(ctx, syms)
	if err != nil {
		return nil, err
	}

	locs := make([]lsp.Location, 0, len(symInfos))
	for _, repoSymInfos := range symInfos {
		for _, symInfo := range repoSymInfos {
			locs = append(locs, symInfo.Location)
		}
	}
	return locs, nil
}

// hoverCrossRepo translates hover requests in the current repository to a
// hover request on the definition in the definition's repository.
//
// Algorithm:
//
// 1. If we are hovering over a symbol in the current repository, use the
//    normal textDocument/hover.
// 2. Use textDocument/xdefinition (sg extension) to retrieve symbol information.
// 3. symbolHover: Using the symbols, use the first successful hover in the
//    definition repos.
// 4. If we do not find a non-empty hover and federation is enabled, we send
//    the package query to Sourcegraph.com's xlang API. The assumption is the
//    dependency is an OSS package so we can consult our public index. If the
//    response is non-empty we return it.
// 5. If we do not find a non-empty hover, fallback to the normal hover.
func (c *xclient) hoverCrossRepo(ctx context.Context, params interface{}, opt ...jsonrpc2.CallOption) (*lsp.Hover, error) {
	// Note: we can't parallelize the hover and xdefinition requests
	// without breaking the request cancellation logic used by LSP
	// proxy

	// xdefinition request
	var syms []lspext.SymbolLocationInformation
	if err := c.Client.Call(ctx, "textDocument/xdefinition", params, &syms, opt...); err != nil {
		return nil, errors.Wrap(err, "hoverCrossRepo: textDocument/xdefinition error")
	}

	// hover request
	var hover lsp.Hover
	if err := c.Client.Call(ctx, "textDocument/hover", params, &hover, opt...); err != nil {
		return nil, errors.Wrap(err, "hoverCrossRepo: textDocument/hover error")
	}

	// return local hover if local definition found
	for _, sym := range syms {
		if sym.Location != (lsp.Location{}) {
			return &hover, nil
		}
	}

	// Cross repo hover is done via the symbols only.
	xhov, err := c.symbolHover(ctx, syms)
	if err != nil {
		return nil, err
	}
	if len(xhov.Contents) > 0 {
		// Range is for the queried token, so we need to use the local
		// hover range.
		xhov.Range = hover.Range
		return xhov, nil
	}

	// Failed to find the hover locally, try symbolHover on Sourcegraph.com
	// which may have indexed the OSS repo used.
	if conf.JumpToDefOSSIndexEnabled() {
		// HACK we need a valid rootURI, even though we are doing symbol queries.
		rootURI := lsp.DocumentURI("git://github.com/gorilla/mux?4dbd923b0c9e99ff63ad54b0e9705ff92d3cdb06")
		var remoteHov lsp.Hover
		err := xlang.RemoteOneShotClientRequest(ctx, sourcegraphDotComBaseURL, c.mode, rootURI, "xsymbol/hover", syms, &remoteHov)
		if err != nil {
			return nil, err
		}
		if len(remoteHov.Contents) > 0 {
			// Range is for the queried token, so we need to use the local
			// hover range.
			remoteHov.Range = hover.Range
			return &remoteHov, nil
		}
	}

	// Fallback to local hover contents.
	return &hover, nil
}

// symbolHover finds a hover contents for the given symbols.
//
// Algorithm:
//
// 1. xdefQuery: Consult our packages index to find potential repositories
//    containing the symbols.
// 2. xdefQuery: For each potential repository use workspace/symbol with our
//    symbol query (sg extension).
// 3. For each symbol do a textDocument/hover. The first non-empty hover
//    content we return to the user.
func (c *xclient) symbolHover(ctx context.Context, syms []lspext.SymbolLocationInformation) (*lsp.Hover, error) {
	symInfos, err := c.xdefQuery(ctx, syms)
	if err != nil {
		return nil, err
	}

	// return first hover found
	for rootURI, repoSymInfos := range symInfos {
		for _, symInfo := range repoSymInfos {
			pos := symInfo.Location.Range.Start
			pos.Character++
			p := lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: symInfo.Location.URI},
				Position:     pos,
			}
			var xhov lsp.Hover
			if err := xlang.UnsafeOneShotClientRequest(ctx, c.mode, rootURI, "textDocument/hover", p, &xhov); err != nil {
				return nil, errors.Wrap(err, "hoverCrossRepo: external textDocument/hover error")
			}
			if len(xhov.Contents) > 0 {
				return &xhov, nil
			}
		}
	}

	// nothing found, so empty response
	return &lsp.Hover{}, nil
}

var sourcegraphDotComBaseURL = &url.URL{Scheme: "https", Host: "sourcegraph.com"}
