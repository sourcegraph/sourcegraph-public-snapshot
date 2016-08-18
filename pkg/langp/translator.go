// Package langp implements Language Processor utilities.
package langp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

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
		preparer:   opts.Preparer,
	}
	return NewServer(t)
}

type translator struct {
	*Translator
	mux      *http.ServeMux
	preparer *Preparer
}

func (t *translator) Prepare(r *RepoRev) error {
	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := t.preparer.prepareWorkspaceTimeout(r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return err
	}
	return nil
}

func (t *translator) Definition(pos *Position) (*Range, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	reqDef := &jsonrpc2.Request{
		Method: "textDocument/definition",
	}
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	if err := reqDef.SetParams(p); err != nil {
		return nil, err
	}

	// TODO: according to spec this could be lsp.Location OR []lsp.Location
	var respDef []lsp.Location
	err = t.lspDo(rootPath, reqDef, &respDef)
	defer observe(start, reqDef.Method, pos.Repo, err, len(respDef) == 0)
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

func (t *translator) Hover(pos *Position) (*Hover, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking hover,
	// but good enough for now.
	reqHover := &jsonrpc2.Request{
		Method: "textDocument/hover",
	}
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	if err := reqHover.SetParams(p); err != nil {
		return nil, err
	}

	var respHover lsp.Hover
	err = t.lspDo(rootPath, reqHover, &respHover)
	defer observe(start, reqHover.Method, pos.Repo, err, err == nil && len(respHover.Contents) == 0)
	if err != nil {
		return nil, err
	}
	return HoverFromLSP(&respHover), nil
}

func (t *translator) ExternalRefs(r *RepoRev) (*ExternalRefs, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbol := &jsonrpc2.Request{
		Method: "workspace/symbol",
	}
	importPath, _ := ResolveRepoAlias(r.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "external " + importPath + "/...",
	}
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}

	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(rootPath, reqSymbol, &respSymbol)
	defer observe(start, reqSymbol.Method, r.Repo, err, len(respSymbol) == 0)
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
			Repo:     UnresolveRepoAlias(repo),
			Commit:   commit,
			UnitType: "GoPackage",
			Unit:     unit,
			Path:     s.Name,
		})
	}
	return &ExternalRefs{Defs: defs}, nil
}

func (t *translator) PositionToDefSpec(pos *Position) (*DefSpec, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbol := &jsonrpc2.Request{
		Method: "workspace/symbol",
	}

	// Repositories are same indicates query local references.
	importPath, _ := ResolveRepoAlias(pos.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "defspec-refs-internal " + importPath + "/...",
	}
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}

	// TODO(slimsag): cache symbol information for quicker access
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(rootPath, reqSymbol, &respSymbol)
	defer observe(start, reqSymbol.Method, pos.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

	// Filter out refs that don't match our position.
	for _, s := range respSymbol {
		r := s.Location.Range
		if pos.Line < r.Start.Line || pos.Line > r.End.Line {
			continue
		}
		if pos.Character < r.Start.Character || pos.Character > r.End.Character {
			continue
		}
		f, err := t.resolveFile(pos.Repo, pos.Commit, s.Location.URI)
		if err != nil {
			return nil, err
		}
		if f.Path != pos.File {
			continue
		}

		// TODO(slimsag): how do we handle positions that map to multiple unique
		// def keys? e.g. see the results for:
		//
		//  curl -s -H "Content-Type: application/json" -X POST -d '{"Repo":"github.com/slimsag/mux","Commit":"780415097119f6f61c55475fe59b66f3c3e9ea53","File":"mux.go","Line":57,"Character":17}' http://localhost:4141/position-to-defkey | jq
		//
		// Right now we assume the first one is right, but this is likely to fail
		// in some cases? Maybe we should have a count/index as part of the key.
		pkgParts := strings.Split(s.ContainerName, "/")
		var unit string
		if len(pkgParts) < 3 {
			// Hack for stdlib
			unit = s.ContainerName
		} else {
			unit = strings.Join(pkgParts, "/")
		}
		// TODO: Go-specific
		return &DefSpec{
			Repo:     f.Repo,
			Commit:   f.Commit,
			Unit:     unit,
			UnitType: "GoPackage",
			Path:     s.Name,
		}, nil
	}
	// TODO: formalize not-found errors
	return nil, errors.New("def key for position not found")
}

