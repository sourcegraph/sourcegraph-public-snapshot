package golang

import (
	"context"
	"errors"
	"io/ioutil"

	"github.com/sourcegraph/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Handler) handleDefinition(ctx context.Context, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	// Find start of definition using guru
	contents, err := h.readFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}
	def, err := godef(ctx, h.goEnv(), h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}

	// No position information, but we didn't fail. Assume this is a valid
	// place to click, but we are looking at a string literal/comment/etc.
	if def.Position.Path == "" {
		return []lsp.Location{}, nil
	}

	uri, err := h.fileURI(def.Position.Path)
	if err != nil {
		return nil, err
	}

	if def.Position.IsDir {
		return []lsp.Location{lsp.Location{
			URI: uri,
		}}, nil
	}

	if uri != params.TextDocument.URI {
		// different file to input
		contents, err = ioutil.ReadFile(def.Position.Path)
		if err != nil {
			return nil, err
		}
	}
	r, err := rangeAtPosition(lsp.Position{Line: def.Position.Line - 1, Character: def.Position.Column - 1}, contents)
	if err != nil {
		return nil, err
	}

	var locs []lsp.Location
	locs = append(locs, lsp.Location{
		URI:   uri,
		Range: r,
	})
	return locs, nil
}
