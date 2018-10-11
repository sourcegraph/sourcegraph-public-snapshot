// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

// ScanOptions specifies options for NewScanner.
type ScanOptions struct {
	Trivia bool // scan and emit whitespace and comment elements (false to ignore)
}

// NewScanner creates a new scanner for the JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L78
func NewScanner(text string, options ScanOptions) *Scanner {
	return &Scanner{
		text:    []rune(text),
		options: options,

		len:   len([]rune(text)),
		token: Unknown,
		err:   None,
	}
}

// A Scanner scans a JSON document.
type Scanner struct {
	text    []rune
	options ScanOptions

	pos         int
	len         int
	value       []rune
	tokenOffset int
	token       SyntaxKind
	err         ScanErrorCode
}

// Pos returns the current character position within the JSON input.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L52
func (s *Scanner) Pos() int { return s.pos }

// Value returns the raw JSON-encoded value of the last-scanned token.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L60
func (s *Scanner) Value() string { return string(s.value) }

// TokenOffset returns the character offset of the last-scanned token.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L64
func (s *Scanner) TokenOffset() int { return s.tokenOffset }

// TokenLength returns the length of the last-scanned token.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L68
func (s *Scanner) TokenLength() int { return s.pos - s.tokenOffset }

// Token returns the kind of the last-scanned token.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L56
func (s *Scanner) Token() SyntaxKind { return s.token }

// Err returns the error code describing the error (if any) encountered
// while scanning the last-scanned token.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L72
func (s *Scanner) Err() ScanErrorCode { return s.err }

// SetPosition sets the scanner's position and resets its other internal state.
// Subsequent calls to Scan will start from the new position.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L44
func (s *Scanner) SetPosition(newPosition int) {
	s.pos = newPosition
	s.value = nil
	s.tokenOffset = 0
	s.token = Unknown
	s.err = None
}

// Scan scans and returns the next token from the input.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L48
func (s *Scanner) Scan() SyntaxKind {
	if s.options.Trivia {
		return s.scanNext()
	}
	return s.scanNextNonTrivia()
}

func (s *Scanner) scanHexDigits(count int, exact bool) rune {
	digits := 0
	var value rune
	for digits < count || !exact {
		ch := s.text[s.pos]
		if ch >= charCode0 && ch <= charCode9 {
			value = rune(value*16) + ch - charCode0
		} else if ch >= charCodeA && ch <= charCodeF {
			value = rune(value*16) + ch - charCodeA + 10
		} else if ch >= charCodeLowerA && ch <= charCodeLowerF {
			value = rune(value*16) + ch - charCodeLowerA + 10
		} else {
			break
		}
		s.pos++
		digits++
	}
	if digits < count {
		value = -1
	}
	return value
}

func (s *Scanner) scanNumber() []rune {
	start := s.pos
	if s.text[s.pos] == charCode0 {
		s.pos++
	} else {
		s.pos++
		for s.pos < len(s.text) && isDigit(s.text[s.pos]) {
			s.pos++
		}
	}
	if s.pos < len(s.text) && s.text[s.pos] == charCodeDot {
		s.pos++
		if s.pos < len(s.text) && isDigit(s.text[s.pos]) {
			s.pos++
			for s.pos < len(s.text) && isDigit(s.text[s.pos]) {
				s.pos++
			}
		} else {
			s.err = UnexpectedEndOfNumber
			return s.text[start:s.pos]
		}
	}
	end := s.pos
	if s.pos < len(s.text) && (s.text[s.pos] == charCodeE || s.text[s.pos] == charCodeLowerE) {
		s.pos++
		if s.pos < len(s.text) && s.text[s.pos] == charCodePlus || s.text[s.pos] == charCodeMinus {
			s.pos++
		}
		if s.pos < len(s.text) && isDigit(s.text[s.pos]) {
			s.pos++
			for s.pos < len(s.text) && isDigit(s.text[s.pos]) {
				s.pos++
			}
			end = s.pos
		} else {
			s.err = UnexpectedEndOfNumber
		}
	}
	return s.text[start:end]
}

