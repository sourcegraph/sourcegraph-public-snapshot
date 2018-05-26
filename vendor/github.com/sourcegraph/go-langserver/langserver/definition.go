package langserver

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"log"
	"path/filepath"

	"github.com/sourcegraph/go-langserver/langserver/internal/godef"
	"github.com/sourcegraph/go-langserver/langserver/internal/refs"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"

	"golang.org/x/tools/go/loader"
)

func (h *LangHandler) handleDefinition(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	if h.config.UseBinaryPkgCache {
		_, _, locs, err := h.definitionGodef(ctx, params)
		if err == godef.ErrNoIdentifierFound {
			// This is expected to happen when j2d over
			// comments/strings/whitespace/etc), just return no info.
			return []lsp.Location{}, nil
		}
		return locs, err
	}

	res, err := h.handleXDefinition(ctx, conn, req, params)
	if err != nil {
		return nil, err
	}
	locs := make([]lsp.Location, 0, len(res))
	for _, li := range res {
		locs = append(locs, li.Location)
	}
	return locs, nil
}

func (h *LangHandler) handleTypeDefinition(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	res, err := h.handleXDefinition(ctx, conn, req, params)
	if err != nil {
		return nil, err
	}
	locs := make([]lsp.Location, 0, len(res))
	for _, li := range res {
		// not everything we find a definition for also has a type definition
		if li.TypeLocation.URI != "" {
			locs = append(locs, li.TypeLocation)
		}
	}
	return locs, nil
}

var testOSToVFSPath func(osPath string) string

func (h *LangHandler) definitionGodef(ctx context.Context, params lsp.TextDocumentPositionParams) (*token.FileSet, *godef.Result, []lsp.Location, error) {
	// In the case of testing, our OS paths and VFS paths do not match. In the
	// real world, this is never the case. Give the test suite the opportunity
	// to correct the path now.
	vfsURI := params.TextDocument.URI
	if testOSToVFSPath != nil {
		vfsURI = util.PathToURI(testOSToVFSPath(util.UriToPath(vfsURI)))
	}

	// Read file contents and calculate byte offset.
	contents, err := h.readFile(ctx, vfsURI)
	if err != nil {
		return nil, nil, nil, err
	}
	// convert the path into a real path because 3rd party tools
	// might load additional code based on the file's package
	filename := util.UriToRealPath(params.TextDocument.URI)
	offset, valid, why := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, nil, nil, fmt.Errorf("invalid position: %s:%d:%d (%s)", filename, params.Position.Line, params.Position.Character, why)
	}

	// Invoke godef to determine the position of the definition.
	fset := token.NewFileSet()
	res, err := godef.Godef(fset, offset, filename, contents)
	if err != nil {
		return nil, nil, nil, err
	}
	if res.Package != nil {
		// TODO: return directory location. This right now at least matches our
		// other implementation.
		return fset, res, []lsp.Location{}, nil
	}
	loc := goRangeToLSPLocation(fset, res.Start, res.End)

	if loc.URI == "file://" {
		// TODO: builtins do not have valid URIs or locations, so we emit a
		// phony location here instead. This is better than our other
		// implementation.
		loc.URI = util.PathToURI(filepath.Join(build.Default.GOROOT, "/src/builtin/builtin.go"))
		loc.Range = lsp.Range{}
	}

	return fset, res, []lsp.Location{loc}, nil
}

type foundNode struct {
	ident *ast.Ident      // the lookup in Uses[] or Defs[]
	typ   *types.TypeName // the object for a named type, if present
}

type dereferencable interface {
	Elem() types.Type
}

// typeLookup looks for a named type, but will search through
// any number of type qualifiers (chan/array/slice/pointer)
// which have an unambiguous base type. If no named type is
// found, we are not interested, because this is only used
// for finding a type's definition.
func typeLookup(prog *loader.Program, typ types.Type) *types.TypeName {
	if typ == nil {
		return nil
	}
	for {
		switch t := typ.(type) {
		case *types.Named:
			return t.Obj()
		case *types.Map:
			return nil
		case dereferencable:
			typ = t.Elem()
		default:
			return nil
		}
	}
}

func (h *LangHandler) handleXDefinition(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]symbolLocationInformation, error) {
	if !util.IsURI(params.TextDocument.URI) {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: fmt.Sprintf("%s not yet supported for out-of-workspace URI (%q)", req.Method, params.TextDocument.URI),
		}
	}

	rootPath := h.FilePath(h.init.Root())
	bctx := h.BuildContext(ctx)

	fset, node, pathEnclosingInterval, prog, pkg, _, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no locations.
		if _, ok := err.(*invalidNodeError); ok {
			return []symbolLocationInformation{}, nil
		}
		return nil, err
	}

	var nodes []foundNode
	obj, ok := pkg.Uses[node]
	if !ok {
		obj, ok = pkg.Defs[node]
	}
	if ok && obj != nil {
		if p := obj.Pos(); p.IsValid() {
			nodes = append(nodes, foundNode{
				ident: &ast.Ident{NamePos: p, Name: obj.Name()},
				typ:   typeLookup(prog, pkg.TypeOf(node)),
			})
		} else {
			// Builtins have an invalid Pos. Just don't emit a definition for
			// them, for now. It's not that valuable to jump to their def.
			//
			// TODO(sqs): find a way to actually emit builtin locations
			// (pointing to builtin/builtin.go).
			return []symbolLocationInformation{}, nil
		}
	}
	if len(nodes) == 0 {
		return nil, errors.New("definition not found")
	}
	findPackage := h.getFindPackageFunc()
	locs := make([]symbolLocationInformation, 0, len(nodes))
	for _, found := range nodes {
		// Determine location information for the node.
		l := symbolLocationInformation{
			Location: goRangeToLSPLocation(fset, found.ident.Pos(), found.ident.End()),
		}
		if found.typ != nil {
			// We don't get an end position, but we can assume it's comparable to
			// the length of the name, I hope.
			l.TypeLocation = goRangeToLSPLocation(fset, found.typ.Pos(), token.Pos(int(found.typ.Pos())+len(found.typ.Name())))
		}

		// Determine metadata information for the node.
		if def, err := refs.DefInfo(pkg.Pkg, &pkg.Info, pathEnclosingInterval, found.ident.Pos()); err == nil {
			symDesc, err := defSymbolDescriptor(ctx, bctx, rootPath, *def, findPackage)
			if err != nil {
				// TODO: tracing
				log.Println("refs.DefInfo:", err)
			} else {
				l.Symbol = symDesc
			}
		} else {
			// TODO: tracing
			log.Println("refs.DefInfo:", err)
		}
		locs = append(locs, l)
	}
	return locs, nil
}
