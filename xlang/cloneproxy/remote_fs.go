package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
)

type remoteFS struct {
	conn *jsonrpc2.Conn
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

type batchOpenResult struct {
	path    string
	content string
}

// Open returns the content of the text file for the given path.
func (fs *remoteFS) Open(ctx context.Context, path string) (string, error) {
	u := &url.URL{
		Scheme: "file",
		Path:   path,
	}
	params := lspext.ContentParams{TextDocument: lsp.TextDocumentIdentifier{URI: lsp.DocumentURI(u.String())}}
	var res lsp.TextDocumentItem

	if err := fs.conn.Call(ctx, "textDocument/xcontent", params, &res); err != nil {
		return "", errors.Wrap(err, "calling textDocument/xcontent failed")
	}

	return res.Text, nil
}

// Walk returns a list of all file paths that are children of "base".
func (fs *remoteFS) Walk(ctx context.Context, base string) ([]string, error) {
	params := lspext.FilesParams{Base: base}
	var res []lsp.TextDocumentIdentifier

	if err := fs.conn.Call(ctx, "workspace/xfiles", &params, &res); err != nil {
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

func (fs *remoteFS) Clone(ctx context.Context, baseDir string) error {
	filePaths, err := fs.Walk(ctx, "/")
	if err != nil {
		return errors.Wrap(err, "failed to fetch all filePaths during clone")
	}

	files, err := fs.BatchOpen(ctx, filePaths)
	if err != nil {
		return errors.Wrap(err, "failed to batch open files during clone")
	}

	for _, file := range files {
		newFilePath := filepath.Join(baseDir, file.path)

		// There is an assumption here that all paths returned from Walk()
		// point to files, not directories
		parentDir := filepath.Dir(newFilePath)

		if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to make parent dirs for %s")
		}

		if err := ioutil.WriteFile(newFilePath, []byte(file.content), os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to write file content for %s", newFilePath)
		}
	}
	return nil
}
