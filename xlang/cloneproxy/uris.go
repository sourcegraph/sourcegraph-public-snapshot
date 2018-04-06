package main

import (
	"fmt"
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
	fs := &remoteFS{conn: p.client}
	err := fs.Clone(p.ctx, p.workspaceCacheDir())
	if err != nil {
		return errors.Wrap(err, "failed to clone workspace to local cache")
	}

	log.Printf("Cloned workspace to %s", p.workspaceCacheDir())
	return nil
}

func (p *cloneProxy) cleanWorkspaceCache() error {
	log.Printf("Removing workspace cache from %s", p.workspaceCacheDir())
	return os.RemoveAll(p.workspaceCacheDir())
}

func (p *cloneProxy) workspaceCacheDir() string {
	return filepath.Join(*cacheDir, p.sessionID.String())
}

func clientToServerURI(uri lsp.DocumentURI, cacheDir string) lsp.DocumentURI {
	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("clientToServerURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if !probablyFileURI(parsedURI) {
		return uri
	}

	// We assume that any path provided by the client to the server
	// is a project path that is relative to '/'
	parsedURI.Path = filepath.Join(cacheDir, parsedURI.Path)
	return lsp.DocumentURI(parsedURI.String())
}

func serverToClientURI(uri lsp.DocumentURI, cacheDir string) lsp.DocumentURI {
	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("serverToClientURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if !probablyFileURI(parsedURI) {
		return uri
	}

	// Only rewrite uris that point to a location in the workspace cache. If it does
	// point to a cache location, then we assume that the path points to a location in the
	// project.
	if pathHasPrefix(parsedURI.Path, cacheDir) {
		parsedURI.Path = filepath.Join("/", pathTrimPrefix(parsedURI.Path, cacheDir))
	}

	return lsp.DocumentURI(parsedURI.String())
}

func probablyFileURI(candidate *url.URL) bool {
	if !(candidate.Scheme == "" || candidate.Scheme == "file") {
		return false
	}

	if candidate.Path == "" {
		return false
	}

	return true
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
// If update is non-nil, it updates all document URIs in an LSP
// params/result with the value of f(existingURI). Callers can use
// this to rewrite paths in the params/result.
//
// TODO(sqs): does not support WorkspaceEdit (with a field whose
// TypeScript type is {[uri: string]: TextEdit[]}.
func WalkURIFields(o interface{}, update func(lsp.DocumentURI) lsp.DocumentURI) {
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
