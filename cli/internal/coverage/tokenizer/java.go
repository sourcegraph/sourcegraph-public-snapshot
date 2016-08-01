package tokenizer

import (
	"bytes"
	"strings"
	"unicode"

	"text/scanner"
)

// javaTokenizer produces tokens from Java source code
type javaTokenizer struct {
	scanner *scanner.Scanner
	errors  []string
}

// list of Java keywords
var javaKeywords = map[string]bool{
	"abstract":     true,
	"continue":     true,
	"for":          true,
	"new":          true,
	"switch":       true,
	"assert":       true,
	"default":      true,
	"goto":         true,
	"package":      true,
	"synchronized": true,
	"boolean":      true,
	"do":           true,
	"if":           true,
	"private":      true,
	"this":         true,
	"break":        true,
	"double":       true,
	"implements":   true,
	"protected":    true,
	"throw":        true,
	"byte":         true,
	"else":         true,
	"import":       true,
	"public":       true,
	"throws":       true,
	"case":         true,
	"enum":         true,
	"instanceof":   true,
	"return":       true,
	"transient":    true,
	"catch":        true,
	"extends":      true,
	"int":          true,
	"short":        true,
	"try":          true,
	"char":         true,
	"final":        true,
	"interface":    true,
	"static":       true,
	"void":         true,
	"class":        true,
	"finally":      true,
	"long":         true,
	"strictfp":     true,
	"volatile":     true,
	"const":        true,
	"float":        true,
	"native":       true,
	"super":        true,
	"while":        true,
	"true":         true,
	"false":        true,
	"null":         true,
}

// Initializes text scanner that extracts only idents
func (s *javaTokenizer) Init(src []byte) {
	s.scanner = &scanner.Scanner{}
	s.scanner.Init(bytes.NewReader(src))
	s.errors = make([]string, 0)
	s.scanner.Error = func(scanner *scanner.Scanner, msg string) {
		s.errors = append(s.errors, msg)
	}
}

func (s *javaTokenizer) Done() {
}

func (s *javaTokenizer) Errors() []string {
	return s.errors
}

// Next returns idents that are not Java keywords
func (s *javaTokenizer) Next() *Token {
	for {
		ch := s.scanner.Peek()
		if ch >= '0' && ch <= '9' {
			s.consumeNumericLiteral()
			continue
		} else if unicode.IsSpace(ch) {
			// consuming spaces
			for unicode.IsSpace(ch) {
				s.scanner.Next()
				ch = s.scanner.Peek()
			}
			continue
		} else if ch == '"' || ch == '\'' {
			s.consumeStringLiteral(ch)
			continue
		}

		r := s.scanner.Scan()
		if r == scanner.EOF {
			return nil
		}
		if r != scanner.Ident {
			continue
		}
		text := s.scanner.TokenText()
		if s.isKeyword(text) {
			// consume package or import qualifiers
			if text == "package" || text == "import" {
				ch = s.scanner.Next()
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

// isKeyword returns true if given identifier denotes a Java keyword
func (s *javaTokenizer) isKeyword(ident string) bool {
	_, ok := javaKeywords[ident]
	return ok
}

func (s *javaTokenizer) consumeNumericLiteral() {
	ch := s.scanner.Peek()
	for strings.ContainsRune("0123456789xXlLdDfFbBaAcCeE_+-.", ch) {
		s.scanner.Next()
		ch = s.scanner.Peek()
	}
}

// consumeStringLiteral consumes all the runes till the closing quote mark
func (s *javaTokenizer) consumeStringLiteral(quote rune) {
	s.scanner.Next()
	ch := s.scanner.Next()
	for ch != quote {
		switch {
		case ch < 0 || ch == '\n':
			return
		case ch == '\\':
			{
				// skip backslash and the following rune
				s.scanner.Next()
			}
		}
		ch = s.scanner.Next()
	}
}

func init() {
	var factory = func() Tokenizer {
		return &javaTokenizer{}
	}
	newExtensionBasedLookup("Java", []string{".java"}, factory)
}
