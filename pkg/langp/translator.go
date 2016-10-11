// Package langp implements Language Processor utilities.
package langp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"
)

// Translator is an HTTP handler which translates from the Language Processor
// REST API (defined in proto.go) directly to Sourcegraph LSP batch requests.
type Translator struct {
	// Addr is the address of the LSP server which translation should occur
	// against.
	Addr string

	// Preparer is the workspace preparer that will be used.
	Preparer *Preparer

	// SymbolsTranslator is an optional SymbolsTranslator to use.
	SymbolsTranslator *SymbolsTranslator

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
	ResolveFile func(ctx context.Context, workspace, repo, commit, uri string) (*File, error)

	// FileURI, if non-nil, is called to form the file URI which is sent to a
	// language server. Provided is the repo and commit, and a file URI which
	// is relative to the repository root.
	//
	// FileURI should return a file URI which is relative to the previously
	// prepared workspace directory.
	FileURI func(ctx context.Context, repo, commit, file string) string
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

// SymbolsTranslator is a translator around our various uses of
// workspace/symbols.
type SymbolsTranslator struct {
	ExportedSymbolsQuery func(*RepoRev) string
	ExportedSymbol       func(*RepoRev, *File, *lsp.SymbolInformation) *Symbol
	ExternalRefsQuery    func(*RepoRev) string
	ExternalRef          func(*RepoRev, *File, *lsp.SymbolInformation) *Ref
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
		p.TextDocument.URI = t.FileURI(ctx, pos.Repo, pos.Commit, pos.File)
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
		f, err := t.resolveFile(ctx, pos.Repo, pos.Commit, respDef[0].URI)
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
		p.TextDocument.URI = t.FileURI(ctx, pos.Repo, pos.Commit, pos.File)
	}

	var respHover lsp.Hover
	err = t.lspDo(ctx, rootPath, "textDocument/hover", p, &respHover)
	defer observe(ctx, start, workspaceStart, pos.Repo, err, err == nil && len(respHover.Contents) == 0)
	if err != nil {
		return nil, err
	}

	// Unresolved
	if len(respHover.Contents) == 0 {
		return &Hover{
			Unresolved: true,
		}, nil
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
		f, err := t.resolveFile(ctx, pos.Repo, pos.Commit, defInfo.URI)
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
		p.TextDocument.URI = t.FileURI(ctx, pos.Repo, pos.Commit, pos.File)
	}

	var resp []lsp.Location
	err = t.lspDo(ctx, rootPath, "textDocument/references", p, &resp)
	defer observe(ctx, start, workspaceStart, pos.Repo, err, len(resp) == 0)
	if err != nil {
		return nil, err
	}
	refs := make([]*Range, 0, len(resp))
	for _, r := range resp {
		f, err := t.resolveFile(ctx, pos.Repo, pos.Commit, r.URI)
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
	if t.SymbolsTranslator == nil || t.SymbolsTranslator.ExternalRefsQuery == nil || t.SymbolsTranslator.ExternalRef == nil {
		return nil, errors.New("ExternalRefs is not supported")
	}

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
	err = t.lspDo(ctx, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{
		Query: t.SymbolsTranslator.ExternalRefsQuery(r),
	}, &respSymbol)
	defer observe(ctx, start, workspaceStart, r.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	refs := make([]*Ref, 0, len(respSymbol))
	for _, s := range respSymbol {
		f, err := t.resolveFile(ctx, r.Repo, r.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		refs = append(refs, t.SymbolsTranslator.ExternalRef(r, f, &s))
	}
	return &ExternalRefs{Refs: refs}, nil
}

func (t *translator) Symbols(ctx context.Context, opt *SymbolsQuery) (*Symbols, error) {
	// Determine the root path for the workspace and prepare it.
	workspaceStart := time.Now()
	rootPath, err := t.workspace.Prepare(ctx, opt.Repo, opt.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(ctx, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{Query: opt.Query}, &respSymbol)
	defer observe(ctx, start, workspaceStart, opt.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	symbols := []*lsp.SymbolInformation{}
	for i, _ := range respSymbol {
		s := respSymbol[i]
		f, err := t.resolveFile(ctx, opt.Repo, opt.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		s.Location.URI = f.Path
		symbols = append(symbols, &s)
	}
	return &Symbols{Symbols: symbols}, nil
}

func (t *translator) ExportedSymbols(ctx context.Context, r *RepoRev) (*ExportedSymbols, error) {
	if t.SymbolsTranslator == nil || t.SymbolsTranslator.ExportedSymbolsQuery == nil || t.SymbolsTranslator.ExportedSymbol == nil {
		return nil, errors.New("ExportedSymbols is not supported")
	}

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
	err = t.lspDo(ctx, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{
		Query: t.SymbolsTranslator.ExportedSymbolsQuery(r),
	}, &respSymbol)
	defer observe(ctx, start, workspaceStart, r.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	symbols := make([]*Symbol, 0, len(respSymbol))
	for _, s := range respSymbol {
		f, err := t.resolveFile(ctx, r.Repo, r.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, t.SymbolsTranslator.ExportedSymbol(r, f, &s))
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

func (t *translator) resolveFile(ctx context.Context, repo, commit, uri string) (*File, error) {
	workspace := t.workspace.pathToWorkspace(repo, commit)
	return t.ResolveFile(ctx, workspace, repo, commit, uri)
}
