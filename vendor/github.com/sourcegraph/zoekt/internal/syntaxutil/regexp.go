// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntaxutil

import (
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode"
)

// Note to implementers:
// In this package, re is always a *Regexp and r is always a rune.

// writeRegexp writes the Perl syntax for the regular expression re to b.
func writeRegexp(b *strings.Builder, re *syntax.Regexp) {
	switch re.Op {
	default:
		b.WriteString("<invalid op" + strconv.Itoa(int(re.Op)) + ">")
	case syntax.OpNoMatch:
		b.WriteString(`[^\x00-\x{10FFFF}]`)
	case syntax.OpEmptyMatch:
		b.WriteString(`(?:)`)
	case syntax.OpLiteral:
		if re.Flags&syntax.FoldCase != 0 {
			b.WriteString(`(?i:`)
		}
		for _, r := range re.Rune {
			escape(b, r, false)
		}
		if re.Flags&syntax.FoldCase != 0 {
			b.WriteString(`)`)
		}
	case syntax.OpCharClass:
		if len(re.Rune)%2 != 0 {
			b.WriteString(`[invalid char class]`)
			break
		}
		b.WriteRune('[')
		if len(re.Rune) == 0 {
			b.WriteString(`^\x00-\x{10FFFF}`)
		} else if re.Rune[0] == 0 && re.Rune[len(re.Rune)-1] == unicode.MaxRune && len(re.Rune) > 2 {
			// Contains 0 and MaxRune. Probably a negated class.
			// Print the gaps.
			b.WriteRune('^')
			for i := 1; i < len(re.Rune)-1; i += 2 {
				lo, hi := re.Rune[i]+1, re.Rune[i+1]-1
				escape(b, lo, lo == '-')
				if lo != hi {
					b.WriteRune('-')
					escape(b, hi, hi == '-')
				}
			}
		} else {
			for i := 0; i < len(re.Rune); i += 2 {
				lo, hi := re.Rune[i], re.Rune[i+1]
				escape(b, lo, lo == '-')
				if lo != hi {
					b.WriteRune('-')
					escape(b, hi, hi == '-')
				}
			}
		}
		b.WriteRune(']')
	case syntax.OpAnyCharNotNL:
		b.WriteString(`(?-s:.)`)
	case syntax.OpAnyChar:
		b.WriteString(`(?s:.)`)
	case syntax.OpBeginLine:
		b.WriteString(`(?m:^)`)
	case syntax.OpEndLine:
		b.WriteString(`(?m:$)`)
	case syntax.OpBeginText:
		b.WriteString(`\A`)
	case syntax.OpEndText:
		if re.Flags&syntax.WasDollar != 0 {
			b.WriteString(`(?-m:$)`)
		} else {
			b.WriteString(`\z`)
		}
	case syntax.OpWordBoundary:
		b.WriteString(`\b`)
	case syntax.OpNoWordBoundary:
		b.WriteString(`\B`)
	case syntax.OpCapture:
		if re.Name != "" {
			b.WriteString(`(?P<`)
			b.WriteString(re.Name)
			b.WriteRune('>')
		} else {
			b.WriteRune('(')
		}
		if re.Sub[0].Op != syntax.OpEmptyMatch {
			writeRegexp(b, re.Sub[0])
		}
		b.WriteRune(')')
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		if sub := re.Sub[0]; sub.Op > syntax.OpCapture || sub.Op == syntax.OpLiteral && len(sub.Rune) > 1 {
			b.WriteString(`(?:`)
			writeRegexp(b, sub)
			b.WriteString(`)`)
		} else {
			writeRegexp(b, sub)
		}
		switch re.Op {
		case syntax.OpStar:
			b.WriteRune('*')
		case syntax.OpPlus:
			b.WriteRune('+')
		case syntax.OpQuest:
			b.WriteRune('?')
		case syntax.OpRepeat:
			b.WriteRune('{')
			b.WriteString(strconv.Itoa(re.Min))
			if re.Max != re.Min {
				b.WriteRune(',')
				if re.Max >= 0 {
					b.WriteString(strconv.Itoa(re.Max))
				}
			}
			b.WriteRune('}')
		}
		if re.Flags&syntax.NonGreedy != 0 {
			b.WriteRune('?')
		}
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			if sub.Op == syntax.OpAlternate {
				b.WriteString(`(?:`)
				writeRegexp(b, sub)
				b.WriteString(`)`)
			} else {
				writeRegexp(b, sub)
			}
		}
	case syntax.OpAlternate:
		for i, sub := range re.Sub {
			if i > 0 {
				b.WriteRune('|')
			}
			writeRegexp(b, sub)
		}
	}
}

func RegexpString(re *syntax.Regexp) string {
	var b strings.Builder
	writeRegexp(&b, re)
	return b.String()
}

const meta = `\.+*?()|[]{}^$`

func escape(b *strings.Builder, r rune, force bool) {
	if unicode.IsPrint(r) {
		if strings.ContainsRune(meta, r) || force {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
		return
	}

	switch r {
	case '\a':
		b.WriteString(`\a`)
	case '\f':
		b.WriteString(`\f`)
	case '\n':
		b.WriteString(`\n`)
	case '\r':
		b.WriteString(`\r`)
	case '\t':
		b.WriteString(`\t`)
	case '\v':
		b.WriteString(`\v`)
	default:
		if r < 0x100 {
			b.WriteString(`\x`)
			s := strconv.FormatInt(int64(r), 16)
			if len(s) == 1 {
				b.WriteRune('0')
			}
			b.WriteString(s)
			break
		}
		b.WriteString(`\x{`)
		b.WriteString(strconv.FormatInt(int64(r), 16))
		b.WriteString(`}`)
	}
}
