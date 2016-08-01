package coverageutil

import (
	"go/scanner"
	"go/token"
)

type goTokenizer struct {
	scanner *scanner.Scanner
	fset    *token.FileSet
	errors  []string
}

// Initializes Go scanner
func (s *goTokenizer) Init(src []byte) {
	s.errors = make([]string, 0)
	s.scanner = &scanner.Scanner{}
	s.fset = token.NewFileSet()
	file := s.fset.AddFile("", s.fset.Base(), len(src))
	s.scanner.Init(file, src, func(pos token.Position, msg string) {
		s.errors = append(s.errors, msg)
	}, 0)
}

func (s *goTokenizer) Done() {
}

func (s *goTokenizer) Errors() []string {
	return s.errors
}

func (s *goTokenizer) Next() *Token {
	for {
		pos, tok, lit := s.scanner.Scan()
		if tok == token.EOF {
			return nil
		}
		if tok != token.IDENT {
			continue
		}
		p := s.fset.Position(pos)
		return &Token{uint32(p.Offset), p.Line, p.Column, lit}
	}
}

func init() {
	var factory = func() Tokenizer {
		return &goTokenizer{}
	}
	newExtensionBasedLookup("Go", []string{".go"}, factory)
}
