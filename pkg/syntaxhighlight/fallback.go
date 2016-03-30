package syntaxhighlight

import (
	"bytes"
	"text/scanner"
	"unicode"
	"unicode/utf8"
)

// fallback syntax highlight implementation, copied from https://github.com/sourcegraph/syntaxhighlight
type fallbackLexer struct {
	// scanner
	s *scanner.Scanner
	// current offset in source code
	offset int
}

// List of known keywords
var keywords = map[string]struct{}{
	"BEGIN":            {},
	"END":              {},
	"False":            {},
	"Infinity":         {},
	"NaN":              {},
	"None":             {},
	"True":             {},
	"abstract":         {},
	"alias":            {},
	"align_union":      {},
	"alignof":          {},
	"and":              {},
	"append":           {},
	"as":               {},
	"asm":              {},
	"assert":           {},
	"auto":             {},
	"axiom":            {},
	"begin":            {},
	"bool":             {},
	"boolean":          {},
	"break":            {},
	"byte":             {},
	"caller":           {},
	"case":             {},
	"catch":            {},
	"char":             {},
	"class":            {},
	"concept":          {},
	"concept_map":      {},
	"const":            {},
	"const_cast":       {},
	"constexpr":        {},
	"continue":         {},
	"debugger":         {},
	"decltype":         {},
	"def":              {},
	"default":          {},
	"defined":          {},
	"del":              {},
	"delegate":         {},
	"delete":           {},
	"die":              {},
	"do":               {},
	"double":           {},
	"dump":             {},
	"dynamic_cast":     {},
	"elif":             {},
	"else":             {},
	"elsif":            {},
	"end":              {},
	"ensure":           {},
	"enum":             {},
	"eval":             {},
	"except":           {},
	"exec":             {},
	"exit":             {},
	"explicit":         {},
	"export":           {},
	"extends":          {},
	"extern":           {},
	"false":            {},
	"final":            {},
	"finally":          {},
	"float":            {},
	"float32":          {},
	"float64":          {},
	"for":              {},
	"foreach":          {},
	"friend":           {},
	"from":             {},
	"func":             {},
	"function":         {},
	"generic":          {},
	"get":              {},
	"global":           {},
	"goto":             {},
	"if":               {},
	"implements":       {},
	"import":           {},
	"in":               {},
	"inline":           {},
	"instanceof":       {},
	"int":              {},
	"int8":             {},
	"int16":            {},
	"int32":            {},
	"int64":            {},
	"interface":        {},
	"is":               {},
	"lambda":           {},
	"last":             {},
	"late_check":       {},
	"local":            {},
	"long":             {},
	"make":             {},
	"map":              {},
	"module":           {},
	"mutable":          {},
	"my":               {},
	"namespace":        {},
	"native":           {},
	"new":              {},
	"next":             {},
	"nil":              {},
	"no":               {},
	"nonlocal":         {},
	"not":              {},
	"null":             {},
	"nullptr":          {},
	"operator":         {},
	"or":               {},
	"our":              {},
	"package":          {},
	"pass":             {},
	"print":            {},
	"private":          {},
	"property":         {},
	"protected":        {},
	"public":           {},
	"raise":            {},
	"redo":             {},
	"register":         {},
	"reinterpret_cast": {},
	"require":          {},
	"rescue":           {},
	"retry":            {},
	"return":           {},
	"self":             {},
	"set":              {},
	"short":            {},
	"signed":           {},
	"sizeof":           {},
	"static":           {},
	"static_assert":    {},
	"static_cast":      {},
	"strictfp":         {},
	"struct":           {},
	"sub":              {},
	"super":            {},
	"switch":           {},
	"synchronized":     {},
	"template":         {},
	"then":             {},
	"this":             {},
	"throw":            {},
	"throws":           {},
	"transient":        {},
	"true":             {},
	"try":              {},
	"type":             {},
	"typedef":          {},
	"typeid":           {},
	"typename":         {},
	"typeof":           {},
	"undef":            {},
	"undefined":        {},
	"union":            {},
	"unless":           {},
	"unsigned":         {},
	"until":            {},
	"use":              {},
	"using":            {},
	"var":              {},
	"virtual":          {},
	"void":             {},
	"volatile":         {},
	"wantarray":        {},
	"when":             {},
	"where":            {},
	"while":            {},
	"with":             {},
	"yield":            {},
}

// Initializes scanner object
func (self *fallbackLexer) Init(source []byte) {
	self.s = &scanner.Scanner{}
	self.s.Init(bytes.NewReader(source))
	self.s.Error = func(_ *scanner.Scanner, _ string) {}
	self.s.Whitespace = 0
	self.s.Mode = self.s.Mode ^ scanner.SkipComments

	self.offset = 0
}

// Produces token using scanner's output
func (self *fallbackLexer) NextToken() *Token {

	tok := self.s.Scan()
	if tok == scanner.EOF {
		return nil
	}

	text := self.s.TokenText()
	ret := NewToken([]byte(text), tokenKind(tok, text), self.offset)
	self.offset += len(text)

	return &ret
}

// Converts pair (token kind, token text) to token type
func tokenKind(tok rune, tokText string) *TokenType {
	switch tok {
	case scanner.Ident:
		if _, isKW := keywords[tokText]; isKW {
			return Keyword
		}
		if r, _ := utf8.DecodeRuneInString(tokText); unicode.IsUpper(r) {
			return Keyword_Type
		}
		return Name_Other
	case scanner.Float:
		return Number_Float
	case scanner.Int:
		return Number_Integer
	case scanner.String, scanner.RawString:
		return String
	case scanner.Char:
		return String_Char
	case scanner.Comment:
		return Comment
	}
	if unicode.IsSpace(tok) {
		return Whitespace
	}
	return Punctuation
}
