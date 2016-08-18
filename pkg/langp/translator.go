// Package langp implements Language Processor utilities.
package langp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
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

	// WorkDir is where workspaces are created by cloning repositories and
	// dependencies.
	WorkDir string

	// PrepareRepo is called when the language processor should clone the given
	// repository into the specified workspace at a subdirectory desired by the
	// language.
	//
	// If update is true, the given workspace is a copy of a prior workspace
	// for the same repository (at e.g. an older revision) and should be
	// updated instead of prepared from scratch (for efficiency purposes).
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareRepo func(update bool, workspace, repo, commit string) error

	// PrepareDeps is called when the language processor should prepare the
	// dependencies for the given workspace/repo/commit.
	//
	// This is where language processors should perform language-specific tasks
	// like downloading dependencies via 'go get', etc. into the workspace
	// directory.
	//
	// If update is true, the given workspace is a copy of a prior workspace
	// for the same repository (at e.g. an older revision) and should be
	// updated instead of prepared from scratch (for efficiency purposes).
	//
	// If an error is returned, it is returned directly to the person who made
	// the API request which triggered the preperation of the workspace.
	PrepareDeps func(update bool, workspace, repo, commit string) error

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
		Translator:     opts,
		preparingRepos: newPending(),
		preparingDeps:  newPending(),
	}
	mux := http.NewServeMux()
	mux.Handle("/prepare", t.handler("/prepare", t.servePrepare))
	mux.Handle("/definition", t.handler("/definition", t.serveDefinition))
	mux.Handle("/exported-symbols", t.handler("/exported-symbols", t.serveExportedSymbols))
	mux.Handle("/external-refs", t.handler("/external-refs", t.serveExternalRefs))
	mux.Handle("/position-to-defspec", t.handler("/position-to-defspec", t.servePositionToDefSpec))
	mux.Handle("/defspec-to-position", t.handler("/defspec-to-position", t.serveDefSpecToPosition))
	mux.Handle("/defspec-refs", t.handler("/defspec-refs", t.serveDefSpecRefs))
	mux.Handle("/hover", t.handler("/hover", t.serveHover))
	mux.Handle("/local-refs", t.handler("/local-refs", t.serveLocalRefs))
	return mux
}

type pending struct {
	*sync.Mutex
	m map[string]bool
}

func newPending() *pending {
	return &pending{
		Mutex: &sync.Mutex{},
		m:     make(map[string]bool),
	}
}

// acquire acquires ownership of preparing k. If k is already being prepared
// by someone else, this methods blocks until preparation of k is finished
// and handled=true is returned.
//
// When finished with preparation, done should always be called. If acquire
// did not acquire ownership, done is no-op.
func (p *pending) acquire(k string, timeout time.Duration) (wasTimeout, handled bool, done func()) {
	// If nobody is preparing k, mark ownership and return:
	p.Lock()
	if _, pending := p.m[k]; !pending {
		p.m[k] = true
		p.Unlock()
		done = func() {
			p.Lock()
			_, pending := p.m[k]
			if !pending {
				p.Unlock()
				panic("pending: done() called for non-acquired k")
			}
			delete(p.m, k)
			p.Unlock()
		}
		handled = false
		return
	}
	p.Unlock()

	// Someone is preparing k, wait for completion.
	done = func() {}
	log.Printf("preparation of k=%q already underway; waiting\n", k)
	start := time.Now()
	for {
		p.Lock()
		_, pending := p.m[k]
		p.Unlock()
		if !pending {
			handled = true
			return
		}
		if time.Since(start) > timeout {
			wasTimeout = true
			log.Printf("preparation of k=%q finished\n", k)
			return
		}
		// TODO: timeout request
		time.Sleep(1 * time.Millisecond)
	}
}

type translator struct {
	*Translator
	mux                           *http.ServeMux
	preparingRepos, preparingDeps *pending
}

type handlerFunc func(body []byte) (interface{}, error)

