package syntaxhighlight

import (
	"go/scanner"
	"go/token"
	"strings"
)

// Returns token type and token text based on given scanner's token and literal
type tokenProcessor func(tok token.Token, lit string) (ttype *TokenType, text string)

// Denotes golang scanner rules
type goTokenDef struct {
	// token type to return if current token matches any of tokens below.
	// May be null if token processing should be done by 'processor' function
	ttype *TokenType
	// list of token types current rule should be applied for
	tokens []token.Token
	// custom token processor function
	processor tokenProcessor
}

// List of golang scanner rules
var goTokenDefs []goTokenDef

// Initializes golang scanner rules
func init() {
	goTokenDefs = []goTokenDef{
		{nil, []token.Token{token.COMMENT}, goComment},
		{Keyword_Namespace, []token.Token{token.IMPORT, token.PACKAGE}, nil},
		{Keyword_Declaration, []token.Token{token.VAR, token.FUNC, token.STRUCT, token.MAP, token.CHAN,
			token.TYPE, token.INTERFACE, token.CONST}, nil},
		{Keyword, []token.Token{token.BREAK, token.DEFAULT, token.SELECT, token.CASE, token.DEFER, token.GO,
			token.ELSE, token.GOTO, token.SWITCH, token.FALLTHROUGH, token.IF, token.RANGE, token.CONTINUE, token.FOR,
			token.RETURN}, nil},
		{Number, []token.Token{token.IMAG}, nil},
		{String_Char, []token.Token{token.CHAR}, nil},
		{Number_Float, []token.Token{token.FLOAT}, nil},
		{String, []token.Token{token.STRING}, nil},
		{nil, []token.Token{token.INT}, goInt},
		{Punctuation, []token.Token{token.OR, token.XOR, token.LSS, token.GTR, token.ASSIGN, token.NOT,
			token.LPAREN, token.RPAREN, token.LBRACK, token.RBRACK, token.LBRACE, token.RBRACE, token.COMMA,
			token.PERIOD, token.SEMICOLON, token.COLON}, nil},
		{Operator, []token.Token{token.ADD, token.AND, token.ADD_ASSIGN, token.AND_ASSIGN, token.LAND,
			token.EQL, token.NEQ, token.SUB, token.SUB_ASSIGN, token.OR_ASSIGN, token.LOR, token.LEQ, token.MUL,
			token.MUL_ASSIGN, token.XOR_ASSIGN, token.ARROW, token.GEQ, token.QUO, token.SHL, token.SHR, token.INC,
			token.QUO, token.QUO_ASSIGN, token.REM, token.REM_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN, token.DEFINE,
			token.ELLIPSIS, token.DEC, token.AND_NOT, token.AND_NOT_ASSIGN}, nil},
		{nil, []token.Token{token.IDENT}, goIdent},
	}
}

// Returns true if token matches any of given tokens (if token in array return true)
func goTokenMatches(tok token.Token, tokens []token.Token) bool {
	for _, token := range tokens {
		if tok == token {
			return true
		}
	}
	return false
}

// Returns true if literal matches any of given words (if literal in array return true)
func goLitMatches(lit string, literals []string) bool {
	for _, literal := range literals {
		if lit == literal {
			return true
		}
	}
	return false
}

// Processes golang comments. Returns single/multiline comment tokens based on comment's text
func goComment(tok token.Token, lit string) (ttype *TokenType, text string) {
	if strings.HasPrefix(lit, `/*`) {
		return Comment_Multiline, lit
	}
	return Comment_Single, lit
}

// Processes golang integers (hexadecimal, octal, and regular ones)
func goInt(tok token.Token, lit string) (ttype *TokenType, text string) {
	if strings.HasPrefix(lit, `0x`) || strings.HasPrefix(lit, `0X`) {
		return Number_Hex, lit
	}
	if len(lit) > 1 && lit[0] == '0' {
		return Number_Oct, lit
	}
	return Number_Integer, lit
}

