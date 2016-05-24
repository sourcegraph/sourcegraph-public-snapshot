package coverageutil

import (
	"go/scanner"
	"go/token"
)

type goTokenizer struct {
	scanner *scanner.Scanner
	fset    *token.FileSet
}

// Initializes Go scanner
func (s *goTokenizer) Init(src []byte) {
	s.scanner = &scanner.Scanner{}
	s.fset = token.NewFileSet()
	file := s.fset.AddFile("", s.fset.Base(), len(src))
	s.scanner.Init(file, src, nil, 0)
}

func (s *goTokenizer) Done() {
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
		return &Token{uint32(p.Offset), lit}
	}
}

func init() {
	var factory = func() Tokenizer {
		return &goTokenizer{}
	}
	newExtensionBasedLookup("Go", []string{".go"}, factory)
}
