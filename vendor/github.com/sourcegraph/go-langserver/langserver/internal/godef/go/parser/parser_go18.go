// +build !go1.9

package parser

import "go/ast"

func parseTypeSpec(p *parser, doc *ast.CommentGroup, decl *ast.GenDecl, _ int) ast.Spec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}

	ident := p.parseIdent()
	// Go spec: The scope of a type identifier declared inside a function begins
	// at the identifier in the TypeSpec and ends at the end of the innermost
	// containing block.
	// (Global identifiers are resolved in a separate phase after parsing.)
	spec := &ast.TypeSpec{Doc: doc, Name: ident, Comment: p.lineComment}
	p.declare(spec, p.topScope, ast.Typ, ident)
	spec.Type = p.parseType()
	p.expectSemi() // call before accessing p.linecomment
	spec.Comment = p.lineComment

	return spec
}