func (s *Scanner) scanString() []rune {
	var result []rune
	start := s.pos

	for {
		if s.pos >= s.len {
			result = append(result, s.text[start:s.pos]...)
			s.err = UnexpectedEndOfString
			break
		}
		ch := s.text[s.pos]
		if ch == charCodeDoubleQuote {
			result = append(result, s.text[start:s.pos]...)
			s.pos++
			break
		}
		if ch == charCodeBackslash {
			result = append(result, s.text[start:s.pos]...)
			s.pos++
			if s.pos >= s.len {
				s.err = UnexpectedEndOfString
				break
			}
			ch = s.text[s.pos]
			s.pos++
			switch ch {
			case charCodeDoubleQuote:
				result = append(result, '"')
			case charCodeBackslash:
				result = append(result, '\\')
			case charCodeSlash:
				result = append(result, '/')
			case charCodeLowerB:
				result = append(result, '\b')
			case charCodeLowerF:
				result = append(result, '\f')
			case charCodeLowerN:
				result = append(result, '\n')
			case charCodeLowerR:
				result = append(result, '\r')
			case charCodeLowerT:
				result = append(result, '\t')
			case charCodeLowerU:
				ch := s.scanHexDigits(4, true)
				if ch >= 0 {
					result = append(result, ch)
				} else {
					s.err = InvalidUnicode
				}
			default:
				s.err = InvalidEscapeCharacter
			}
			start = s.pos
			continue
		}
		if ch >= 0 && ch <= 0x1f {
			if isLineBreak(ch) {
				result = append(result, s.text[start:s.pos]...)
				s.err = UnexpectedEndOfString
				break
			} else {
				s.err = InvalidCharacter
				// mark as error but continue with string
			}
		}
		s.pos++
	}
	return result
}

// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L241
func (s *Scanner) scanNext() SyntaxKind {
	s.value = nil
	s.err = None

	s.tokenOffset = s.pos

	if s.pos >= s.len {
		// at the end
		s.tokenOffset = s.len
		s.token = EOF
		return s.token
	}

	code := s.text[s.pos]
	// trivia: whitespace
	if isWhiteSpace(code) {
		for {
			s.pos++
			s.value = append(s.value, code)
			if s.pos >= s.len {
				break
			}
			code = s.text[s.pos]
			if !isWhiteSpace(code) {
				break
			}
		}

		s.token = Trivia
		return s.token
	}

	// trivia: newlines
	if isLineBreak(code) {
		s.pos++
		s.value = append(s.value, code)
		if code == charCodeCarriageReturn && s.pos < s.len && s.text[s.pos] == charCodeLineFeed {
			s.pos++
			s.value = append(s.value, '\n')
		}
		s.token = LineBreakTrivia
		return s.token
	}

	switch code {
	// tokens: []{}:,
	case charCodeOpenBrace:
		s.pos++
		s.token = OpenBraceToken
		return s.token
	case charCodeCloseBrace:
		s.pos++
		s.token = CloseBraceToken
		return s.token
	case charCodeOpenBracket:
		s.pos++
		s.token = OpenBracketToken
		return s.token
	case charCodeCloseBracket:
		s.pos++
		s.token = CloseBracketToken
		return s.token
	case charCodeColon:
		s.pos++
		s.token = ColonToken
		return s.token
	case charCodeComma:
		s.pos++
		s.token = CommaToken
		return s.token

	// strings
	case charCodeDoubleQuote:
		s.pos++
		s.value = s.scanString()
		s.token = StringLiteral
		return s.token

	// comments
	case charCodeSlash:
		start := s.pos - 1
		// Single-line comment
		if s.text[s.pos+1] == charCodeSlash {
			s.pos += 2

			for s.pos < s.len {
				if isLineBreak(s.text[s.pos]) {
					break
				}
				s.pos++

			}
			if start == -1 {
				start = 0
			}
			s.value = s.text[start:s.pos]
			s.token = LineCommentTrivia
			return s.token
		}

		// Multi-line comment
		if s.text[s.pos+1] == charCodeAsterisk {
			s.pos += 2

			safeLength := s.len - 1 // For lookahead.
			commentClosed := false
			for s.pos < safeLength {
				ch := s.text[s.pos]

				if ch == charCodeAsterisk && s.text[s.pos+1] == charCodeSlash {
					s.pos += 2
					commentClosed = true
					break
				}
				s.pos++
			}

			if !commentClosed {
				s.pos++
				s.err = UnexpectedEndOfComment
			}

			if start == -1 {
				start = 0
			}
			s.value = s.text[start:s.pos]
			s.token = BlockCommentTrivia
			return s.token
		}
		// just a single slash
		s.value = append(s.value, code)
		s.pos++
		s.token = Unknown
		return s.token

	// numbers
	case charCodeMinus:
		s.value = append(s.value, code)
		s.pos++
		if s.pos == s.len || !isDigit(s.text[s.pos]) {
			s.token = Unknown
			return s.token
		}
		fallthrough

	// found a minus, followed by a number so
	// we fall through to proceed with scanning
	// numbers
	case charCode0, charCode1, charCode2, charCode3, charCode4, charCode5, charCode6, charCode7, charCode8, charCode9:
		s.value = append(s.value, s.scanNumber()...)
		s.token = NumericLiteral
		return s.token
	// literals and unknown symbols
	default:
		// is a literal? Read the full word.
		for s.pos < s.len && isUnknownContentCharacter(code) {
			s.pos++
			if s.pos == s.len {
				break
			}
			code = s.text[s.pos]
		}
		if s.tokenOffset != s.pos {
			s.value = s.text[s.tokenOffset:s.pos]
			// keywords: true, false, null
			switch string(s.value) {
			case "true":
				s.token = TrueKeyword
				return s.token
			case "false":
				s.token = FalseKeyword
				return s.token
			case "null":
				s.token = NullKeyword
				return s.token
			}
			s.token = Unknown
			return s.token
		}
		// some
		s.value = append(s.value, code)
		s.pos++
		s.token = Unknown
		return s.token
	}
}

