package tokenizer

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

// See https://golang.org/ref/spec#Predeclared_identifiers
var universeBlock = []string{
	// Types:
	"bool", "byte", "complex64", "complex128", "error", "float32", "float64",
	"int", "int8", "int16", "int32", "int64", "rune", "string",
	"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",

	// Constants:
	"true", "false", "iota",

	// Zero value:
	"nil",

	// Functions:
	//
	// These are omitted from exclusion here because we should have
	// hover / j2d for these from the 'builtin' package.
	//
	//"append", "cap", "close", "complex", "copy", "delete", "imag", "len",
	//"make", "new", "panic", "print", "println", "real", "recover",
}

func (s *goTokenizer) Next() *Token {
t:
	for {
		pos, tok, lit := s.scanner.Scan()
		if tok == token.EOF {
			return nil
		}
		if tok != token.IDENT {
			continue
		}
		if lit == "_" {
			continue
		}
		for _, ident := range universeBlock {
			if lit == ident {
				continue t
			}
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
