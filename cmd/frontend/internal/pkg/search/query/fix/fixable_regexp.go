package fix

import (
	"fmt"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode"
)

type FixableRegexp struct {
	value string
	re    *syntax.Regexp
	Err   error
}

func NewFixableRegexp(val string) *FixableRegexp {
	r, err := syntax.Parse(val, syntax.Perl)

	return &FixableRegexp{val, r, err}
}

func (fr *FixableRegexp) Fix() {
	if fr.Err != nil {
		val := fr.value
		err := fr.Err

		fr.fixupCompileErrors()
		if fr.Err != nil {
			// Set to the original values
			fr.value = val
			fr.Err = err
			return
		}

		nfr := NewFixableRegexp(fr.value)
		fr.re = nfr.re
		fr.value = nfr.value
		fr.Err = nfr.Err
	}

	escapeNonTerminalEOL(fr.re)
}

func escapeNonTerminalEOL(r *syntax.Regexp) {
	if r.Op == syntax.OpConcat && len(r.Sub) > 0 {
		for i, child := range r.Sub[:len(r.Sub)-1] {
			if child.Op == syntax.OpEndLine || child.Op == syntax.OpEndText {
				r.Sub[i] = &syntax.Regexp{
					Op:   syntax.OpLiteral,
					Rune: []rune{'$'},
				}
			}
		}
	}
	for _, child := range r.Sub {
		escapeNonTerminalEOL(child)
	}
}

// flipRune maps opening block characters (e.g. ), ]) to their opening
// counterparts. If the rune provided is not one of those, this func returns
// the identity of the rune.
func flipRune(r rune) rune {
	switch r {
	case ')':
		return '('
	case ']':
		return '['
	default:
		return r
	}
}

var escapeErrorMessages = []struct {
	message         string
	getRuneToEscape func(rune) rune
}{
	{"missing closing ", flipRune},
	{"missing argument to repetition operator: `", func(r rune) rune { return r }},
	{"unexpected ", func(r rune) rune { return r }},
}

func (fr *FixableRegexp) fixupCompileErrors() {
	msg := fr.Err.Error()
	var runeToEscape rune

	for _, errorMsg := range escapeErrorMessages {
		index := strings.Index(msg, errorMsg.message)
		if index > -1 {
			index = len(errorMsg.message) + index

			runeToEscape = errorMsg.getRuneToEscape(rune(msg[index]))

			break
		}
	}

	if runeToEscape == 0 {
		return
	}

	out := ""
	// Loop through and escape all runeToEscape
	for _, r := range fr.value {
		if r == runeToEscape {
			out += string('\\') + string(r)
		} else {
			out += string(r)
		}
	}

	fr.value = out
	fr.Err = nil
}

func (fr *FixableRegexp) String() string {
	if fr.Err != nil {
		return fr.value
	}

	var b strings.Builder
	writeRegexp(&b, fr.re)
	return b.String()
}

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
		} else if re.Rune[0] == 0 && re.Rune[len(re.Rune)-1] == unicode.MaxRune {
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
			fmt.Println(1, re.Rune)

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
		b.WriteString(".")
	case syntax.OpAnyChar:
		b.WriteString(".")
	case syntax.OpBeginLine:
		b.WriteString("^")
	case syntax.OpEndLine:
		b.WriteString("$")
	case syntax.OpBeginText:
		b.WriteString("^")
	case syntax.OpEndText:
		b.WriteString("$")
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
				writeRegexp(b, re.Sub[0])
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
