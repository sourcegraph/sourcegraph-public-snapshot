package langp

import (
	"context"
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

func (p *proxy) Prepare(ctx context.Context, r *RepoRev) error {
	// Prepare the workspace, and timeout immediately if someone else is
	// already preparing it.
	_, err := p.workspace.PrepareTimeout(ctx, r.Repo, r.Commit, 0*time.Second)
	if err != nil && err != errTimeout {
		return err
	}
	return nil
}

func (p *proxy) DefSpecToPosition(ctx context.Context, defSpec *DefSpec) (*Position, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.DefSpecToPosition(ctx, defSpec)
}

func (p *proxy) PositionToDefSpec(ctx context.Context, pos *Position) (*DefSpec, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.PositionToDefSpec(ctx, pos)
}

func (p *proxy) Definition(ctx context.Context, pos *Position) (*Range, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.Definition(ctx, pos)
}

func (p *proxy) Hover(ctx context.Context, pos *Position) (*Hover, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.Hover(ctx, pos)
}

func (p *proxy) LocalRefs(ctx context.Context, pos *Position) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, pos.Repo, pos.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.LocalRefs(ctx, pos)
}

func (p *proxy) ExternalRefs(ctx context.Context, r *RepoRev) (*ExternalRefs, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.ExternalRefs(ctx, r)
}

func (p *proxy) DefSpecRefs(ctx context.Context, defSpec *DefSpec) (*RefLocations, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, defSpec.Repo, defSpec.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.DefSpecRefs(ctx, defSpec)
}

func (p *proxy) ExportedSymbols(ctx context.Context, r *RepoRev) (*ExportedSymbols, error) {
	// Determine the root path for the workspace and prepare it.
	_, err := p.workspace.Prepare(ctx, r.Repo, r.Commit)
	if err != nil {
		return nil, err
	}

	return p.Client.ExportedSymbols(ctx, r)
}
