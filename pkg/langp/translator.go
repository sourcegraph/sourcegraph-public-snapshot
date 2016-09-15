// Package langp implements Language Processor utilities.
package langp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// Translator is an HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests.
type Translator struct {
	// Addr is the address of the LSP server which translation should occur
	// against.
	Addr string

	// Preparer is the workspace preparer that will be used.
	Preparer *Preparer

	// ResolveFile is used to convert file URIs returned by LSP into
	// repo specific paths that Sourcegraph can use.
	//
	// This is where language processors should perform language-specific
	// tasks like translating paths to fetched dependencies into a path
	// inside the fetched dependency.
	//
	// workspace, repo, commit are the same as passed to PrepareDeps and
	// PrepareRepo. uri is the uri that requires resolving. It is usually
	// a file:/// relative to the workspace.
	ResolveFile func(workspace, repo, commit, uri string) (*File, error)

	// FileURI, if non-nil, is called to form the file URI which is sent to a
	// language server. Provided is the repo and commit, and a file URI which
	// is relative to the repository root.
	//
	// FileURI should return a file URI which is relative to the previously
	// prepared workspace directory.
	FileURI func(repo, commit, file string) string
}

// New creates a new HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests
// using the specified options.
func New(opts *Translator) http.Handler {
	t := &translator{
		Translator: opts,
		workspace:  opts.Preparer,
	}
	return NewServer(t)
}

type translator struct {
	*Translator
	mux       *http.ServeMux
	workspace *Preparer
}

func (t *translator) Prepare(ctx context.Context, r *RepoRev) error {
	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := t.workspace.PrepareTimeout(ctx, r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return err
	}
	return nil
}

func (t *translator) DefSpecToPosition(ctx context.Context, defSpec *DefSpec) (*Position, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	importPath, _ := ResolveRepoAlias(defSpec.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "exported " + importPath + "/...",
	}

	// TODO(slimsag): cache symbol information for quicker access
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "workspace/symbol", p, &respSymbol)
	defer observe(ctx, start, workspaceStart, defSpec.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	for _, s := range respSymbol {
		// Collect out refs only from name.
		if s.Name != defSpec.Path {
			continue
		}

		// TODO(slimsag): how do we handle def keys that map to multiple identical
		// positions? e.g. see the results for:
		//
		// 	curl -s -H "Content-Type: application/json" -X POST -d '{"Repo":"github.com/slimsag/mux","Commit":"780415097119f6f61c55475fe59b66f3c3e9ea53","Def":"GoPackage/github.com/slimsag/mux/-/Router/Match"}' http://localhost:4141/defspec-to-position
		//
		// Right now we assume the first one is right, but this is likely to fail
		// in some cases? Maybe we should have a count/index as part of the key.
		f, err := t.resolveFile(defSpec.Repo, defSpec.Commit, s.Location.URI)

		// Collect out refs only from unit.
		if f.Repo != defSpec.Unit {
			continue
		}

		if err != nil {
			return nil, err
		}
		return &Position{
			Repo:      f.Repo,
			Commit:    f.Commit,
			File:      f.Path,
			Line:      s.Location.Range.Start.Line,
			Character: s.Location.Range.Start.Character,
		}, nil
	}
	// TODO: formalize not-found errors
	return nil, errors.New("position for def key not found")
}

func (t *translator) Definition(ctx context.Context, pos *Position) (*Range, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}

	// TODO: according to spec this could be lsp.Location OR []lsp.Location
	var respDef []lsp.Location
	err = t.lspDo(ctx, rootPath, "textDocument/definition", p, &respDef)
	defer observe(ctx, start, workspaceStart, pos.Repo, err, len(respDef) == 0)
	if err != nil {
		return nil, err
	}

	// TODO: standardize our response when there are no results.
	var r Range
	if len(respDef) > 0 {
		f, err := t.resolveFile(pos.Repo, pos.Commit, respDef[0].URI)
		if err != nil {
			return nil, err
		}
		r = Range{
			Repo:           f.Repo,
			Commit:         f.Commit,
			File:           f.Path,
			StartLine:      respDef[0].Range.Start.Line,
			StartCharacter: respDef[0].Range.Start.Character,
			EndLine:        respDef[0].Range.End.Line,
			EndCharacter:   respDef[0].Range.End.Character,
		}
	}
	return &r, nil
}

