package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/pkg/errors"
)

func (p *cloneProxy) cloneWorkspaceToCache() error {
	fs := &remoteFS{client: p.client}

	filePaths, err := fs.Walk(p.ctx, "/")
	if err != nil {
		return errors.Wrap(err, "failed to fetch all filePaths when cloning workspace")
	}

	files, err := fs.BatchOpen(p.ctx, filePaths)
	if err != nil {
		return errors.Wrap(err, "failed to batch open files when cloning workspace")
	}

	for _, file := range files {
		cacheFilePath := filepath.Join(p.workspaceCacheDir(), file.path)
		cacheFileDir := filepath.Dir(cacheFilePath)

		if err := os.MkdirAll(cacheFileDir, os.ModePerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to make parent dirs for %s", cacheFilePath))
		}

		if err := ioutil.WriteFile(cacheFilePath, []byte(file.content), os.ModePerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to write file content for %s", cacheFilePath))
		}
	}

	log.Printf("Cloned workspace to %s", p.workspaceCacheDir())
	return nil
}

func (p *cloneProxy) workspaceCacheDir() string {
	return filepath.Join(*cacheDir, p.sessionID.String())
}

func (p *cloneProxy) clientToServerURI(uri lsp.DocumentURI) lsp.DocumentURI {
	if uri == "" {
		return uri
	}

	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("clientToServerURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if parsedURI.Scheme != "" && parsedURI.Scheme != "file" {
		return uri
	}

	parsedURI.Path = filepath.Join(p.workspaceCacheDir(), parsedURI.Path)
	return lsp.DocumentURI(parsedURI.String())
}

func (p *cloneProxy) serverToClientURI(uri lsp.DocumentURI) lsp.DocumentURI {
	if uri == "" {
		return uri
	}

	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("serverToClientURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if parsedURI.Scheme != "" && parsedURI.Scheme != "file" {
		return uri
	}

	if pathHasPrefix(parsedURI.Path, p.workspaceCacheDir()) {
		parsedURI.Path = filepath.Join("/", pathTrimPrefix(parsedURI.Path, p.workspaceCacheDir()))
	}

	return lsp.DocumentURI(parsedURI.String())
}

// copied from sourcegraph/go-langserver/util.go
func pathHasPrefix(s, prefix string) bool {
	var prefixSlash string
	if prefix != "" && !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefixSlash = prefix + string(os.PathSeparator)
	}
	return s == prefix || strings.HasPrefix(s, prefixSlash)
}

// copied from sourcegraph/go-langserver/util.go
func pathTrimPrefix(s, prefix string) string {
	if s == prefix {
		return ""
	}
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	return strings.TrimPrefix(s, prefix)
}

// WalkURIFields walks the LSP params/result object for fields
// containing document URIs.
//
// If collect is non-nil, it calls collect(uri) for every URI
// encountered. Callers can use this to collect a list of all document
// URIs referenced in the params/result.
//
// If update is non-nil, it updates all document URIs in an LSP
// params/result with the value of f(existingURI). Callers can use
// this to rewrite paths in the params/result.
//
// TODO(sqs): does not support WorkspaceEdit (with a field whose
// TypeScript type is {[uri: string]: TextEdit[]}.
func WalkURIFields(o interface{}, collect func(lsp.DocumentURI), update func(lsp.DocumentURI) lsp.DocumentURI) {
	var walk func(o interface{})
	walk = func(o interface{}) {
		switch o := o.(type) {
		case map[string]interface{}:
			for k, v := range o { // Location, TextDocumentIdentifier, TextDocumentItem, etc.
				// Handling "rootPath" and "rootUri" special cases the initalize method.
				if k == "uri" || k == "rootPath" || k == "rootUri" {
					s, ok := v.(string)
					if !ok {
						s2, ok2 := v.(lsp.DocumentURI)
						s = string(s2)
						ok = ok2
					}
					if ok {
						if collect != nil {
							collect(lsp.DocumentURI(s))
						}
						if update != nil {
							o[k] = update(lsp.DocumentURI(s))
						}
						continue
					}
				}
				walk(v)
			}
		case []interface{}: // Location[]
			for _, v := range o {
				walk(v)
			}
		default: // structs with a "URI" field
			rv := reflect.ValueOf(o)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				if fv := rv.FieldByName("URI"); fv.Kind() == reflect.String {
					if collect != nil {
						collect(lsp.DocumentURI(fv.String()))
					}
					if update != nil {
						fv.SetString(string(update(lsp.DocumentURI(fv.String()))))
					}
				}
				for i := 0; i < rv.NumField(); i++ {
					fv := rv.Field(i)
					if fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Struct || fv.Kind() == reflect.Array {
						walk(fv.Interface())
					}
				}
			}
		}
	}
	walk(o)
}
