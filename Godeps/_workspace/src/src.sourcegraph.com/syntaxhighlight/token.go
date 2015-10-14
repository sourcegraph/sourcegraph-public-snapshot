package syntaxhighlight

import "fmt"

// Source code token produced by source code lexer
type Token struct {
	// Token's text (for example "public")
	Text string `json:"text,omitempty"`
	// Token's type (for example Keyword)
	Type *TokenType `json:"type,omitempty"`
	// Byte offset relative to source code's start of token occurrence
	Offset int `json:"offset"`
}

// Constructs new token from a slice of bytes (text), token type, and offset
func NewToken(text []byte, ttype *TokenType, offset int) Token {
	return Token{string(text), ttype, offset}
}

// String representation of token
func (self Token) String() string {
	return fmt.Sprintf("{`%s` [%s] at %d}", self.Text, self.Type.Name, self.Offset)
}
