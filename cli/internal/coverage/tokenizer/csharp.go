package tokenizer

import (
	"bytes"
	"text/scanner"
	"unicode"
)

// csharpScanner extracts C# identifiers from source code.
// It's based on scanner.Scanner with adjustments to handle C#-specific
// things such as:
// - strings can be multiline
// - numeric literals may have a suffix
// - verbatim strings
// - preprocessor directives
type csharpScanner struct {
	*scanner.Scanner
}

// Scan extracts next identifier or EOF token from C# source
func (s *csharpScanner) Scan() rune {
	for {
		tok := s.scan()
		switch {
		case tok == scanner.EOF:
			return tok
		case tok == scanner.Ident:
			return tok
		case tok == scanner.Int, tok == scanner.Float:
			s.consumeNumericSuffix()
		}
	}
}

// scan adds treatments of C#-specific features.
// It returns the next token found
func (s *csharpScanner) scan() rune {
	ch := s.Peek()
	switch {
	case ch < 0:
		return scanner.EOF
	case ch == '"':
		{
			// C# strings
			s.consumeString()
			return s.scan()
		}
	case ch == '@':
		{
			s.Next()
			ch = s.Peek()
			if ch == '"' {
				// verbatim strings
				s.consumeVerbatimString()
				return s.scan()
			}
		}
	case ch == '#':
		{
			for ch >= 0 && ch != '\n' {
				ch = s.Next()
			}
			return s.scan()
		}
	case unicode.IsSpace(ch):
		{
			// consuming spaces
			for unicode.IsSpace(ch) {
				s.Next()
				ch = s.Peek()
			}
			return s.scan()
		}
	}
	return s.Scanner.Scan()
}

// consumeString consumes all the runes till the closing quote mark
func (s *csharpScanner) consumeString() rune {
	s.Next()
	ch := s.Next()
	for ch != '"' {
		switch {
		case ch < 0:
			return ch
		case ch == '\\':
			{
				// skip backslash and the following rune
				s.Next()
			}
		}
		ch = s.Next()
	}
	return ch
}

// consumeVerbatimString consumes C# verbatim strings
func (s *csharpScanner) consumeVerbatimString() rune {
	s.Next()
	ch := s.Next()
L:
	for {
		switch {
		case ch < 0:
			return ch
		case ch == '"':
			{
				ch = s.Next()
				if ch != '"' {
					break L
				}
			}
		}
		ch = s.Next()
	}
	return ch
}

// consumeNumericSuffix consumes suffix of numeric literals
func (s *csharpScanner) consumeNumericSuffix() {
	ch := s.Peek()
	for unicode.IsLetter(ch) {
		s.Next()
		ch = s.Peek()
	}
}

// newCsharpScanner initializes and return new scanner for C# language
func newCsharpScanner() *csharpScanner {
	s := &csharpScanner{&scanner.Scanner{}}
	return s
}

// csharpTokenizer produces tokens from C# source code
type csharpTokenizer struct {
	scanner *csharpScanner
	errors  []string
}

// list of C# keywords
var csharpKeywords = map[string]bool{
	"abstract":   true,
	"as":         true,
	"base":       true,
	"bool":       true,
	"break":      true,
	"byte":       true,
	"case":       true,
	"catch":      true,
	"char":       true,
	"checked":    true,
	"class":      true,
	"const":      true,
	"continue":   true,
	"decimal":    true,
	"default":    true,
	"delegate":   true,
	"do":         true,
	"double":     true,
	"else":       true,
	"enum":       true,
	"event":      true,
	"explicit":   true,
	"extern":     true,
	"false":      true,
	"finally":    true,
	"fixed":      true,
	"float":      true,
	"for":        true,
	"foreach":    true,
	"goto":       true,
	"if":         true,
	"implicit":   true,
	"in":         true,
	"int":        true,
	"interface":  true,
	"internal":   true,
	"is":         true,
	"lock":       true,
	"long":       true,
	"namespace":  true,
	"new":        true,
	"null":       true,
	"object":     true,
	"operator":   true,
	"out":        true,
	"override":   true,
	"params":     true,
	"private":    true,
	"protected":  true,
	"public":     true,
	"readonly":   true,
	"ref":        true,
	"return":     true,
	"sbyte":      true,
	"sealed":     true,
	"short":      true,
	"sizeof":     true,
	"stackalloc": true,
	"static":     true,
	"string":     true,
	"struct":     true,
	"switch":     true,
	"this":       true,
	"throw":      true,
	"true":       true,
	"try":        true,
	"typeof":     true,
	"uint":       true,
	"ulong":      true,
	"unchecked":  true,
	"unsafe":     true,
	"ushort":     true,
	"using":      true,
	"virtual":    true,
	"void":       true,
	"volatile":   true,
	"while":      true,
	"add":        true,
	"alias":      true,
	"ascending":  true,
	"async":      true,
	"await":      true,
	"descending": true,
	"dynamic":    true,
	"from":       true,
	"get":        true,
	"global":     true,
	"group":      true,
	"into":       true,
	"join":       true,
	"let":        true,
	"orderby":    true,
	"partial":    true,
	"remove":     true,
	"select":     true,
	"set":        true,
	"value":      true,
	"var":        true,
	"where":      true,
	"yield":      true,
}

// Initializes text scanner that extracts only idents
func (s *csharpTokenizer) Init(src []byte) {
	s.errors = make([]string, 0)
	s.scanner = newCsharpScanner()
	s.scanner.Init(bytes.NewReader(src))
	s.scanner.Error = func(scanner *scanner.Scanner, msg string) {
		s.errors = append(s.errors, msg)
	}
}

func (s *csharpTokenizer) Done() {
}

func (s *csharpTokenizer) Errors() []string {
	return s.errors
}

// Next returns idents that are not C# keywords
func (s *csharpTokenizer) Next() *Token {
	for {
		r := s.scanner.Scan()
		if r == scanner.EOF {
			return nil
		}
		text := s.scanner.TokenText()
		if s.isKeyword(text) {
			if text == "using" {
				ch := s.scanner.Next()
				for ch >= 0 && ch != ';' {
					ch = s.scanner.Next()
				}
			}
			continue
		}
		p := s.scanner.Pos()
		return &Token{uint32(p.Offset - len([]byte(text))), p.Line, p.Column, text}
	}
}

// isKeyword returns true if given identifier denotes a C# keyword
func (s *csharpTokenizer) isKeyword(ident string) bool {
	_, ok := csharpKeywords[ident]
	return ok
}

func init() {
	var factory = func() Tokenizer {
		return &csharpTokenizer{}
	}
	newExtensionBasedLookup("C#", []string{".cs"}, factory)
}
