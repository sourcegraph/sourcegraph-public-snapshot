package langp

import (
	"errors"
	"net/http"
	"time"
)

// TODO: add metrics along the lines of Translator

// Proxy returns an HTTP handler which proxies Language Processor REST API
// requests to a separate server. It handles validation of the requests (e.g.
// in case required JSON fields are missing), and workspace preparation.
//
// The requests are proxied as-is to the specified language processor.
func Proxy(target string, preparer *Preparer) (http.Handler, error) {
	client, err := NewClient(target)
	if err != nil {
		return nil, err
	}
	return NewServer(&proxy{
		client:   client,
		preparer: preparer,
	}), nil
}

type proxy struct {
	client   *Client
	preparer *Preparer
}

func (p *proxy) Prepare(r *RepoRev) error {
	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := p.preparer.prepareWorkspaceTimeout(r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return err
	}
	return nil
}

func (p *proxy) DefSpecToPosition(defSpec *DefSpec) (*Position, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.DefSpecToPosition not implemented")
}

func (p *proxy) PositionToDefSpec(pos *Position) (*DefSpec, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.PositionToDefSpec not implemented")
}

func (p *proxy) Definition(pos *Position) (*Range, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.Definition not implemented")
}

func (p *proxy) Hover(pos *Position) (*Hover, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.Hover not implemented")
}

func (p *proxy) LocalRefs(pos *Position) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.LocalRefs not implemented")
}

func (p *proxy) ExternalRefs(r *RepoRev) (*ExternalRefs, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.ExternalRefs not implemented")
}

func (p *proxy) DefSpecRefs(defSpec *DefSpec) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.DefSpecRefs not implemented")
}

func (p *proxy) ExportedSymbols(r *RepoRev) (*ExportedSymbols, error) {
	// Determine the root path for the workspace and prepare it.
	rootPath, err := p.preparer.prepareWorkspace(r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}
	_ = rootPath // Probably communicate rootPath via a query parameter.

	// TODO: implement
	return nil, errors.New("pkg/langp.Proxy.ExportedSymbols not implemented")
}
