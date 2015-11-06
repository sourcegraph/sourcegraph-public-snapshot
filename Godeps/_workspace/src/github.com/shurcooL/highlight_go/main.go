// Package highlight_go provides a syntax highlighter for Go, using go/scanner.
package highlight_go

import (
	"go/scanner"
	"go/token"
	"io"

	"github.com/sourcegraph/annotate"
	"github.com/sourcegraph/syntaxhighlight"
)

// TokenKind returns a syntaxhighlight token kind value for the given tok and lit.
func TokenKind(tok token.Token, lit string) syntaxhighlight.Kind {
	switch {
	case tok.IsKeyword() || (tok.IsOperator() && tok <= token.ELLIPSIS):
		return syntaxhighlight.Keyword

	// Literals.
	case tok == token.INT || tok == token.FLOAT || tok == token.IMAG || tok == token.CHAR:
		return syntaxhighlight.Decimal
	case tok == token.STRING:
		return syntaxhighlight.String
	case lit == "true" || lit == "false" || lit == "iota" || lit == "nil":
		return syntaxhighlight.Literal

	case tok == token.COMMENT:
		return syntaxhighlight.Comment
	default:
		return syntaxhighlight.Plaintext
	}
}

func Print(src []byte, w io.Writer, p syntaxhighlight.Printer) error {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, scanner.ScanComments)

	var lastOffset int

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		var tokString string
		if lit != "" {
			tokString = lit
		} else {
			tokString = tok.String()
		}

		// TODO: Clean this up.
		//if tok == token.SEMICOLON {
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}

		// Whitespace between previous and current tokens.
		offset := int(fset.Position(pos).Offset)
		if whitespace := string(src[lastOffset:offset]); whitespace != "" {
			err := p.Print(w, syntaxhighlight.Whitespace, whitespace)
			if err != nil {
				return err
			}
		}
		lastOffset = offset + len(tokString)

		err := p.Print(w, TokenKind(tok, lit), tokString)
		if err != nil {
			return err
		}
	}

	// Print final whitespace after the last token.
	if whitespace := string(src[lastOffset:]); whitespace != "" {
		err := p.Print(w, syntaxhighlight.Whitespace, whitespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func Annotate(src []byte, a syntaxhighlight.Annotator) (annotate.Annotations, error) {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, scanner.ScanComments)

	var anns annotate.Annotations

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		offset := int(fset.Position(pos).Offset)

		var tokString string
		if lit != "" {
			tokString = lit
		} else {
			tokString = tok.String()
		}

		// TODO: Clean this up.
		//if tok == token.SEMICOLON {
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}

		ann, err := a.Annotate(offset, TokenKind(tok, lit), tokString)
		if err != nil {
			return nil, err
		}
		if ann != nil {
			anns = append(anns, ann)
		}
	}

	return anns, nil
}
