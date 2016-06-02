package coverageutil

import (
	"bytes"
	"text/scanner"
	"unicode"
)

// pythonScanner extracts Python identifiers from source code.
// It's based on scanner.Scanner with adjustments to handle Python-specific
// things such as:
// - single quote denotes a string
// - multiline strings ('''...''', """...""")
// - raw and Unicode strings ([rRuU]+STRING)
// - complex numbers NUMBERj
// - # comments
type pythonScanner struct {
	*scanner.Scanner
}

// Scan extracts next identifier or EOF token from Python source
func (s *pythonScanner) Scan() rune {
	for {
		tok := s.scan()
		switch {
		case tok == scanner.EOF:
			return tok
		case tok == scanner.Ident:
			if s.isStringPrefix(s.TokenText()) {
				// raw and Unicode strings
				ch := s.Peek()
				if ch == '"' || ch == '\'' {
					s.consumeString(ch, true)
					continue
				}
			}
			return tok
		}
	}
}

// scan adds treatments of Python-specific features.
// It returns the next token found
func (s *pythonScanner) scan() rune {
	ch := s.Peek()
	switch {
	case ch < 0:
		return scanner.EOF
	case ch == '\'', ch == '"':
		{
			// Python strings
			s.consumeString(ch, true)
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
	case ch >= '0' && ch <= '9':
		{
			// added complex numbers
			ch = s.Scanner.Scan()
			ch = s.Peek()
			if ch == 'j' {
				s.Next()
			}
			return s.scan()
		}
	case ch == '#':
		{
			// comments
			for ch >= 0 && ch != '\n' {
				ch = s.Next()
			}
			return s.scan()
		}
	}
	return s.Scanner.Scan()
}

// consumeString handles '...', "...", '''...''', and """..."""
func (s *pythonScanner) consumeString(quote rune, backslashEnabled bool) rune {
	s.Next()
	ch := s.Peek()
	var counter int
	var multiline bool
	if ch == quote {
		// ''' or """
		s.Next()
		s.Next()
		ch = s.Next()
		counter = 3
		multiline = true
	} else {
		ch = s.Next()
		counter = 1
		multiline = false
	}
	for counter > 0 {
		switch {
		case ch < 0:
			return ch
		case !multiline && ch == '\n':
			return ch
		case backslashEnabled && ch == '\\':
			{
				// skip backslash and the following rune
				s.Next()
			}
		case ch == quote:
			counter--
		}
		ch = s.Next()
	}
	return ch
}

// isStringPrefix returns true is string denotes Python string prefix
// (combination of upper and lowercase 'u' and 'r')
func (s *pythonScanner) isStringPrefix(prefix string) bool {
	for _, ch := range prefix {
		if ch != 'r' && ch != 'R' && ch != 'u' && ch != 'U' {
			return false
		}
	}
	return true
}

// newPythonScanner initializes and return new scanner for Python language
func newPythonScanner() *pythonScanner {
	s := &pythonScanner{&scanner.Scanner{}}
	return s
}

// pythonTokenizer produces tokens from Python source code
type pythonTokenizer struct {
	scanner *pythonScanner
	errors  []string
}

// list of Python keywords
var pythonKeywords = map[string]bool{
	"and":      true,
	"as":       true,
	"assert":   true,
	"break":    true,
	"class":    true,
	"continue": true,
	"def":      true,
	"del":      true,
	"elif":     true,
	"else":     true,
	"except":   true,
	"exec":     true,
	"False":    true,
	"finally":  true,
	"for":      true,
	"from":     true,
	"global":   true,
	"if":       true,
	"import":   true,
	"in":       true,
	"is":       true,
	"lambda":   true,
	"None":     true,
	"nonlocal": true,
	"not":      true,
	"or":       true,
	"pass":     true,
	"print":    true,
	"raise":    true,
	"return":   true,
	"True":     true,
	"try":      true,
	"while":    true,
	"with":     true,
	"yield":    true,
	"async":    true,
	"await":    true,
}

// Initializes text scanner that extracts only idents
func (s *pythonTokenizer) Init(src []byte) {
	s.errors = make([]string, 0)
	s.scanner = newPythonScanner()
	s.scanner.Init(bytes.NewReader(src))
	s.scanner.Error = func(scanner *scanner.Scanner, msg string) {
		s.errors = append(s.errors, msg)
	}
}

func (s *pythonTokenizer) Done() {
}

func (s *pythonTokenizer) Errors() []string {
	return s.errors
}

// Next returns idents that are not Java keywords
func (s *pythonTokenizer) Next() *Token {
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

// isKeyword returns true if given identifier denotes a Python keyword
func (s *pythonTokenizer) isKeyword(ident string) bool {
	_, ok := pythonKeywords[ident]
	return ok
}

func init() {
	var factory = func() Tokenizer {
		return &pythonTokenizer{}
	}
	newExtensionBasedLookup("Python", []string{".py"}, factory)
}
