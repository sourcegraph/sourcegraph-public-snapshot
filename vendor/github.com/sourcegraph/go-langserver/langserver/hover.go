package langserver

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	doc "github.com/slimsag/godocmd"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleHover(ctx context.Context, conn JSONRPC2Conn, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fset, node, prog, pkg, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no information.
		if _, ok := err.(*invalidNodeError); ok {
			return nil, nil
		}
		return nil, err
	}

	o := pkg.ObjectOf(node)
	t := pkg.TypeOf(node)
	if o == nil && t == nil {
		comments := packageDoc(pkg.Files, node.Name)

		// Package statement idents don't have an object, so try that separately.
		if pkgName := packageStatementName(fset, pkg.Files, node); pkgName != "" {
			return &lsp.Hover{
				Contents: maybeAddComments(comments, []lsp.MarkedString{{Language: "go", Value: "package " + pkgName}}),
				Range:    rangeForNode(fset, node),
			}, nil
		}
		return nil, fmt.Errorf("type/object not found at %+v", params.Position)
	}

	// Don't package-qualify the string output.
	qf := func(*types.Package) string { return "" }

	var s string
	if f, ok := o.(*types.Var); ok && f.IsField() {
		// TODO(sqs): make this be like (T).F not "struct field F string".
		s = "struct " + o.String()
	} else if o != nil {
		if obj, ok := o.(*types.TypeName); ok {
			typ := obj.Type().Underlying()
			if _, ok := typ.(*types.Struct); ok {
				s = "type " + obj.Name() + " struct"
			}
		}
		if s == "" {
			s = types.ObjectString(o, qf)
		}
	} else if t != nil {
		s = types.TypeString(t, qf)
	}

	findComments := func(o types.Object) string {
		// Package names must be resolved specially, so do this now to avoid
		// additional overhead.
		if v, ok := o.(*types.PkgName); ok {
			return packageDoc(prog.Package(v.Imported().Path()).Files, node.Name)
		}

		// Resolve the object o into its respective ast.Node
		_, path, _ := prog.PathEnclosingInterval(o.Pos(), o.Pos())
		if path == nil {
			return ""
		}

		// Pull the comment out of the comment map for the file.
		f := path[len(path)-1].(*ast.File)
		cmap := ast.NewCommentMap(fset, f, f.Comments)
		pathIndex := 1
		switch v := o.(type) {
		case *types.Var:
			if !v.IsField() {
				pathIndex = 2
			}
		case *types.TypeName:
			pathIndex = 2
		}
		return commentsToText(cmap[path[pathIndex]])
	}

	comments := findComments(o)
	return &lsp.Hover{
		Contents: maybeAddComments(comments, []lsp.MarkedString{{Language: "go", Value: s}}),
		Range:    rangeForNode(fset, node),
	}, nil
}

// packageStatementName returns the package name ((*ast.Ident).Name)
// of node iff node is the package statement of a file ("package p").
func packageStatementName(fset *token.FileSet, files []*ast.File, node *ast.Ident) string {
	for _, f := range files {
		if f.Name == node {
			return node.Name
		}
	}
	return ""
}

// maybeAddComments appends the specified comments converted to Markdown godoc
// form to the specified contents slice, if the comments string is not empty.
func maybeAddComments(comments string, contents []lsp.MarkedString) []lsp.MarkedString {
	if comments == "" {
		return contents
	}
	var b bytes.Buffer
	doc.ToMarkdown(&b, comments, nil)
	return append(contents, lsp.MarkedString{
		Language: "markdown",
		Value:    b.String(),
	})
}

// packageDoc finds the documentation for the named package from its files or
// additional files.
func packageDoc(files []*ast.File, pkgName string) string {
	for _, f := range files {
		if f.Name.Name == pkgName {
			txt := f.Doc.Text()
			if strings.TrimSpace(txt) != "" {
				return txt
			}
		}
	}
	return ""
}

// commentsToText converts a slice of []*ast.CommentGroup to a flat string,
// ensuring whitespace-only comment groups are dropped.
func commentsToText(cgroups []*ast.CommentGroup) (text string) {
	for _, c := range cgroups {
		if strings.TrimSpace(c.Text()) != "" {
			text += c.Text()
		}
	}
	return text
}