// handler returns an HTTP handler which serves the given method at the
// specified path.
func (t *translator) handler(path string, m handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Confirm the path because http.ServeMux is fuzzy.
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// All Language Processor methods are POST and have no query
		// parameters.
		if r.Method != "POST" || len(r.URL.Query()) > 0 {
			resp := &Error{ErrorMsg: "expected POST with no query parameters"}
			t.writeResponse(w, http.StatusBadRequest, resp, path, nil)
			return
		}

		// Handle the request.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			t.writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		resp, err := m(body)
		if err != nil {
			resp := &Error{ErrorMsg: err.Error()}
			t.writeResponse(w, http.StatusBadRequest, resp, path, body)
			return
		}
		t.writeResponse(w, http.StatusOK, resp, path, body)
	})
}

func (t *translator) writeResponse(w http.ResponseWriter, status int, v interface{}, path string, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(v)
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}
	_, err = io.Copy(w, bytes.NewReader(respBody))
	if err != nil {
		// TODO: configurable logging
		log.Println(err)
	}

	// TODO: configurable logging
	log.Printf("POST %s -> %d %s\n\tBody:     %s\n\tResponse: %s\n", path, status, http.StatusText(status), string(body), string(respBody))
}

// dirExists tells if the directory p exists or not.
func dirExists(p string) (bool, error) {
	info, err := os.Stat(p)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// pathToWorkspace returns an absolute path to the workspace for the given
// repo at a specific commit.
func (t *translator) pathToWorkspace(repo, commit string) string {
	// btrfs subvolumes/snapshots cannot be deleted due to Docker permissions,
	// so we nest the directory structure one level deeper in order to have a
	// directory which we can remove in the event of failed workspace
	// preparation, like so:
	//
	//  <WorkDir>/<Repo>/<Commit>/workspace
	//
	// Where <Commit> is the btrfs subvolume/snapshot. Additionally, the
	// workspace subdir also gives us flexibility to store more data in the
	// future so it will likely stick around regardless of btrfs.
	return filepath.Join(t.WorkDir, repo, commit, "workspace")
}

// pathToSubvolume returns an absolute path to the subvolume for the given repo
// and commit.
func (t *translator) pathToSubvolume(repo, commit string) string {
	return filepath.Join(t.WorkDir, repo, commit)
}

// pathToLatest returns an absolute path to the "latest" file, which holds the
// commit of the most recently prepared workspace for the given repo.
func (t *translator) pathToLatest(repo string) string {
	return filepath.Join(t.WorkDir, repo, "latest")
}

// createWorkspace is called by prepareWorkspace and it creates the workspace
// directory as needed.
//
// This method should only ever be called when t.preparingRepos is acquired.
func (t *translator) createWorkspace(repo, commit string) (update bool, err error) {
	workspace := t.pathToWorkspace(repo, commit)
	subvolume := filepath.Join(t.WorkDir, repo, commit)

	// At this point, we know that the workspace directory doesn't exist,
	// but if the subvolume does exist then it means the workspace was
	// removed after a previous failed attempt at preparation. We can't
	// recreate the btrfs subvolume/snapshot due to Docker container
	// permissions, so to resolve this we must either prepare from scratch
	// OR copy from a previously-prepared workspace for this repo if one
	// exists.
	exists, err := dirExists(subvolume)
	if err != nil {
		return false, err
	}
	if exists {
		// Prepare the workspace from scratch.
		// TODO: Optimize this case by recursively copying an existing
		// btrfs subvolume/snapshot if one exists for this repo. Or if we
		// can solve the permission issue, just delete the subvolume to
		// really start from scratch / use a clone as we would in the
		// normal code path.
		if err := os.Mkdir(workspace, 0700); err != nil {
			return false, err
		}
		return false, err
	}

	// Create the parent directory.
	if err := os.MkdirAll(filepath.Dir(subvolume), 0700); err != nil {
		return false, err
	}

	// Determine whether or not we should create a snapshot of an
	// existing btrfs subvolume/snapshot for this repository. We simply
	// use the last-prepared commit for this repository, since that is
	// usually (but not always) the most up-to-date. This spares us of
	// some more complex commit-date comparison logic.
	latestSubvolume := t.pathToLatest(repo)
	_, err = os.Stat(latestSubvolume)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	} else if err == nil {
		// We have a recently prepared workspace, so clone and update
		// it instead of preparing a new one from scratch. We must first read
		// the symlink or else we would create a subvolume of a symlink which
		// isn't what we want (only the 'latest' file is a symlink).
		latestSubvolume, err = os.Readlink(latestSubvolume)
		if err != nil {
			return false, err
		}
		if err := btrfsSubvolumeSnapshot(latestSubvolume, subvolume); err != nil {
			return false, err
		}
		return true, nil
	}

	// We don't have a recently prepared workspace (we will be the
	// first successful one), so create a new subvolume.
	if err := btrfsSubvolumeCreate(subvolume); err != nil {
		return false, err
	}
	// Create the workspace subdirectory.
	if err := os.Mkdir(workspace, 0700); err != nil {
		return false, err
	}
	return false, nil
}

