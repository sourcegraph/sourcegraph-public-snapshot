package coverageutil

import (
	"bytes"
	"path/filepath"
	"strings"
	"text/scanner"
	"unicode"
)

// javascriptScanner extracts JavaScript identifiers from source code.
// It's based on scanner.Scanner with adjustments to handle JavaScript-specific
// things such as:
// - binary numeric literals
// - single quote denotes a string
// - identifier may start with $ or _
// /foo/(bar)? denotes RE and not a two identifiers
type javascriptScanner struct {
	*scanner.Scanner
}

// Scan extracts next identifier or EOF token from JavaScript source
func (s *javascriptScanner) Scan() rune {
	for {
		tok := s.scan()
		switch {
		case tok == scanner.EOF:
			return tok
		case tok == scanner.Ident:
			return tok
		}
	}
}

// scan adds treatments of JavaScript-specific features.
// It returns the next token found
func (s *javascriptScanner) scan() rune {
	ch := s.Peek()
	switch {
	case ch < 0:
		return scanner.EOF
	case ch == '\'', ch == '"':
		{
			// JavaScript strings
			s.consumeString(ch)
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
	case ch == '0':
		{
			// added binary numeric literals
			// cutting 0[bBxXoO]
			s.Next()
			ch = s.Peek()
			if ch == 'x' || ch == 'X' || ch == 'b' || ch == 'B' || ch == 'o' || ch == 'O' {
				s.Next()
			}
			return s.scan()
		}
	case ch == '/':
		{
			// RE?
			s.consumeRegexp()
			return s.scan()
		}
	}
	return s.Scanner.Scan()
}

// consumeString consumes all the runes till the closing quote mark
func (s *javascriptScanner) consumeString(quote rune) rune {
	s.Next()
	ch := s.Next()
	for ch != quote {
		switch {
		case ch < 0 || ch == '\n':
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

// consumeString consumes regular expressions blocks.
// Does not consume division - a / b
func (s *javascriptScanner) consumeRegexp() rune {
	// leading /
	s.Next()
	ch := s.Peek()

	switch {
	case unicode.IsSpace(ch):
		{
			// division
			return ch
		}
	case ch == '/':
		{
			// comment
			for ch >= 0 && ch != '\n' {
				ch = s.Next()
			}
			return ch
		}
	case ch == '*':
		{
			// multiline comment
			ch = s.Next()
			state := 0
			for ch >= 0 {
				if ch == '*' {
					state = 1
				} else if ch == '/' {
					if state == 1 {
						return ch
					} else {
						state = 0
					}
				} else {
					state = 0
				}
				ch = s.Next()
			}
			return ch
		}
	}

	ch = s.Next()
	for ch != '/' {
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
	// modifier part
	ch = s.Peek()
	for unicode.IsLetter(ch) {
		s.Next()
		ch = s.Peek()
	}
	return ch
}

// newJavascriptScanner initializes and return new scanner for JavaScript language
func newJavascriptScanner() *javascriptScanner {
	s := &javascriptScanner{&scanner.Scanner{}}
	s.IsIdentRune = func(ch rune, i int) bool {
		return ch == '_' || ch == '$' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	}
	return s
}

// javascriptTokenizer produces tokens from JavaScript source code
type javascriptTokenizer struct {
	scanner *javascriptScanner
	errors  []string
}

// list of JavaScript keywords
var javascriptKeywords = map[string]bool{
	"break":        true,
	"case":         true,
	"catch":        true,
	"class":        true,
	"const":        true,
	"continue":     true,
	"debugger":     true,
	"default":      true,
	"delete":       true,
	"do":           true,
	"else":         true,
	"export":       true,
	"extends":      true,
	"finally":      true,
	"for":          true,
	"function":     true,
	"if":           true,
	"import":       true,
	"in":           true,
	"instanceof":   true,
	"new":          true,
	"of":           true,
	"return":       true,
	"super":        true,
	"switch":       true,
	"this":         true,
	"throw":        true,
	"try":          true,
	"typeof":       true,
	"var":          true,
	"void":         true,
	"while":        true,
	"with":         true,
	"yield":        true,
	"enum":         true,
	"implements":   true,
	"interface":    true,
	"let":          true,
	"package":      true,
	"private":      true,
	"protected":    true,
	"public":       true,
	"static":       true,
	"await":        true,
	"abstract":     true,
	"boolean":      true,
	"byte":         true,
	"char":         true,
	"double":       true,
	"final":        true,
	"float":        true,
	"goto":         true,
	"int":          true,
	"long":         true,
	"native":       true,
	"short":        true,
	"synchronized": true,
	"throws":       true,
	"transient":    true,
	"volatile":     true,
	"null":         true,
	"true":         true,
	"false":        true,
	"undefined":    true,
	"from":         true,
}

// Initializes text scanner that extracts only idents
func (s *javascriptTokenizer) Init(src []byte) {
	s.errors = make([]string, 0)
	s.scanner = newJavascriptScanner()
	s.scanner.Init(bytes.NewReader(src))
	s.scanner.Error = func(scanner *scanner.Scanner, msg string) {
		s.errors = append(s.errors, msg)
	}
}

func (s *javascriptTokenizer) Done() {
}

func (s *javascriptTokenizer) Errors() []string {
	return s.errors
}

// Next returns idents that are not Java keywords
func (s *javascriptTokenizer) Next() *Token {
	for {
		r := s.scanner.Scan()
		if r == scanner.EOF {
			return nil
		}
		text := s.scanner.TokenText()
		if s.isKeyword(text) {
			continue
		}
		p := s.scanner.Pos()
		return &Token{uint32(p.Offset - len([]byte(text))), p.Line, text}
	}
}

// isKeyword returns true if given identifier denotes a JavaScript keyword
func (s *javascriptTokenizer) isKeyword(ident string) bool {
	_, ok := javascriptKeywords[ident]
	return ok
}

func init() {
	var factory = func() Tokenizer {
		return &javascriptTokenizer{}
	}

	register(func(lang, path string) tokenizerFactory {
		if lang != "JavaScript" {
			return nil
		}
		if !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".jsx") {
			return nil
		}
		if strings.HasSuffix(path, ".min.js") {
			return nil
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			// fallback
			abs = path
		}
		abs = filepath.ToSlash(abs)
		if strings.Contains(abs, "/node_modules/") || strings.Contains(abs, "/bower_components/") {
			return nil
		}
		return factory
	})
}
