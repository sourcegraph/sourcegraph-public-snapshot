package scanner

import (
	"bytes"
	"fmt"
	"io"
	"text/scanner"
	"unicode"
)

// Token is either a (negative) token type constant or a rune.
type Token int

// The result of Scan is one of these tokens or a Unicode character.
const (
	Nothing Token = -(iota + 1)
	EOF
	Space
	Ident
	Number
	Comment
	Operator
)

var tokenString = map[Token]string{
	EOF:      "EOF",
	Nothing:  "Nothing",
	Space:    "Space",
	Ident:    "Ident",
	Number:   "Number",
	Comment:  "Comment",
	Operator: "Operator",
}

// TokenString returns a printable string for a token or Unicode character.
func TokenString(tok Token) string {
	if s, found := tokenString[tok]; found {
		return s
	}
	return fmt.Sprintf("%q", tok)
}

type transitionFunc func(*Scanner) error

// A Scanner implements reading of Unicode characters and tokens from an io.Reader.
type Scanner struct {
	is      scanner.Scanner
	hasNext bool
	ch      rune
	pos     scanner.Position
	tok     Token
	text    string
	buf     bytes.Buffer
	tf      transitionFunc
}

// All transition functions are built up from the following primitives.
func (s *Scanner) peek() rune {
	if !s.hasNext {
		s.pos = s.is.Pos()
		s.ch = s.is.Next()
		s.hasNext = true
	}
	return s.ch
}

func (s *Scanner) consume() {
	s.hasNext = false
}

func (s *Scanner) accept() {
	s.buf.WriteRune(s.ch)
}

func (s *Scanner) emit(tok Token) {
	s.tok = tok
	s.text = s.buf.String()
	s.buf.Reset()
}

func (s *Scanner) transition(tf transitionFunc) {
	s.tf = tf
}

func isOperator(ch rune) bool {
	return unicode.IsSymbol(ch) || unicode.IsPunct(ch)
}

func isWordChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// Init initializes a Scanner with a new source and returns s.
func (s *Scanner) Init(src io.Reader) *Scanner {
	s.is.Init(src)
	s.tf = tfLineStart
	return s
}

// Scan reads the next token or Unicode character from source and returns it.
func (s *Scanner) Scan() (Token, error) {
	for {
		s.tok = Nothing
		err := s.tf(s)
		if err != nil {
			return Nothing, err
		}
		if s.tok != Nothing {
			return s.tok, nil
		}
	}
}

func tfEnd(s *Scanner) error {
	s.emit(EOF)
	return nil
}

func tfLineStart(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.transition(tfEnd)
	case ch == '\n':
		s.consume()
	case ch == '#':
		s.transition(tfComment)
	case unicode.IsSpace(ch):
		s.transition(tfSpace)
	case unicode.IsLetter(ch) || ch == '_':
		s.transition(tfIdent)
	case unicode.IsDigit(ch):
		s.transition(tfNumber)
	case isOperator(ch):
		s.transition(tfOperator)
	default:
		return fmt.Errorf("tfLineStart found unexpected character: '%q'", ch)
	}
	return nil
}

func tfIdent(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.emit(Number)
		s.transition(tfEnd)
	case ch == '\n':
		s.emit(Ident)
		s.transition(tfLineStart)
	case ch == '#':
		s.emit(Ident)
		s.transition(tfComment)
	case unicode.IsSpace(ch):
		s.emit(Ident)
		s.transition(tfSpace)
	// TODO: make sure there are no overlapping cases
	case isWordChar(ch):
		s.consume()
		s.accept()
	case isOperator(ch):
		s.emit(Ident)
		s.transition(tfOperator)
	default:
		return fmt.Errorf("tfIdent found unexpected character: '%q'", ch)
	}
	return nil
}

func tfNumber(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.emit(Number)
		s.transition(tfEnd)
	case ch == '\n':
		s.emit(Number)
		s.transition(tfLineStart)
	case ch == '#':
		s.emit(Number)
		s.transition(tfComment)
	case unicode.IsSpace(ch):
		s.emit(Number)
		s.transition(tfSpace)
	case unicode.IsDigit(ch):
		s.consume()
		s.accept()
	case isWordChar(ch):
		s.emit(Number)
		s.transition(tfIdent)
	case isOperator(ch):
		s.emit(Number)
		s.transition(tfOperator)
	default:
		return fmt.Errorf("tfNumber found unexpected character: '%q'", ch)
	}
	return nil
}

func tfSpace(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.emit(Space)
		s.transition(tfEnd)
	case ch == '\n':
		s.emit(Space)
		s.transition(tfLineStart)
	case ch == '#':
		s.emit(Space)
		s.transition(tfComment)
	case unicode.IsSpace(ch):
		s.consume()
		s.accept()
	case unicode.IsDigit(ch):
		s.emit(Space)
		s.transition(tfNumber)
	case isWordChar(ch):
		s.emit(Space)
		s.transition(tfIdent)
	case isOperator(ch):
		s.emit(Space)
		s.transition(tfOperator)
	default:
		return fmt.Errorf("tfSpace found unexpected character: '%q'", ch)
	}
	return nil
}

func tfComment(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.emit(Comment)
		s.transition(tfEnd)
	case ch == '\n':
		s.emit(Comment)
		s.transition(tfLineStart)
	default:
		s.consume()
		s.accept()
	}
	return nil
}

func tfOperator(s *Scanner) error {
	switch ch := s.peek(); {
	case ch == scanner.EOF:
		s.emit(Operator)
		s.transition(tfEnd)
	case ch == '\n':
		s.emit(Operator)
		s.transition(tfLineStart)
	case ch == '#':
		s.emit(Operator)
		s.transition(tfComment)
	case unicode.IsSpace(ch):
		s.emit(Operator)
		s.transition(tfSpace)
	case unicode.IsDigit(ch):
		s.emit(Operator)
		s.transition(tfNumber)
	case isOperator(ch):
		s.consume()
		s.accept()
	case isWordChar(ch):
		s.emit(Operator)
		s.transition(tfIdent)
	default:
		return fmt.Errorf("tfOperator found unexpected character: '%q'", ch)
	}
	return nil
}

// Pos returns the position of the character immediately after
// the character or token returned by the last call to Next or Scan.
func (s *Scanner) Pos() (pos scanner.Position) {
	return s.pos
}

// TokenText returns the string corresponding to the most recently scanned token.
// Valid after calling Scan().
func (s *Scanner) TokenText() string {
	return s.text
}
