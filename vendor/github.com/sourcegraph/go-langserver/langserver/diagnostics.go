package langserver

import (
	"context"
	"fmt"
	"go/scanner"
	"go/token"
	"go/types"
	"strings"
	"sync"

	"golang.org/x/tools/go/loader"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/sourcegraph/go-langserver/langserver/util"
)

type diagnostics map[string][]*lsp.Diagnostic // map of URI to diagnostics (for PublishDiagnosticParams)

type diagnosticsCache struct {
	mu    sync.Mutex
	cache diagnostics
}

// update the cached diagnostics. In order to keep the cache in good shape it
// is required that only one go routine is able to modify the cache at a time.
func (p *diagnosticsCache) update(fn func(diagnostics) diagnostics) {
	p.mu.Lock()
	if p.cache == nil {
		p.cache = diagnostics{}
	}
	p.cache = fn(p.cache)
	p.mu.Unlock()
}

func newDiagnosticsCache() *diagnosticsCache {
	return &diagnosticsCache{
		cache: diagnostics{},
	}
}

// publishDiagnostics sends diagnostic information (such as compile
// errors) to the client.
func (h *LangHandler) publishDiagnostics(ctx context.Context, conn jsonrpc2.JSONRPC2, diags diagnostics, files []string) error {
	if !h.config.DiagnosticsEnabled {
		return nil
	}

	if diags == nil {
		diags = diagnostics{}
	}

	h.diagnosticsCache.update(func(cached diagnostics) diagnostics {
		return updateCachedDiagnostics(cached, diags, files)
	})

	for filename, diags := range diags {
		params := lsp.PublishDiagnosticsParams{
			URI:         util.PathToURI(filename),
			Diagnostics: make([]lsp.Diagnostic, len(diags)),
		}
		for i, d := range diags {
			params.Diagnostics[i] = *d
		}
		if err := conn.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
			return err
		}
	}
	return nil
}

func updateCachedDiagnostics(cachedDiagnostics diagnostics, newDiagnostics diagnostics, files []string) diagnostics {
	// add/update existing diagnostics
	for file, diags := range newDiagnostics {
		cachedDiagnostics[file] = diags
	}

	// remove all cached diagnostics for each files that is not present in `newDiagnostics`
	// and add an entry to the output diagnostics for them (to clean the clients)
	for _, file := range files {
		if _, ok := newDiagnostics[file]; !ok {
			if _, ok := cachedDiagnostics[file]; ok {
				delete(cachedDiagnostics, file)
				newDiagnostics[file] = nil
			}
		}
	}

	return cachedDiagnostics
}

func errsToDiagnostics(typeErrs []error, prog *loader.Program) (diagnostics, error) {
	var diags diagnostics
	for _, typeErr := range typeErrs {
		var (
			p    token.Position
			pEnd token.Position
			msg  string
		)
		switch e := typeErr.(type) {
		case types.Error:
			p = e.Fset.Position(e.Pos)
			_, path, _ := prog.PathEnclosingInterval(e.Pos, e.Pos)
			if len(path) > 0 {
				pEnd = e.Fset.Position(path[0].End())
			}
			msg = e.Msg
		case scanner.Error:
			p = e.Pos
			msg = e.Msg
		case scanner.ErrorList:
			if len(e) == 0 {
				continue
			}
			p = e[0].Pos
			msg = e[0].Msg
			if len(e) > 1 {
				msg = fmt.Sprintf("%s (and %d more errors)", msg, len(e)-1)
			}
		default:
			return nil, fmt.Errorf("unexpected type error: %#+v", typeErr)
		}
		// LSP is 0-indexed, so subtract one from the numbers Go reports.
		start := lsp.Position{Line: p.Line - 1, Character: p.Column - 1}
		end := lsp.Position{Line: pEnd.Line - 1, Character: pEnd.Column - 1}
		if !pEnd.IsValid() {
			end = start
		}
		diag := &lsp.Diagnostic{
			Range: lsp.Range{
				Start: start,
				End:   end,
			},
			Severity: lsp.Error,
			Source:   "go",
			Message:  strings.TrimSpace(msg),
		}
		if diags == nil {
			diags = diagnostics{}
		}
		diags[p.Filename] = append(diags[p.Filename], diag)
	}
	return diags, nil
}
