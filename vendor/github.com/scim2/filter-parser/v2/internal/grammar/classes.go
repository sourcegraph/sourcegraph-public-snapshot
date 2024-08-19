package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/op"
)

func Alpha(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(op.Or{
		parser.CheckRuneRange('A', 'Z'),
		parser.CheckRuneRange('a', 'z'),
	})
}

func Digit(p *parser.Parser) (*parser.Cursor, bool) {
	return p.Check(parser.CheckRuneRange('0', '9'))
}
