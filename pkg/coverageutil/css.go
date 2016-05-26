package coverageutil

import (
	"github.com/chris-ramon/douceur/parser"
)

// cssTokenizer produces tokens from CSS source code
type cssTokenizer struct {
	tokens      []*Token
	index       int
	lineOffsets []int
}

// Initializes CSS parser
func (s *cssTokenizer) Init(src []byte) {
	css, err := parser.Parse(string(src))
	s.tokens = make([]*Token, 0)
	s.index = 0
	s.calcLineOffsets(src)

	if err != nil {
		return
	}
	for _, r := range css.Rules {
		for _, sel := range r.Selectors {
			s.tokens = append(s.tokens, &Token{s.byteOffset(sel.Line, sel.Column), sel.Value})
		}
		for _, decl := range r.Declarations {
			s.tokens = append(s.tokens, &Token{s.byteOffset(decl.Line, decl.Column), decl.Value})
		}
	}
}

func (s *cssTokenizer) Done() {
}

// Next returns next token (selector or declaration)
func (s *cssTokenizer) Next() *Token {
	if s.index > len(s.tokens) {
		return nil
	}
	ret := s.tokens[s.index]
	s.index++
	return ret
}

// calcLineOffsets calculates line offsets table
// Ith table item represents byte offset of line I in the source code
func (s *cssTokenizer) calcLineOffsets(src []byte) {
	s.lineOffsets = make([]int, 0)
	s.lineOffsets = append(s.lineOffsets, 0)
	for i, ch := range src {
		if ch == '\n' {
			s.lineOffsets = append(s.lineOffsets, i+1)
		}
	}
}

// byteOffset returns byte-base offset of token located at (L, C)
func (s *cssTokenizer) byteOffset(line, column int) uint32 {
	return uint32(s.lineOffsets[line-1] + column - 1)
}

func init() {
	var factory = func() Tokenizer {
		return &cssTokenizer{}
	}
	newExtensionBasedLookup("CSS", []string{".css"}, factory)
}
