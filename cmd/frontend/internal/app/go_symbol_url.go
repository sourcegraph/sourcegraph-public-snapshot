package app

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/token"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/sourcegraph/pkg/vfsutil"
	"golang.org/x/tools/go/buildutil"

	"github.com/sourcegraph/go-lsp"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/pkg/golangserverutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/gosrc"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
)

// serveGoSymbolURL handles Go symbol URLs (e.g.,
// https://sourcegraph.com/go/github.com/gorilla/mux/-/Vars) by
// redirecting them to the file and line/column URL of the definition.
func serveGoSymbolURL(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid symbol URL path: %q", r.URL.Path)
	}
	mode := parts[0]
	symbolID := strings.Join(parts[1:], "/")

	if mode != "go" {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("invalid mode (only \"go\" is supported"),
		}
	}

	//                                                        def
	//                                                    vvvvvvvvvvvv
	// http://sourcegraph.com/go/github.com/gorilla/mux/-/Router/Match
	//                           ^^^^^^^^^^^^^^^^^^^^^^   ^^^^^^ ^^^^^
	//                                 importPath      receiver? symbolname
	importPath := strings.Split(symbolID, "/-/")[0]
	def := strings.Split(symbolID, "/-/")[1]
	var symbolName string
	var receiver *string
	symbolComponents := strings.Split(def, "/")
	switch len(symbolComponents) {
	case 1:
		symbolName = symbolComponents[0]
	case 2:
		// This is a method call.
		receiver = &symbolComponents[0]
		symbolName = symbolComponents[1]
	default:
		return fmt.Errorf("invalid def %s (must have 1 or 2 path components)", def)
	}

	dir, err := gosrc.ResolveImportPath(httputil.CachingClient, importPath)
	if err != nil {
		return err
	}
	cloneURL := dir.CloneURL

	if cloneURL == "" || !strings.HasPrefix(cloneURL, "https://github.com") {
		return fmt.Errorf("non-github clone URL resolved for import path %s", importPath)
	}

	repoName := api.RepoName(strings.TrimSuffix(strings.TrimPrefix(cloneURL, "https://"), ".git"))
	repo, err := backend.Repos.GetByName(ctx, repoName)
	if err != nil {
		return err
	}
	if err := backend.Repos.RefreshIndex(ctx, repo); err != nil {
		return err
	}

	commitID, err := backend.Repos.ResolveRev(ctx, repo, "")
	if err != nil {
		return err
	}

	vfs, err := repoVFS(r.Context(), repoName, commitID)
	if err != nil {
		return err
	}

	location, err := symbolLocation(r.Context(), vfs, commitID, importPath, path.Join("/", dir.RepoPrefix, strings.TrimPrefix(dir.ImportPath, string(dir.ProjectRoot))), receiver, symbolName)
	if err != nil {
		return err
	}
	if location == nil {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("symbol not found"),
		}
	}

	uri, err := gituri.Parse(string(location.URI))
	if err != nil {
		return err
	}
	filePath := uri.Fragment
	dest := &url.URL{
		Path:     "/" + path.Join(string(repo.Name), "-/blob", filePath),
		Fragment: fmt.Sprintf("L%d:%d$references", location.Range.Start.Line+1, location.Range.Start.Character+1),
	}
	http.Redirect(w, r, dest.String(), http.StatusFound)
	return nil
}

func symbolLocation(ctx context.Context, vfs ctxvfs.FileSystem, commitID api.CommitID, importPath string, path string, receiver *string, symbol string) (*lsp.Location, error) {
	bctx := buildContextFromVFS(ctx, vfs)

	fileSet := token.NewFileSet()
	pkg, err := parseFiles(fileSet, &bctx, importPath, path)
	if err != nil {
		return nil, err
	}

	pos := (func() *token.Pos {
		docPackage := doc.New(pkg, importPath, doc.AllDecls)
		for _, docConst := range docPackage.Consts {
			for _, spec := range docConst.Decl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, ident := range valueSpec.Names {
						if ident.Name == symbol {
							return &ident.NamePos
						}
					}
				}
			}
		}
		for _, docType := range docPackage.Types {
			if receiver != nil && docType.Name == *receiver {
				for _, method := range docType.Methods {
					if method.Name == symbol {
						return &method.Decl.Name.NamePos
					}
				}
			}
			for _, fun := range docType.Funcs {
				if fun.Name == symbol {
					return &fun.Decl.Name.NamePos
				}
			}
			for _, spec := range docType.Decl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == symbol {
					return &typeSpec.Name.NamePos
				}
			}
		}
		for _, docVar := range docPackage.Vars {
			for _, spec := range docVar.Decl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, ident := range valueSpec.Names {
						if ident.Name == symbol {
							return &ident.NamePos
						}
					}
				}
			}
		}
		for _, docFunc := range docPackage.Funcs {
			if docFunc.Name == symbol {
				return &docFunc.Decl.Name.NamePos
			}
		}
		return nil
	})()

	if pos == nil {
		return nil, nil
	}

	position := fileSet.Position(*pos)
	location := lsp.Location{
		URI: lsp.DocumentURI("https://" + string(importPath) + "?" + string(commitID) + "#" + position.Filename),
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      position.Line - 1,
				Character: position.Column - 1,
			},
			End: lsp.Position{
				Line:      position.Line - 1,
				Character: position.Column - 1,
			},
		},
	}

	return &location, nil
}

func buildContextFromVFS(ctx context.Context, vfs ctxvfs.FileSystem) build.Context {
	bctx := build.Default
	golangserverutil.PrepareContext(&bctx, ctx, vfs)
	return bctx
}

func repoVFS(ctx context.Context, name api.RepoName, rev api.CommitID) (ctxvfs.FileSystem, error) {
	if strings.HasPrefix(string(name), "github.com/") {
		return vfsutil.NewGitHubRepoVFS(string(name), string(rev))
	}

	// Fall back to a full git clone for non-github.com repos.
	return nil, fmt.Errorf("unable to fetch repo %s (only github.com repos are supported)", name)
}

func parseFiles(fset *token.FileSet, bctx *build.Context, importPath, srcDir string) (*ast.Package, error) {
	bpkg, err := bctx.ImportDir(srcDir, 0)
	if err != nil {
		return nil, err
	}

	pkg := &ast.Package{
		Files: map[string]*ast.File{},
	}
	var errs error
	for _, file := range append(bpkg.GoFiles, bpkg.TestGoFiles...) {
		if src, err := buildutil.ParseFile(fset, bctx, nil, buildutil.JoinPath(bctx, srcDir), file, 0); err == nil {
			pkg.Name = src.Name.Name
			pkg.Files[file] = src
		} else {
			errs = multierror.Append(errs, err)
		}
	}

	return pkg, errs
}
