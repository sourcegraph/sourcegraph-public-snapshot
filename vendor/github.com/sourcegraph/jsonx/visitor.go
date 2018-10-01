// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

import "encoding/json"

// A Visitor has its funcs invoked by Walk as it traverses the parse tree of a
// JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L1008
type Visitor struct {
	// Invoked when an open brace is encountered and an object is started. The
	// offset and length represent the location of the open brace.
	OnObjectBegin func(offset, length int)

	// Invoked when a property is encountered. The offset and length represent
	// the location of the property name.
	OnObjectProperty func(property string, offset, length int)

	// Invoked when a closing brace is encountered and an object is completed.
	// The offset and length represent the location of the closing brace.
	OnObjectEnd func(offset, length int)

	// Invoked when an open bracket is encountered. The offset and length represent
	// the location of the open bracket.
	OnArrayBegin func(offset, length int)

	// Invoked when a closing bracket is encountered. The offset and length represent
	// the location of the closing bracket.
	OnArrayEnd func(offset, length int)

	// Invoked when a literal value is encountered. The offset and length represent
	// the location of the literal value.
	OnLiteralValue func(value interface{}, offset, length int)

	// Invoked when a comma or colon separator is encountered. The offset and length
	// represent the location of the separator.
	OnSeparator func(character rune, offset, length int)

	// Invoked on an error.
	OnError func(errorCode ParseErrorCode, offset, length int)
}

// Walk parses the JSON document text and calls the visitor's funcs for
// each object, array and literal reached.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L799
func Walk(text string, options ParseOptions, visitor Visitor) bool {
	scanner := NewScanner(text, ScanOptions{Trivia: true})
	walker := walker{scanner: scanner, options: options, visitor: visitor}
	return walker.walk()
}

type walker struct {
	scanner *Scanner
	options ParseOptions
	visitor Visitor
}

func (w *walker) walk() bool {
	w.scanNext()
	if w.scanner.Token() == EOF {
		return true
	}
	if !w.parseValue() {
		w.handleError(ValueExpected, nil, nil)
		return false
	}
	if w.scanner.Token() != EOF {
		w.handleError(EndOfFileExpected, nil, nil)
	}
	return true
}

