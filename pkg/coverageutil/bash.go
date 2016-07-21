package coverageutil

import (
	"bytes"
	"fmt"

	"github.com/mkovacs/bash/scanner"
)

// bashIdentScanner extracts identifiers from a Bash script
type bashIdentScanner struct {
	is     *scanner.Scanner
	errors []string
}

// Init initializes text scanner that extracts only idents
func (s *bashIdentScanner) Init(src []byte) {
	s.is = &scanner.Scanner{}
	s.is.Init(bytes.NewReader(src))
}

func (s *bashIdentScanner) Done() {
}

func (s *bashIdentScanner) Errors() []string {
	return []string{}
}

// Next returns the next identifier
func (s *bashIdentScanner) Next() *Token {
	for {
		tok, err := s.is.Scan()
		if err != nil {
			s.errors = append(s.errors, fmt.Sprintf("error: %s", err))
			return nil
		}
		switch {
		case tok == scanner.EOF:
			return nil
		case tok == scanner.Ident:
			text := s.is.TokenText()
			p := s.is.Pos()
			return &Token{uint32(p.Offset - len(text)), p.Line, p.Column, text}
		}
	}
}

func init() {
	factory := func() Tokenizer {
		return &bashIdentScanner{}
	}
	newExtensionBasedLookup("Bash", []string{".sh", ".bash"}, factory)
}