func (t *translator) Hover(ctx context.Context, pos *Position) (*Hover, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}

	var respHover lsp.Hover
	err = t.lspDo(ctx, rootPath, "textDocument/hover", p, &respHover)
	defer observe(ctx, start, workspaceStart, pos.Repo, err, err == nil && len(respHover.Contents) == 0)
	if err != nil {
		return nil, err
	}

	// extract DefSpec if we have the data
	var defSpec *DefSpec
	var defInfo struct {
		URI      string
		UnitType string
		Unit     string
		Path     string
	}
	for _, m := range respHover.Contents[1:] {
		if m.Language == "text/definfo" {
			err := json.Unmarshal([]byte(m.Value), &defInfo)
			if err != nil {
				return nil, err
			}
		}
	}
	if defInfo.URI != "" {
		f, err := t.resolveFile(pos.Repo, pos.Commit, defInfo.URI)
		if err != nil {
			return nil, err
		}
		defSpec = &DefSpec{
			Repo:     f.Repo,
			Commit:   f.Commit,
			Unit:     defInfo.Unit,
			UnitType: defInfo.UnitType,
			Path:     defInfo.Path,
		}
	}

	hover := HoverFromLSP(&respHover)
	hover.DefSpec = defSpec
	return hover, nil
}

func (t *translator) LocalRefs(ctx context.Context, pos *Position) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking references,
	// but good enough for now.
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}

	var resp []lsp.Location
	err = t.lspDo(ctx, rootPath, "textDocument/references", p, &resp)
	defer observe(ctx, start, workspaceStart, pos.Repo, err, len(resp) == 0)
	if err != nil {
		return nil, err
	}
	refs := make([]*Range, 0, len(resp))
	for _, r := range resp {
		f, err := t.resolveFile(pos.Repo, pos.Commit, r.URI)
		if err != nil {
			return nil, err
		}
		refs = append(refs, &Range{
			Repo:           f.Repo,
			Commit:         f.Commit,
			File:           f.Path,
			StartLine:      r.Range.Start.Line,
			StartCharacter: r.Range.Start.Character,
			EndLine:        r.Range.End.Line,
			EndCharacter:   r.Range.End.Character,
		})
	}
	return &RefLocations{Refs: refs}, nil
}

func (t *translator) ExternalRefs(ctx context.Context, r *RepoRev) (*ExternalRefs, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	importPath, _ := ResolveRepoAlias(r.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "external " + importPath + "/...",
	}

	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "workspace/symbol", p, &respSymbol)
	defer observe(ctx, start, workspaceStart, r.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	var defs []*DefSpec
	// TODO(keegancsmith) go specific
	for _, s := range respSymbol {
		// TODO(keegancsmith) we should inspect the workspace to find
		// out the repo of the dependency
		commit := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		pkgParts := strings.Split(s.ContainerName, "/")
		var repo, unit string
		if len(pkgParts) < 3 {
			// Hack for stdlib
			repo = "github.com/golang/go"
			unit = s.ContainerName
		} else {
			repo = strings.Join(pkgParts[:3], "/")
			unit = strings.Join(pkgParts, "/")
		}
		defs = append(defs, &DefSpec{
			Repo:     repo,
			Commit:   commit,
			UnitType: "GoPackage",
			Unit:     unit,
			Path:     s.Name,
		})
	}
	return &ExternalRefs{Defs: defs}, nil
}

