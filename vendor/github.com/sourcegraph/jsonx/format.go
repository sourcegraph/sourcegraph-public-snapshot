// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

import (
	"strings"
)

// FormatOptions specifies formatting options.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts#L9
type FormatOptions struct {
	TabSize      int    // If indentation is based on spaces (InsertSpaces == true), then what is the number of spaces that make an indent?
	InsertSpaces bool   // Is indentation based on spaces?
	EOL          string // The default end of line line character
}

// Format returns edits that format the entire JSON document according to the format
// options. To apply the edits and obtain the formatted document content, use ApplyEdits.
func Format(text string, options FormatOptions) []Edit {
	return FormatRange(text, 0, len([]rune(text)), options)
}

// FormatRange returns edits that format the JSON document (starting at the character
// offset and continuing for the character length) according to the format options.
// To apply the edits and obtain the formatted document content, use ApplyEdits.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts#L41
func FormatRange(text string, offset, length int, options FormatOptions) []Edit {
	chars := []rune(text)

	rangeStart := offset
	rangeEnd := rangeStart + length
	for rangeStart > 0 && !isEOL(chars, rangeStart-1) {
		rangeStart--
	}

	{
		scanner := NewScanner(text, ScanOptions{Trivia: false})
		scanner.SetPosition(rangeEnd)
		scanner.Scan()
		rangeEnd = scanner.Pos()
	}

	value := chars[rangeStart:rangeEnd]
	initialIndentLevel := computeIndentLevel(value, 0, options)

	eol := getEOL(options, chars)

	lineBreak := false
	indentValue := ""
	if options.InsertSpaces {
		indentValue = strings.Repeat(" ", options.TabSize)
	} else {
		indentValue = "\t"
	}

	scanner := NewScanner(string(value), ScanOptions{Trivia: true})
	formatter := formatter{
		input:              chars,
		scanner:            scanner,
		eol:                eol,
		indentLevel:        0,
		indentValue:        indentValue,
		initialIndentLevel: initialIndentLevel,
		lineBreak:          lineBreak,
	}
	return formatter.format(rangeStart, rangeEnd)
}

type formatter struct {
	input              []rune
	scanner            *Scanner
	eol                string
	indentLevel        int
	indentValue        string
	initialIndentLevel int
	lineBreak          bool

	edits []Edit
}

func (f *formatter) format(rangeStart, rangeEnd int) (editOperations []Edit) {
	firstToken := f.scanNext()
	if firstToken != EOF {
		firstTokenStart := f.scanner.TokenOffset() + rangeStart
		initialIndent := strings.Repeat(f.indentValue, f.initialIndentLevel)
		f.addEdit(initialIndent, rangeStart, firstTokenStart)
	}

	for firstToken != EOF {
		firstTokenEnd := f.scanner.TokenOffset() + f.scanner.TokenLength() + rangeStart
		secondToken := f.scanNext()

		replaceContent := ""
		for !f.lineBreak && (secondToken == LineCommentTrivia || secondToken == BlockCommentTrivia) {
			// comments on the same line: keep them on the same line, but ignore them otherwise
			commentTokenStart := f.scanner.TokenOffset() + rangeStart
			f.addEdit(" ", firstTokenEnd, commentTokenStart)
			firstTokenEnd = f.scanner.TokenOffset() + f.scanner.TokenLength() + rangeStart
			if secondToken == LineCommentTrivia {
				replaceContent = f.newLineAndIndent()
			} else {
				replaceContent = ""
			}
			secondToken = f.scanNext()
		}

		if secondToken == CloseBraceToken {
			if firstToken != OpenBraceToken {
				f.indentLevel--
				replaceContent = f.newLineAndIndent()
			}
		} else if secondToken == CloseBracketToken {
			if firstToken != OpenBracketToken {
				f.indentLevel--
				replaceContent = f.newLineAndIndent()
			}
		} else if secondToken != EOF {
			switch firstToken {
			case OpenBracketToken, OpenBraceToken:
				f.indentLevel++
				replaceContent = f.newLineAndIndent()
			case CommaToken, LineCommentTrivia:
				replaceContent = f.newLineAndIndent()
				break
			case BlockCommentTrivia:
				if f.lineBreak {
					replaceContent = f.newLineAndIndent()
				} else {
					// symbol following comment on the same line: keep on same line, separate with ' '
					replaceContent = " "
				}
			case ColonToken:
				replaceContent = " "
			case NullKeyword, TrueKeyword, FalseKeyword, NumericLiteral:
				if secondToken == NullKeyword || secondToken == FalseKeyword || secondToken == NumericLiteral {
					replaceContent = " "
				}
			}
			if f.lineBreak && (secondToken == LineCommentTrivia || secondToken == BlockCommentTrivia) {
				replaceContent = f.newLineAndIndent()
			}

		}
		secondTokenStart := f.scanner.TokenOffset() + rangeStart
		f.addEdit(replaceContent, firstTokenEnd, secondTokenStart)
		firstToken = secondToken
	}
	return f.edits
}

func (f *formatter) newLineAndIndent() string {
	n := f.initialIndentLevel + f.indentLevel
	if n < 0 {
		n = 0
	}
	return f.eol + strings.Repeat(f.indentValue, n)
}

func (f *formatter) scanNext() SyntaxKind {
	token := f.scanner.Scan()
	f.lineBreak = false
	for token == Trivia || token == LineBreakTrivia {
		f.lineBreak = f.lineBreak || (token == LineBreakTrivia)
		token = f.scanner.Scan()
	}
	return token
}

func (f *formatter) addEdit(text string, startOffset, endOffset int) {
	if string(f.input[startOffset:endOffset]) != text {
		f.edits = append(f.edits, Edit{Offset: startOffset, Length: endOffset - startOffset, Content: text})
	}
}

func computeIndentLevel(chars []rune, offset int, options FormatOptions) int {
	i := 0
	nChars := 0
	tabSize := options.TabSize
	if tabSize == 0 {
		tabSize = 4
	}
	for i < len(chars) {
		ch := chars[i]
		if ch == ' ' {
			nChars++
		} else if ch == '\t' {
			nChars += tabSize
		} else {
			break
		}
		i++
	}
	return nChars / tabSize
}

func getEOL(options FormatOptions, chars []rune) string {
	for i, ch := range chars {
		if ch == '\r' {
			if i+1 < len(chars) && chars[i+1] == '\n' {
				return "\r\n"
			}
			return "\r"
		} else if ch == '\n' {
			return "\n"
		}
	}
	if options.EOL != "" {
		return options.EOL
	}
	return "\n"
}

func isEOL(chars []rune, offset int) bool {
	return chars[offset] == '\r' || chars[offset] == '\n'
}
