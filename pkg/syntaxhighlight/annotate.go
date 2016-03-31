package syntaxhighlight

import "github.com/sourcegraph/annotate"

// Annotator receives tokens from lexer and converts them into annotations
type Annotator interface {
	// Initializes annotator if needed, called before feeding first lexer's token
	Init() error
	// Function takes token as an input and converts it to annotation if possible
	Annotate(token Token) (*annotate.Annotation, error)
	// Allows to free resources or flush unsaved data, called when there are no more tokens left
	Done() error
}

// Annotate scans source code with the lexer and annotator.
func Annotate(src []byte, lexer Lexer, annotator Annotator) (annotate.Annotations, error) {
	annotations := make(annotate.Annotations, 0, 100)

	lexer.Init(src)
	err := annotator.Init()
	if err != nil {
		return nil, err
	}
	pos := 0
	t := lexer.NextToken()
	for t != nil {
		if pos < t.Offset {
			a, err := annotator.Annotate(NewToken(src[pos:t.Offset], Whitespace, pos))
			if err != nil {
				return nil, err
			}
			annotations = append(annotations, a)
		}
		a, err := annotator.Annotate(*t)
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
		pos = t.Offset + len(t.Text)
		t = lexer.NextToken()
	}
	if pos < len(src)-1 {
		a, err := annotator.Annotate(NewToken(src[pos:], Whitespace, pos))
		if err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}

	err = annotator.Done()
	if err != nil {
		return nil, err
	}
	return annotations, nil
}

// Token collector receives tokens produced by lexer plus whitespaces (text between tokens)
type TokenCollectorAnnotator struct {
	Tokens []Token
}

// Initializes token collector
func (self *TokenCollectorAnnotator) Init() error {
	self.Tokens = make([]Token, 0, 100)
	return nil
}

// Shuts down token collector
func (self *TokenCollectorAnnotator) Done() error {
	return nil
}

// Token collector simply appends tokens to the list
func (self *TokenCollectorAnnotator) Annotate(token Token) (*annotate.Annotation, error) {
	self.Tokens = append(self.Tokens, token)
	return nil, nil
}