func (t *translator) DefSpecToPosition(defSpec *DefSpec) (*Position, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbol := &jsonrpc2.Request{
		Method: "workspace/symbol",
	}

	importPath, _ := ResolveRepoAlias(defSpec.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "defspec-refs-internal " + importPath + "/...",
	}
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}

	// TODO(slimsag): cache symbol information for quicker access
	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(rootPath, reqSymbol, &respSymbol)
	defer observe(start, reqSymbol.Method, defSpec.Repo, err, len(respSymbol) == 0)
	if err != nil {
		return nil, err
	}

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

		// TODO(slimsag): how do we handle def keys that map to multiple identical
		// positions? e.g. see the results for:
		//
		// 	curl -s -H "Content-Type: application/json" -X POST -d '{"Repo":"github.com/slimsag/mux","Commit":"780415097119f6f61c55475fe59b66f3c3e9ea53","Def":"GoPackage/github.com/slimsag/mux/-/Router/Match"}' http://localhost:4141/defspec-to-position
		//
		// Right now we assume the first one is right, but this is likely to fail
		// in some cases? Maybe we should have a count/index as part of the key.
		f, err := t.resolveFile(defSpec.Repo, defSpec.Commit, s.Location.URI)
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

func (t *translator) DefSpecRefs(defSpec *DefSpec) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbol := &jsonrpc2.Request{
		Method: "workspace/symbol",
	}

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
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}

	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(rootPath, reqSymbol, &respSymbol)
	defer observe(start, reqSymbol.Method, defSpec.Repo, err, len(respSymbol) == 0)
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

func (t *translator) ExportedSymbols(r *RepoRev) (*ExportedSymbols, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking symbol,
	// but good enough for now.
	reqSymbol := &jsonrpc2.Request{
		Method: "workspace/symbol",
	}
	importPath, _ := ResolveRepoAlias(r.Repo)
	p := lsp.WorkspaceSymbolParams{
		// TODO(keegancsmith) this is go specific
		Query: "exported " + importPath + "/...",
	}
	if err := reqSymbol.SetParams(p); err != nil {
		return nil, err
	}

	var respSymbol []lsp.SymbolInformation
	err = t.lspDo(rootPath, reqSymbol, &respSymbol)
	defer observe(start, reqSymbol.Method, r.Repo, err, len(respSymbol) == 0)
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

func (t *translator) LocalRefs(pos *Position) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	start := time.Now()

	// TODO: should probably check server capabilities before invoking references,
	// but good enough for now.
	req := &jsonrpc2.Request{
		Method: "textDocument/references",
	}
	p := pos.LSP()
	if t.FileURI != nil {
		p.TextDocument.URI = t.FileURI(pos.Repo, pos.Commit, pos.File)
	}
	if err := req.SetParams(p); err != nil {
		return nil, err
	}

	var resp []lsp.Location
	err = t.lspDo(rootPath, req, &resp)
	defer observe(start, req.Method, pos.Repo, err, len(resp) == 0)
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

func (t *translator) lspDo(rootPath string, request *jsonrpc2.Request, result interface{}) error {
	// TODO(slimsag): We don't need to create a new JSON RPC 2 connection every
	// time, but we will need reconnection logic and a non-dumb jsonrpc2.Client
	// which can handle concurrency (according to Sourcegraph LSP spec we can
	// and should use one connection for all requests).
	conn, err := net.Dial("tcp", t.Addr)
	if err != nil {
		return err
	}
	cl := jsonrpc2.NewClient(conn)
	defer func() {
		if err := cl.Close(); err != nil {
			// TODO: configurable logging
			log.Println(err)
		}
	}()

	// Build the LSP requests.
	reqInit := &jsonrpc2.Request{
		ID:     "0",
		Method: "initialize",
	}
	if err := reqInit.SetParams(&lsp.InitializeParams{
		RootPath: rootPath,
	}); err != nil {
		return err
	}
	request.ID = "1"
	reqShutdown := &jsonrpc2.Request{ID: "2", Method: "shutdown"}

	// Make the batched LSP request.
	resps, err := cl.RequestBatchAndWaitForAllResponses(
		reqInit,
		request,
		reqShutdown,
	)
	if err != nil {
		return err
	}

	// Unmarshal the LSP responses.
	resp, ok := resps["1"]
	if !ok {
		return fmt.Errorf("response to %s request from LSP server not found", request.Method)
	}
	if resp.Error != nil {
		return resp.Error
	}
	return json.Unmarshal(*resp.Result, result)
}

func (t *translator) resolveFile(repo, commit, uri string) (*File, error) {
	workspace := t.preparer.pathToWorkspace(repo, commit)
	return t.ResolveFile(workspace, repo, commit, uri)
}
