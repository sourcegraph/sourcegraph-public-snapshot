package ast

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/op"
)

func node(p *Parser) (*Node, error) {
	return p.Expect(
		Capture{
			Type:        NodeType,
			TypeStrings: NodeTypes,
			Value: op.And{
				'[',
				integer,
				',',
				op.Or{
					children,
					literal,
				},
				']',
			},
		},
	)
}

func literal(p *Parser) (*Node, error) {
	return p.Expect(
		Capture{
			Type:        LiteralType,
			TypeStrings: NodeTypes,
			Value: op.And{
				'"',
				op.MinZero(
					character,
				),
				'"',
			},
		},
	)
}

func character(p *Parser) (*Node, error) {
	return p.Expect(
		op.Or{
			escaped,
			parser.CheckRuneRange(0x0020, 0x0021),
			parser.CheckRuneRange(0x0023, 0x005B),
			parser.CheckRuneRange(0x005D, 0x0010FFFF),
		},
	)
}

func escaped(p *Parser) (*Node, error) {
	return p.Expect(
		op.And{
			'\\',
			op.Or{
				'b',
				'f',
				'n',
				'r',
				't',
				'"',
				'\\',
			},
		},
	)
}

func children(p *Parser) (*Node, error) {
	return p.Expect(
		Capture{
			Type:        ChildrenType,
			TypeStrings: NodeTypes,
			Value: op.And{
				'[',
				node,
				op.MinZero(
					op.And{
						',',
						node,
					},
				),
				']',
			},
		},
	)
}

func integer(p *Parser) (*Node, error) {
	return p.Expect(
		Capture{
			Type:        IntegerType,
			TypeStrings: NodeTypes,
			Value: op.And{
				op.Optional(
					'-',
				),
				op.Or{
					'0',
					op.And{
						parser.CheckRuneRange('1', '9'),
						op.MinZero(
							parser.CheckRuneRange('0', '9'),
						),
					},
				},
			},
		},
	)
}

// Node Types
const (
	Unknown = iota

	// PEGN-AST (github.com/di-wu/parser)
	NodeType     // 001
	LiteralType  // 002
	ChildrenType // 003
	IntegerType  // 004
)

var NodeTypes = []string{
	"UNKNOWN",

	// PEGN-AST (github.com/di-wu/parser)
	"Node",
	"Literal",
	"Children",
	"Integer",
}
