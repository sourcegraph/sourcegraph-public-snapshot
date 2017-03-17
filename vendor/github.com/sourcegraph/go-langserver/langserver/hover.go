package langserver

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"strings"

	doc "github.com/slimsag/godocmd"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LangHandler) handleHover(ctx context.Context, conn jsonrpc2.JSONRPC2, req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	fset, node, _, prog, pkg, _, err := h.typecheck(ctx, conn, params.TextDocument.URI, params.Position)
	if err != nil {
		// Invalid nodes means we tried to click on something which is
		// not an ident (eg comment/string/etc). Return no information.
		if _, ok := err.(*invalidNodeError); ok {
			return nil, nil
		}
		// This is a common error we get in production when a user is
		// browsing a go pkg which only contains files we can't
		// analyse (usually due to build tags). To reduce signal of
		// actual bad errors, we return no error in this case.
		if _, ok := err.(*build.NoGoError); ok {
			return nil, nil
		}
		return nil, err
	}

	o := pkg.ObjectOf(node)
	t := pkg.TypeOf(node)
	if o == nil && t == nil {
		comments := packageDoc(pkg.Files, node.Name)

		// Package statement idents don't have an object, so try that separately.
		r := rangeForNode(fset, node)
		if pkgName := packageStatementName(fset, pkg.Files, node); pkgName != "" {
			return &lsp.Hover{
				Contents: maybeAddComments(comments, []lsp.MarkedString{{Language: "go", Value: "package " + pkgName}}),
				Range:    &r,
			}, nil
		}
		return nil, fmt.Errorf("type/object not found at %+v", params.Position)
	}
	if o != nil && !o.Pos().IsValid() {
		// Only builtins have invalid position, and don't have useful info.
		return nil, nil
	}
	// Don't package-qualify the string output.
	qf := func(*types.Package) string { return "" }

	var s string
	var extra string
	if f, ok := o.(*types.Var); ok && f.IsField() {
		// TODO(sqs): make this be like (T).F not "struct field F string".
		s = "struct " + o.String()
	} else if o != nil {
		if obj, ok := o.(*types.TypeName); ok {
			typ := obj.Type().Underlying()
			if _, ok := typ.(*types.Struct); ok {
				s = "type " + obj.Name() + " struct"
				extra = prettyPrintTypesString(types.TypeString(typ, qf))
			}
			if _, ok := typ.(*types.Interface); ok {
				s = "type " + obj.Name() + " interface"
				extra = prettyPrintTypesString(types.TypeString(typ, qf))
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

		// Pull the comment out of the comment map for the file. Do
		// not search too far away from the current path.
		var doc *ast.CommentGroup
		for i := 0; i < 3 && i < len(path) && doc == nil; i++ {
			switch v := path[i].(type) {
			case *ast.Field:
				doc = v.Doc
			case *ast.ValueSpec:
				doc = v.Doc
			case *ast.TypeSpec:
				doc = v.Doc
			case *ast.GenDecl:
				doc = v.Doc
			case *ast.FuncDecl:
				doc = v.Doc
			}
		}
		if doc == nil {
			return ""
		}
		return doc.Text()
	}

	contents := maybeAddComments(findComments(o), []lsp.MarkedString{{Language: "go", Value: s}})
	if extra != "" {
		// If we have extra info, ensure it comes after the usually
		// more useful documentation
		contents = append(contents, lsp.MarkedString{Language: "go", Value: extra})
	}

	r := rangeForNode(fset, node)
	return &lsp.Hover{
		Contents: contents,
		Range:    &r,
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

// prettyPrintTypesString is pretty printing specific to the output of
// types.*String. Instead of re-implementing the printer, we can just
// transform its output.
func prettyPrintTypesString(s string) string {
	// Don't bother including the fields if it is empty
	if strings.HasSuffix(s, "{}") {
		return ""
	}
	var b bytes.Buffer
	b.Grow(len(s))
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ';':
			b.WriteByte('\n')
			for j := 0; j < depth; j++ {
				b.WriteString("    ")
			}
			// Skip following space
			i++

		case '{':
			if i == len(s)-1 {
				// This should never happen, but in case it
				// does give up
				return s
			}

			n := s[i+1]
			if n == '}' {
				// Do not modify {}
				b.WriteString("{}")
				// We have already written }, so skip
				i++
			} else {
				// We expect fields to follow, insert a newline and space
				depth++
				b.WriteString(" {\n")
				for j := 0; j < depth; j++ {
					b.WriteString("    ")
				}
			}

		case '}':
			depth--
			if depth < 0 {
				return s
			}
			b.WriteString("\n}")

		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