// Processes golang identifiers, splits them to constants, types, and the rest
func goIdent(tok token.Token, lit string) (ttype *TokenType, text string) {
	if goLitMatches(lit, []string{`true`, `false`, `iota`, `nil`}) {
		return Keyword_Constant, lit
	}
	if goLitMatches(lit, []string{`uint`, `uint8`, `uint16`, `uint32`, `uint64`, `int`, `int8`, `int16`, `int32`,
		`int64`, `float`, `float32`, `float64`, `complex64`, `complex128`, `byte`, `rune`, `string`, `bool`, `error`,
		`uintptr`}) {
		return Keyword_Type, lit
	}
	return Name_Other, lit
}

// Returns true if given name is a golang's builtin function
func goBuiltIn(name string) bool {
	return goLitMatches(name, []string{`print`, `println`, `panic`, `recover`, `close`, `complex`, `real`, `imag`,
		`len`, `cap`, `append`, `copy`, `delete`, `new`, `make`})
}

// Lexer for golang
type goLexer struct {
	// tokens cache
	cache []*Token
	// scanner
	scanner *scanner.Scanner
	// fileset object that keeps positions
	fset *token.FileSet
	// last token emitted
	tok token.Token
}

// Emits zero or one token. Some tokens require special processing,
// for example "KEYWORD(" may denote a built-in function
func (self *goLexer) emit(tok *Token) *Token {

	// special treatment of built-in functions
	if tok.Type == Punctuation && tok.Text == `(` {
		l := len(self.cache)
		if l == 0 {
			return tok
		}
		prev := self.cache[l-1]
		if prev.Type == Keyword_Type {
			// type cast, for example string()
			prev.Type = Name_Builtin
		} else if prev.Type == Name_Other {
			// built-in functions such as panic()
			if goBuiltIn(prev.Text) {
				prev.Type = Name_Builtin
			} else {
				prev.Type = Name_Attribute
			}
		}
		self.cache = append(self.cache, tok)
		return self.consumeCache()
	}
	self.cache = append(self.cache, tok)
	if tok.Type == Keyword_Type || tok.Type == Name_Other {
		return self.NextToken()
	}
	return self.consumeCache()
}

// Emits one token from cache if any
func (self *goLexer) consumeCache() *Token {
	l := len(self.cache)
	if l > 0 {
		ret := self.cache[0]
		self.cache = self.cache[1:]
		return ret
	}
	return nil
}

// Initializes lexer
func (self *goLexer) Init(source []byte) {

	self.cache = make([]*Token, 0, 10)
	self.tok = token.ILLEGAL
	self.scanner = &scanner.Scanner{}
	self.fset = token.NewFileSet()
	file := self.fset.AddFile(``, self.fset.Base(), len(source))
	self.scanner.Init(file, source, nil /* no error handler */, scanner.ScanComments)
}

// Produces next token if possible
func (self *goLexer) NextToken() *Token {

	if self.tok == token.EOF {
		// end of input, return cached tokens if any
		return self.consumeCache()
	}

	pos, tok, lit := self.scanner.Scan()
	self.tok = tok
	if tok == token.EOF {
		return self.NextToken()
	}
	for _, def := range goTokenDefs {
		if !goTokenMatches(tok, def.tokens) {
			continue
		}

		var ttype *TokenType
		var text string
		if tok == token.SEMICOLON && lit == "\n" {
			return self.NextToken()
		}
		if def.processor != nil {
			ttype, text = def.processor(tok, lit)
		} else {
			ttype = def.ttype
			text = lit
		}
		if text == `` {
			if tok.IsOperator() {
				text = tok.String()
			}
		}
		return self.emit(&Token{text, ttype, self.fset.Position(pos).Offset})
	}
	return nil
}

func init() {
	var factory LexerFactory
	factory = func() Lexer {
		return &goLexer{}
	}
	register([]string{`.go`}, []string{`text\x-gosrc`}, factory)
}