func (s *Scanner) scanNextNonTrivia() SyntaxKind {
	var result SyntaxKind
	for {
		result = s.scanNext()
		if !(result >= LineCommentTrivia && result <= Trivia) {
			break
		}
	}
	return result
}

// A ScanErrorCode is a category of error that can occur while scanning a
// JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L7
type ScanErrorCode int

// Scan error codes
const (
	None ScanErrorCode = iota
	UnexpectedEndOfComment
	UnexpectedEndOfString
	UnexpectedEndOfNumber
	InvalidUnicode
	InvalidEscapeCharacter
	InvalidCharacter
)

// A SyntaxKind is a kind of syntax element in a JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L17
type SyntaxKind int

// Syntax kinds
const (
	Unknown SyntaxKind = iota
	OpenBraceToken
	CloseBraceToken
	OpenBracketToken
	CloseBracketToken
	CommaToken
	ColonToken
	NullKeyword
	TrueKeyword
	FalseKeyword
	StringLiteral
	NumericLiteral
	LineCommentTrivia
	BlockCommentTrivia
	LineBreakTrivia
	Trivia
	EOF
)

// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L436
func isWhiteSpace(ch rune) bool {
	return ch == charCodeSpace || ch == charCodeTab || ch == charCodeVerticalTab || ch == charCodeFormFeed ||
		ch == charCodeNonBreakingSpace || ch == charCodeOgham || ch >= charCodeEnQuad && ch <= charCodeZeroWidthSpace ||
		ch == charCodeNarrowNoBreakSpace || ch == charCodeMathematicalSpace || ch == charCodeIdeographicSpace || ch == charCodeByteOrderMark
}

// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L442
func isLineBreak(ch rune) bool {
	return ch == charCodeLineFeed || ch == charCodeCarriageReturn || ch == charCodeLineSeparator || ch == charCodeParagraphSeparator
}

// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L446
func isDigit(ch rune) bool {
	return ch >= charCode0 && ch <= charCode9
}

func isUnknownContentCharacter(code rune) bool {
	if isWhiteSpace(code) || isLineBreak(code) {
		return false
	}
	switch code {
	case charCodeCloseBrace, charCodeCloseBracket, charCodeOpenBrace, charCodeOpenBracket, charCodeDoubleQuote, charCodeColon, charCodeComma:
		return false
	}
	return true
}

