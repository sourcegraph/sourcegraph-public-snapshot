package langp

import (
	"net/http"
	"time"
)

// TODO: add metrics along the lines of Translator

// NewProxy returns an HTTP handler which proxies Language Processor REST API
// requests to a separate server. It handles validation of the requests (e.g.
// in case required JSON fields are missing), and workspace preparation.
//
// The requests are proxied as-is to the specified language processor.
func NewProxy(client *Client, preparer *Preparer) (http.Handler, error) {
	return NewServer(&proxy{Client: client, workspace: preparer}), nil
}

type proxy struct {
	*Client
	workspace *Preparer
}

func (p *proxy) Prepare(r *RepoRev) error {
	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := p.workspace.prepareTimeout(r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return err
	}
	return nil
}

func (p *proxy) DefSpecToPosition(defSpec *DefSpec) (*Position, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.DefSpecToPosition(defSpec)
}

func (p *proxy) PositionToDefSpec(pos *Position) (*DefSpec, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.PositionToDefSpec(pos)
}

func (p *proxy) Definition(pos *Position) (*Range, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.Definition(pos)
}

func (p *proxy) Hover(pos *Position) (*Hover, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.Hover(pos)
}

func (p *proxy) LocalRefs(pos *Position) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.LocalRefs(pos)
}

func (p *proxy) ExternalRefs(r *RepoRev) (*ExternalRefs, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.ExternalRefs(r)
}

func (p *proxy) DefSpecRefs(defSpec *DefSpec) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.DefSpecRefs(defSpec)
}

func (p *proxy) ExportedSymbols(r *RepoRev) (*ExportedSymbols, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.prepare(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.ExportedSymbols(r)
}