// prepareWorkspace prepares a new workspace for the given repository and
// revision.
func (t *translator) prepareWorkspace(repo, commit string) (workspace string, err error) {
	// TODO(slimsag): use a smaller timeout by default and ensure the timeout
	// error is properly handled by the frontend.
	return t.prepareWorkspaceTimeout(repo, commit, 1*time.Hour)
}

var errTimeout = errors.New("request timed out")

func (t *translator) prepareWorkspaceTimeout(repo, commit string, timeout time.Duration) (workspace string, err error) {
	start := time.Now()

	// Acquire ownership of repository preparation. Essentially this is a
	// sync.Mutex unique to the workspace.
	workspace = t.pathToWorkspace(repo, commit)
	didTimeout, handled, done := t.preparingRepos.acquire(workspace, timeout)
	if didTimeout {
		observePrepare(start, repo, prepStatusTimeout)
		return "", errTimeout
	}
	if handled {
		// A different request prepared the repository.
		observePrepare(start, repo, prepStatusWaiting)
		return workspace, nil
	}
	defer done()
	defer observePrepare(start, repo, prepStatusOK)

	// If the workspace exists already, it has been fully prepared and we don't
	// need to do anything.
	exists, err := dirExists(workspace)
	if err != nil {
		return "", err
	}
	if exists {
		return workspace, nil
	}

	// Create the workspace directory.
	update, err := t.createWorkspace(repo, commit)
	if err != nil {
		return "", err
	}

	// Prepare the workspace by creating the directory and cloning the
	// repository.
	if err := t.PrepareRepo(update, workspace, repo, commit); err != nil {
		// Preparing the workspace has failed, and thus the workspace is
		// incomplete. Remove the directory so that the next request causes
		// preparation again (this is our best chance at keeping the workspace
		// in a working state).
		log.Println("preparing workspace repo:", err)
		if err2 := os.RemoveAll(workspace); err2 != nil {
			log.Println(err2)
		}
		return "", err
	}

	// Prepare the dependencies asynchronously.
	go func() {
		// Acquire ownership of dependency preparation.
		didTimeout, handled, done = t.preparingDeps.acquire(workspace, 0*time.Second)
		if didTimeout || handled {
			// A different request is preparing the dependencies.
			return
		}
		defer done()

		if err := t.PrepareDeps(update, workspace, repo, commit); err != nil {
			// Preparing the workspace has failed, and thus the workspace is
			// incomplete. Remove the directory so that the next request causes
			// preparation again (this is our best chance at keeping the workspace
			// in a working state).
			log.Println("preparing workspace deps:", err)
			if err2 := os.RemoveAll(workspace); err2 != nil {
				log.Println(err2)
				return
			}
		}

		// We are the latest commit, so update the symlink.
		latest := t.pathToLatest(repo)
		if err := os.Remove(latest); err != nil && !os.IsNotExist(err) {
			log.Println(err)
			return
		}
		if err := os.Symlink(t.pathToSubvolume(repo, commit), latest); err != nil {
			log.Println(err)
			return
		}
	}()
	return workspace, nil
}

