package tokenizer

import "github.com/chris-ramon/douceur/parser"

// cssTokenizer produces tokens from CSS source code
type cssTokenizer struct {
	tokens      []*Token
	index       int
	lineOffsets [][]int
	errors      []string
}

// Initializes CSS parser
func (s *cssTokenizer) Init(src []byte) {

	text := string(src)
	s.errors = make([]string, 0)

	css, err := parser.Parse(text)
	if err != nil {
		s.errors = append(s.errors, err.Error())
		return
	}

	s.tokens = make([]*Token, 0)
	s.index = 0
	s.calcLineOffsets(text)

	for _, r := range css.Rules {
		for _, sel := range r.Selectors {
			s.tokens = append(s.tokens, &Token{s.byteOffset(sel.Line, sel.Column), sel.Line, sel.Column, sel.Value})
		}
		for _, decl := range r.Declarations {
			s.tokens = append(s.tokens, &Token{s.byteOffset(decl.Line, decl.Column), decl.Line, decl.Column, decl.Property})
		}
	}
}

func (s *cssTokenizer) Done() {
}

func (s *cssTokenizer) Errors() []string {
	return s.errors
}

// Next returns next token (selector or declaration)
func (s *cssTokenizer) Next() *Token {
	if s.index >= len(s.tokens) {
		return nil
	}
	ret := s.tokens[s.index]
	s.index++
	return ret
}

// calcLineOffsets calculates line offsets table.
// Table item (L,C) points to byte offset of rune located at the line L and column C
func (s *cssTokenizer) calcLineOffsets(src string) {
	s.lineOffsets = make([][]int, 0)
	line := make([]int, 0)
	for i, ch := range src {
		if ch == '\n' {
			s.lineOffsets = append(s.lineOffsets, line)
			line = make([]int, 0)
		} else {
			line = append(line, i)
		}
	}
	s.lineOffsets = append(s.lineOffsets, line)
}

// byteOffset returns byte-base offset of token located at (L, C)
func (s *cssTokenizer) byteOffset(line, column int) uint32 {
	// TODO(alexsaveliev) CSS parser sometimes returns incorrect locations
	if line > len(s.lineOffsets) || column > len(s.lineOffsets[line-1]) {
		return 0
	}
	return uint32(s.lineOffsets[line-1][column-1])
}

func init() {
	var factory = func() Tokenizer {
		return &cssTokenizer{}
	}
	newExtensionBasedLookup("CSS", []string{".css"}, factory)
}