func (t *translator) DefSpecRefs(ctx context.Context, defSpec *DefSpec) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// Unit belongs to Repo indicates query local references.
	queryType := "defspec-refs-external"
	if strings.HasPrefix(defSpec.Unit, defSpec.Repo) {
		queryType = "defspec-refs-internal"
	}
	importPath, _ := ResolveRepoAlias(defSpec.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: queryType + " " + importPath + "/...",
	}

	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "defspec-refs-internal", p, &respSymbol)
	defer observe(ctx, start, workspaceStart, defSpec.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	var refs []*Range
	// TODO(keegancsmith) go specific
	for _, s := range respSymbol {
		pkgParts := strings.Split(s.ContainerName, "/")
		var unit string
		if len(pkgParts) < 3 {
			unit = s.ContainerName
		} else {
			unit = strings.Join(pkgParts, "/")
		}

		// Collect out refs only from def unit and name.
		if unit != defSpec.Unit || s.Name != defSpec.Path {
			continue
		}

		f, err := t.resolveFile(defSpec.Repo, defSpec.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}

		refs = append(refs, &Range{
			Repo:           f.Repo,
			Commit:         f.Commit,
			File:           f.Path,
			StartLine:      s.Location.Range.Start.Line,
			EndLine:        s.Location.Range.End.Line,
			StartCharacter: s.Location.Range.Start.Character,
			EndCharacter:   s.Location.Range.End.Character,
		})
	}

	return &RefLocations{Refs: refs}, nil
}

func (t *translator) Symbols(ctx context.Context, r *RepoRev) (*Symbols, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{}, &respSymbol)
	defer observe(ctx, start, workspaceStart, r.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	symbols := []*lsp.SymbolInformation{}
	for i, _ := range respSymbol {
		s := respSymbol[i]
		f, err := t.resolveFile(r.Repo, r.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		s.Location.URI = f.Path
		symbols = append(symbols, &s)
	}
	return &Symbols{Symbols: symbols}, nil
}

func (t *translator) ExportedSymbols(ctx context.Context, r *RepoRev) (*ExportedSymbols, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	importPath, _ := ResolveRepoAlias(r.Repo)
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "exported " + importPath + "/...",
	}, &respSymbol)
	defer observe(ctx, start, workspaceStart, r.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	var symbols []*Symbol
	// TODO(keegancsmith) go specific
	for _, s := range respSymbol {
		f, err := t.resolveFile(r.Repo, r.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		pkgParts := strings.Split(s.ContainerName, "/")
		var unit string
		if len(pkgParts) < 3 {
			// Hack for stdlib
			unit = s.ContainerName
		} else {
			unit = strings.Join(pkgParts, "/")
		}
		symbols = append(symbols, &Symbol{
			DefSpec: DefSpec{
				Repo:     f.Repo,
				Commit:   f.Commit,
				UnitType: "GoPackage",
				Unit:     unit,
				Path:     s.Name,
			},
			Name: s.Name,
			File: f.Path,
			Kind: lspKindToSymbol(s.Kind),
		})
	}
	return &ExportedSymbols{Symbols: symbols}, nil
}

func (t *translator) lspDo(ctx context.Context, rootPath string, method string, request interface{}, result interface{}) error {
	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return err
	}
	c := jsonrpc2.NewConn(context.Background(), conn, nil)
	defer func() {
		if err := c.Close(); err != nil {
			// TODO: configurable logging
			log.Println(err)
		}
	}()

	// Extract opentracing HTTP headers for propagation across LSP borders.
	var opts []jsonrpc2.CallOption
	if span := opentracing.SpanFromContext(ctx); span != nil {
		header := make(http.Header)
		carrier := opentracing.HTTPHeadersCarrier(header)
		err := span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			return err
		}
		if len(header) > 0 {
			opts = append(opts, jsonrpc2.Meta(header))
		}
	}

	// Initialize.
	if err := c.Call(ctx, "initialize", lsp.InitializeParams{RootPath: rootPath}, nil, opts...); err != nil {
		return err
	}
	if err := c.Call(ctx, method, request, result, opts...); err != nil {
		return err
	}
	if err := c.Call(ctx, "shutdown", nil, nil, opts...); err != nil {
		return err
	}
	return nil
}

func (t *translator) resolveFile(repo, commit, uri string) (*File, error) {
	workspace := t.workspace.pathToWorkspace(repo, commit)
	return t.ResolveFile(workspace, repo, commit, uri)
}