// Character codes
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L450
const (
	charCodeNullCharacter     rune = 0
	charCodeMaxASCIICharacter rune = 0x7F

	charCodeLineFeed           rune = 0x0A // \n
	charCodeCarriageReturn     rune = 0x0D // \r
	charCodeLineSeparator      rune = 0x2028
	charCodeParagraphSeparator rune = 0x2029

	// REVIEW: do we need to support this?  The scanner doesn't, but our IText does.  This seems
	// like an odd disparity?  (Or maybe it's completely fine for them to be different).
	charCodeNextLine rune = 0x0085

	// Unicode 3.0 space characters
	charCodeSpace              rune = 0x0020 // " "
	charCodeNonBreakingSpace   rune = 0x00A0 //
	charCodeEnQuad             rune = 0x2000
	charCodeEmQuad             rune = 0x2001
	charCodeEnSpace            rune = 0x2002
	charCodeEmSpace            rune = 0x2003
	charCodeThreePerEmSpace    rune = 0x2004
	charCodeFourPerEmSpace     rune = 0x2005
	charCodeSixPerEmSpace      rune = 0x2006
	charCodeFigureSpace        rune = 0x2007
	charCodePunctuationSpace   rune = 0x2008
	charCodeThinSpace          rune = 0x2009
	charCodeHairSpace          rune = 0x200A
	charCodeZeroWidthSpace     rune = 0x200B
	charCodeNarrowNoBreakSpace rune = 0x202F
	charCodeIdeographicSpace   rune = 0x3000
	charCodeMathematicalSpace  rune = 0x205F
	charCodeOgham              rune = 0x1680

	charCodeUnderscore rune = 0x5F
	charCodeDollarSign rune = 0x24

	charCode0 rune = 0x30
	charCode1 rune = 0x31
	charCode2 rune = 0x32
	charCode3 rune = 0x33
	charCode4 rune = 0x34
	charCode5 rune = 0x35
	charCode6 rune = 0x36
	charCode7 rune = 0x37
	charCode8 rune = 0x38
	charCode9 rune = 0x39

	charCodeLowerA rune = 0x61
	charCodeLowerB rune = 0x62
	charCodeLowerC rune = 0x63
	charCodeLowerD rune = 0x64
	charCodeLowerE rune = 0x65
	charCodeLowerF rune = 0x66
	charCodeLowerG rune = 0x67
	charCodeLowerH rune = 0x68
	charCodeLowerI rune = 0x69
	charCodeLowerJ rune = 0x6A
	charCodeLowerK rune = 0x6B
	charCodeLowerL rune = 0x6C
	charCodeLowerM rune = 0x6D
	charCodeLowerN rune = 0x6E
	charCodeLowerO rune = 0x6F
	charCodeLowerP rune = 0x70
	charCodeLowerQ rune = 0x71
	charCodeLowerR rune = 0x72
	charCodeLowerS rune = 0x73
	charCodeLowerT rune = 0x74
	charCodeLowerU rune = 0x75
	charCodeLowerV rune = 0x76
	charCodeLowerW rune = 0x77
	charCodeLowerX rune = 0x78
	charCodeLowerY rune = 0x79
	charCodeLowerZ rune = 0x7A

	charCodeA rune = 0x41
	charCodeB rune = 0x42
	charCodeC rune = 0x43
	charCodeD rune = 0x44
	charCodeE rune = 0x45
	charCodeF rune = 0x46
	charCodeG rune = 0x47
	charCodeH rune = 0x48
	charCodeI rune = 0x49
	charCodeJ rune = 0x4A
	charCodeK rune = 0x4B
	charCodeL rune = 0x4C
	charCodeM rune = 0x4D
	charCodeN rune = 0x4E
	charCodeO rune = 0x4F
	charCodeP rune = 0x50
	charCodeQ rune = 0x51
	charCodeR rune = 0x52
	charCodeS rune = 0x53
	charCodeT rune = 0x54
	charCodeU rune = 0x55
	charCodeV rune = 0x56
	charCodeW rune = 0x57
	charCodeX rune = 0x58
	charCodeY rune = 0x59
	charCodeZ rune = 0x5a

	charCodeAmpersand    rune = 0x26 // &
	charCodeAsterisk     rune = 0x2A // *
	charCodeAt           rune = 0x40 // @
	charCodeBackslash    rune = 0x5C // \
	charCodeBar          rune = 0x7C // |
	charCodeCaret        rune = 0x5E // ^
	charCodeCloseBrace   rune = 0x7D // }
	charCodeCloseBracket rune = 0x5D // ]
	charCodeCloseParen   rune = 0x29 // )
	charCodeColon        rune = 0x3A // :
	charCodeComma        rune = 0x2C // ,
	charCodeDot          rune = 0x2E // .
	charCodeDoubleQuote  rune = 0x22 // "
	charCodeEquals       rune = 0x3D // =
	charCodeExclamation  rune = 0x21 // !
	charCodeGreaterThan  rune = 0x3E // >
	charCodeLessThan     rune = 0x3C // <
	charCodeMinus        rune = 0x2D // -
	charCodeOpenBrace    rune = 0x7B // {
	charCodeOpenBracket  rune = 0x5B // [
	charCodeOpenParen    rune = 0x28 // (
	charCodePercent      rune = 0x25 // %
	charCodePlus         rune = 0x2B // +
	charCodeQuestion     rune = 0x3F // ?
	charCodeSemicolon    rune = 0x3B // ;
	charCodeSingleQuote  rune = 0x27 // '
	charCodeSlash        rune = 0x2F // /
	charCodeTilde        rune = 0x7E // ~

	charCodeBackspace     rune = 0x08 // \b
	charCodeFormFeed      rune = 0x0C // \f
	charCodeByteOrderMark rune = 0xFEFF
	charCodeTab           rune = 0x09 // \t
	charCodeVerticalTab   rune = 0x0B // \v
)