func (t *translator) servePrepare(body []byte) (interface{}, error) {
	// Decode the user request.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}

	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := t.prepareWorkspaceTimeout(r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return nil, err
	}
	return map[string]string{}, nil
}

func (t *translator) serveDefinition(body []byte) (interface{}, error) {
	// Decode the user request.
	var pos Position
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(pos.Repo, pos.Commit)
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
	return r, nil
}

func (t *translator) serveHover(body []byte) (interface{}, error) {
	// Decode the user request.
	var pos Position
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(pos.Repo, pos.Commit)
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

func (t *translator) serveExternalRefs(body []byte) (interface{}, error) {
	// Decode the user request.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(r.Repo, r.Commit)
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

func (t *translator) servePositionToDefSpec(body []byte) (interface{}, error) {
	// Decode the user request.
	var pos Position
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(pos.Repo, pos.Commit)
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

func (t *translator) serveDefSpecToPosition(body []byte) (interface{}, error) {
	// Decode the user request.
	var defSpec DefSpec
	if err := json.Unmarshal(body, &defSpec); err != nil {
		return nil, err
	}
	if defSpec.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if defSpec.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if defSpec.Unit == "" {
		return nil, fmt.Errorf("Unit field must be set")
	}
	if defSpec.UnitType == "" {
		return nil, fmt.Errorf("UnitType field must be set")
	}
	if defSpec.Path == "" {
		return nil, fmt.Errorf("Path field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(defSpec.Repo, defSpec.Commit)
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

func (t *translator) serveDefSpecRefs(body []byte) (interface{}, error) {
	// Decode the user request.
	var defSpec DefSpec
	if err := json.Unmarshal(body, &defSpec); err != nil {
		return nil, err
	}
	if defSpec.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if defSpec.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if defSpec.Unit == "" {
		return nil, fmt.Errorf("Unit field must be set")
	}
	if defSpec.UnitType == "" {
		return nil, fmt.Errorf("UnitType field must be set")
	}
	if defSpec.Path == "" {
		return nil, fmt.Errorf("Path field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(defSpec.Repo, defSpec.Commit)
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

func (t *translator) serveExportedSymbols(body []byte) (interface{}, error) {
	// Decode the user request.
	var r RepoRev
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if r.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(r.Repo, r.Commit)
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

func (t *translator) serveLocalRefs(body []byte) (interface{}, error) {
	// Decode the user request.
	var pos Position
	if err := json.Unmarshal(body, &pos); err != nil {
		return nil, err
	}
	if pos.Repo == "" {
		return nil, fmt.Errorf("Repo field must be set")
	}
	if pos.Commit == "" {
		return nil, fmt.Errorf("Commit field must be set")
	}
	if pos.File == "" {
		return nil, fmt.Errorf("File field must be set")
	}

	// Determine the root path for the workspace and prepare it.
	rootPath, err := t.prepareWorkspace(pos.Repo, pos.Commit)
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
	workspace := t.pathToWorkspace(repo, commit)
	return t.ResolveFile(workspace, repo, commit, uri)
}

func lspKindToSymbol(kind lsp.SymbolKind) string {
	switch kind {
	case lsp.SKPackage:
		return "package"
	case lsp.SKField:
		return "field"
	case lsp.SKFunction:
		return "func"
	case lsp.SKMethod:
		return "method"
	case lsp.SKVariable:
		return "var"
	case lsp.SKClass:
		return "type"
	case lsp.SKInterface:
		return "interface"
	case lsp.SKConstant:
		return "const"
	default:
		// TODO(keegancsmith) We haven't implemented all types yet,
		// just what Go uses
		return "unknown"
	}
}

// ExpandSGPath expands the $SGPATH variable in the given string, except it
// uses ~/.sourcegraph as the default if $SGPATH is not set.
func ExpandSGPath(s string) (string, error) {
	sgpath := os.Getenv("SGPATH")
	if sgpath == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		sgpath = filepath.Join(u.HomeDir, ".sourcegraph")
	}
	return strings.Replace(s, "$SGPATH", sgpath, -1), nil
}