func (w *walker) onObjectBegin() {
	if w.visitor.OnObjectBegin != nil {
		w.visitor.OnObjectBegin(w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onObjectProperty(property string) {
	if w.visitor.OnObjectProperty != nil {
		w.visitor.OnObjectProperty(property, w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onObjectEnd() {
	if w.visitor.OnObjectEnd != nil {
		w.visitor.OnObjectEnd(w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onArrayBegin() {
	if w.visitor.OnArrayBegin != nil {
		w.visitor.OnArrayBegin(w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onArrayEnd() {
	if w.visitor.OnArrayEnd != nil {
		w.visitor.OnArrayEnd(w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onLiteralValue(value interface{}) {
	if w.visitor.OnLiteralValue != nil {
		w.visitor.OnLiteralValue(value, w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onSeparator(character rune) {
	if w.visitor.OnSeparator != nil {
		w.visitor.OnSeparator(character, w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) onError(errorCode ParseErrorCode) {
	if w.visitor.OnError != nil {
		w.visitor.OnError(errorCode, w.scanner.TokenOffset(), w.scanner.TokenLength())
	}
}

func (w *walker) scanNext() SyntaxKind {
	for {
		token := w.scanner.Scan()
		switch token {
		case LineCommentTrivia, BlockCommentTrivia:
			if !w.options.Comments {
				w.handleError(InvalidSymbol, nil, nil)
			}
		case Unknown:
			w.handleError(InvalidSymbol, nil, nil)
		case Trivia, LineBreakTrivia:
		default:
			return token
		}
	}
}

func (w *walker) handleError(errorCode ParseErrorCode, skipUntilAfter, skipUntil []SyntaxKind) {
	indexOf := func(slice []SyntaxKind, candidateElement SyntaxKind) int {
		for i, e := range slice {
			if e == candidateElement {
				return i
			}
		}
		return -1
	}

	w.onError(errorCode)
	if len(skipUntilAfter)+len(skipUntil) > 0 {
		token := w.scanner.Token()
		for token != EOF {
			if indexOf(skipUntilAfter, token) != -1 {
				w.scanNext()
				break
			} else if indexOf(skipUntil, token) != -1 {
				break
			}
			token = w.scanNext()
		}
	}
}

func (w *walker) parseString(isValue bool) bool {
	value := string(w.scanner.Value())
	if isValue {
		w.onLiteralValue(value)
	} else {
		w.onObjectProperty(value)
	}
	w.scanNext()
	return true
}

func (w *walker) parseLiteral() bool {
	switch w.scanner.Token() {
	case NumericLiteral:
		value := json.Number(w.scanner.Value())
		if _, err := value.Float64(); err != nil {
			w.handleError(InvalidNumberFormat, nil, nil)
		}
		w.onLiteralValue(value)
	case NullKeyword:
		w.onLiteralValue(nil)
	case TrueKeyword:
		w.onLiteralValue(true)
	case FalseKeyword:
		w.onLiteralValue(false)
	default:
		return false
	}
	w.scanNext()
	return true
}

func (w *walker) parseProperty() bool {
	if w.scanner.Token() != StringLiteral {
		w.handleError(PropertyNameExpected, nil, []SyntaxKind{CloseBraceToken, CommaToken})
		return false
	}
	w.parseString(false)
	if w.scanner.Token() == ColonToken {
		w.onSeparator(':')
		w.scanNext() // consume colon

		if !w.parseValue() {
			w.handleError(ValueExpected, nil, []SyntaxKind{CloseBraceToken, CommaToken})
		}
	} else {
		w.handleError(ColonExpected, nil, []SyntaxKind{CloseBraceToken, CommaToken})
	}
	return true
}

func (w *walker) parseObject() bool {
	w.onObjectBegin()
	w.scanNext() // consume open brace

	needsComma := false
	for w.scanner.Token() != CloseBraceToken && w.scanner.Token() != EOF {
		if w.scanner.Token() == CommaToken {
			if !needsComma {
				w.handleError(ValueExpected, nil, nil)
			}
			w.onSeparator(',')
			w.scanNext() // consume comma
			if w.scanner.Token() == CloseBraceToken && w.options.TrailingCommas {
				break
			}
		} else if needsComma {
			w.handleError(CommaExpected, nil, nil)
		}
		if !w.parseProperty() {
			w.handleError(ValueExpected, nil, []SyntaxKind{CloseBraceToken, CommaToken})
		}
		needsComma = true
	}
	w.onObjectEnd()
	if w.scanner.Token() != CloseBraceToken {
		w.handleError(CloseBraceExpected, []SyntaxKind{CloseBraceToken}, nil)
	} else {
		w.scanNext() // consume close brace
	}
	return true
}

func (w *walker) parseArray() bool {
	w.onArrayBegin()
	w.scanNext() // consume open bracket

	needsComma := false
	for w.scanner.Token() != CloseBracketToken && w.scanner.Token() != EOF {
		if w.scanner.Token() == CommaToken {
			if !needsComma {
				w.handleError(ValueExpected, nil, nil)
			}
			w.onSeparator(',')
			w.scanNext() // consume comma
			if w.scanner.Token() == CloseBracketToken && w.options.TrailingCommas {
				break
			}
		} else if needsComma {
			w.handleError(CommaExpected, nil, nil)
		}
		if !w.parseValue() {
			w.handleError(ValueExpected, nil, []SyntaxKind{CloseBracketToken, CommaToken})
		}
		needsComma = true
	}
	w.onArrayEnd()
	if w.scanner.Token() != CloseBracketToken {
		w.handleError(CloseBracketExpected, []SyntaxKind{CloseBracketToken}, nil)
	} else {
		w.scanNext() // consume close bracket
	}
	return true
}

func (w *walker) parseValue() bool {
	switch w.scanner.Token() {
	case OpenBracketToken:
		return w.parseArray()
	case OpenBraceToken:
		return w.parseObject()
	case StringLiteral:
		return w.parseString(true)
	default:
		return w.parseLiteral()
	}
}
