package main

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

type remoteFS struct {
	client *jsonrpc2.Conn
}

// BatchOpen opens all of the content for the specified paths.
func (fs *remoteFS) BatchOpen(ctx context.Context, paths []string) ([]batchOpenResult, error) {
	par := parallel.NewRun(8)

	var mut sync.Mutex
	var results []batchOpenResult

	for _, path := range paths {
		par.Acquire()

		go func(path string) {
			defer par.Release()

			text, err := fs.Open(ctx, path)
			if err != nil {
				par.Error(err)
				return
			}

			mut.Lock()
			defer mut.Unlock()

			results = append(results, batchOpenResult{path: path, content: text})

		}(path)
	}

	if err := par.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// Open returns the content of the text file for the given path.
func (fs *remoteFS) Open(ctx context.Context, path string) (string, error) {
	u := &url.URL{
		Scheme: "file",
		Path:   path,
	}
	params := lspext.ContentParams{TextDocument: lsp.TextDocumentIdentifier{URI: lsp.DocumentURI(u.String())}}
	var res lsp.TextDocumentItem

	if err := fs.client.Call(ctx, "textDocument/xcontent", params, &res); err != nil {
		return "", errors.Wrap(err, "calling textDocument/xcontent failed")
	}

	return res.Text, nil
}

// Walk returns a list of all file paths that are children of "base".
func (fs *remoteFS) Walk(ctx context.Context, base string) ([]string, error) {
	params := lspext.FilesParams{Base: base}
	var res []lsp.TextDocumentIdentifier

	if err := fs.client.Call(ctx, "workspace/xfiles", &params, &res); err != nil {
		return nil, errors.Wrap(err, "calling workspace/xfiles failed")
	}

	var paths []string
	var parseErrors *multierror.Error
	for _, ident := range res {
		parsedURI, err := url.Parse(string(ident.URI))
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("when parsing rawURI %s for walk", ident.URI))
			parseErrors = multierror.Append(parseErrors, err)
		} else {
			paths = append(paths, parsedURI.Path)
		}
	}

	return paths, parseErrors.ErrorOrNil()
}

type batchOpenResult struct {
	path    string
	content string
}
