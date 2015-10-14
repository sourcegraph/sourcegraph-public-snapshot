package syntaxhighlight

// Lexer is an object capable to produce lexer tokens
type Lexer interface {

	// This function may be called before starting producing tokens
	// with the aim to initialize lexer internal structures or reset lexer's state before re-using.
	// source is source code
	Init(source []byte)
	// Produces new token if possible or nil to indicate EOF
	NextToken() *Token
}
