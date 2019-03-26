package syntax

import "fmt"

// ParseError describes an error in query parsing.
type ParseError struct {
	Pos int    // the character position where the error occurred
	Msg string // description of the error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at character %d: %s", e.Pos, e.Msg)
}

type parser struct {
	tokens []Token
	pos    int
}

// context holds settings active within a given scope during parsing.
type context struct {
	field string // name of the field currently in scope (or "")
}

// Parse parses the query and returns its parse tree. Returned errors are of
// type *ParseError, which includes the error position and message.
//
// BNF-ish query syntax:
//
//   exprList  := {exprSign} | exprSign (sep exprSign)*
//   exprSign  := {"-"} expr
//   expr      := fieldExpr | lit | quoted | pattern
//   fieldExpr := lit ":" value
//   value     := lit | quoted
func Parse(input string) (*Query, error) {
	tokens := Scan(input)
	p := parser{tokens: tokens}
	ctx := context{field: ""}
	exprs, err := p.parseExprList(ctx)
	if err != nil {
		return nil, err
	}
	return &Query{Expr: exprs, Input: input}, nil
}

// peek returns the next token without consuming it. Peeking beyond the end of
// the token stream will return TokenEOF.
func (p *parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

// backup steps back one position in the token stream.
func (p *parser) backup() {
	p.pos--
}

// next returns the next token in the stream and advances the cursor.
func (p *parser) next() Token {
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		p.pos++
		return tok
	}
	p.pos++ // to make sure (*parser).backup works
	return Token{Type: TokenEOF}
}

// exprList := {exprSign} | exprSign (sep exprSign)*
func (p *parser) parseExprList(ctx context) (exprList []*Expr, err error) {
	if p.peek().Type == TokenEOF {
		return nil, nil
	}

	for {
		tok := p.peek()
		if tok.Type == TokenEOF {
			break
		}
		if tok.Type == TokenSep {
			p.next()
			continue
		}

		expr, err := p.parseExprSign(ctx)
		if err != nil {
			return nil, err
		}
		exprList = append(exprList, expr)
	}

	return exprList, nil
}

// exprSign := {"-"} expr
func (p *parser) parseExprSign(ctx context) (*Expr, error) {
	tok := p.next()
	switch tok.Type {
	case TokenMinus:
		// consume token
	default:
		tok = Token{Type: TokenEOF}
		p.backup()
	}

	expr, err := p.parseExpr(ctx)
	if err != nil {
		return nil, err
	}

	switch tok.Type {
	case TokenMinus:
		expr.Not = true
	}

	return expr, nil
}

// expr := exprField | lit | quoted | pattern
func (p *parser) parseExpr(ctx context) (*Expr, error) {
	tok := p.next()
	switch tok.Type {
	case TokenLiteral:
		tok2 := p.next()
		switch tok2.Type {
		case TokenColon:
			valueTok := p.next()
			switch valueTok.Type {
			case TokenLiteral, TokenQuoted:
				if tok3 := p.next(); tok3.Type != TokenSep && tok3.Type != TokenEOF {
					return nil, &ParseError{Pos: tok3.Pos, Msg: fmt.Sprintf("got %s, want separator or EOF", tok3.Type)}
				}
				return &Expr{Pos: tok.Pos, Field: tok.Value, Value: valueTok.Value, ValueType: valueTok.Type}, nil
			case TokenSep, TokenEOF:
				return &Expr{Pos: tok.Pos, Field: tok.Value, Value: "", ValueType: TokenLiteral}, nil
			default:
				return nil, &ParseError{Pos: valueTok.Pos, Msg: fmt.Sprintf("got %s, want value", valueTok.Type)}
			}
		case TokenSep, TokenEOF:
			return &Expr{Pos: tok.Pos, Value: tok.Value, ValueType: tok.Type}, nil
		default:
			panic("unreachable")
		}
	case TokenQuoted, TokenPattern:
		tok2 := p.next()
		switch tok2.Type {
		case TokenSep, TokenEOF:
			return &Expr{Pos: tok.Pos, Value: tok.Value, ValueType: tok.Type}, nil
		default:
			return nil, &ParseError{Pos: tok2.Pos, Msg: fmt.Sprintf("got %s, want separator or EOF", tok2.Type)}
		}
	}

	return nil, &ParseError{Pos: tok.Pos, Msg: fmt.Sprintf("got %s, want expr", tok.Type)}
}
