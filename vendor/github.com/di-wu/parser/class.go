package parser

import (
	"github.com/di-wu/parser/op"
	"strconv"
	"strings"
	"unicode"
)

// AnonymousClass represents an anonymous Class.Check function.
//
// The cursor should never be nil except if it fails at the first rune.
// e.g. "121".Check("123") should return a mark to the 2nd value.
type AnonymousClass func(p *Parser) (*Cursor, bool)

// CheckInteger returns an AnonymousClass that checks whether the following
// runes are equal to the given integer. It also consumes leading zeros
// when indicated to do so.
func CheckInteger(i int, leadingZeros bool) AnonymousClass {
	// Edge case: i == 0.
	if i == 0 {
		// Consume all zeroes if leadingZeros == true.
		if leadingZeros {
			return func(p *Parser) (*Cursor, bool) {
				return p.Check(op.MinOne('0'))
			}
		}
		return func(p *Parser) (*Cursor, bool) {
			return p.Check('0')
		}
	}

	var and op.And
	str := strconv.Itoa(i)
	// Negative integers.
	if i < 0 {
		and = append(and, '-')
		str = strings.TrimPrefix(str, "-")
	}
	// Leading zeroes.
	if leadingZeros {
		// Consume all leading zeros.
		and = append(and, op.MinZero('0'))
	}
	return func(p *Parser) (*Cursor, bool) {
		return p.Check(append(and, str))
	}
}

// CheckIntegerRange returns an AnonymousClass that checks whether the following
// runes are inside the given range (inclusive). It also consumes leading zeros
// when indicated to do so.
//
// Note that this check consumes all the sequential numbers it possibly can.
// e.g. "12543" is not in the range (0, 12345), even the prefix "1254" is.
func CheckIntegerRange(min, max uint, leadingZeros bool) AnonymousClass {
	return func(p *Parser) (*Cursor, bool) {
		digit := CheckRuneRange('0', '9')
		check := CheckRuneRange('1', '9')
		if leadingZeros {
			check = digit
		}

		var last *Cursor
		var str string
		for r, ok := p.Check(check); ok; r, ok = p.Check(digit) {
			str += string(r.Rune)
			last = r
		}
		i, _ := strconv.Atoi(str)
		if i := uint(i); min <= i && i <= max {
			return last, true
		}
		return nil, false
	}
}

// CheckRune returns an AnonymousClass that checks whether the current rune of
// the parser matches the given rune. The same result can be achieved by using
// p.Expect(r). Where 'p' is a reference to the parser an 'r' a rune value.
func CheckRune(expected rune) AnonymousClass {
	return CheckRuneFunc(func(actual rune) bool {
		return expected == actual
	})
}

// CheckRuneCI returns an AnonymousClass that checks whether the current (lower
// cased) rune of the parser matches the given (lower cased) rune. The given
// rune does not need to be lower case.
func CheckRuneCI(expected rune) AnonymousClass {
	expected = unicode.ToLower(expected)
	return CheckRuneFunc(func(actual rune) bool {
		return expected == unicode.ToLower(actual)
	})
}

// CheckRuneRange returns an AnonymousClass that checks whether the current rune of
// the parser is inside the given range (inclusive).
func CheckRuneRange(min, max rune) AnonymousClass {
	return CheckRuneFunc(func(r rune) bool {
		return min <= r && r <= max
	})
}

// CheckRuneFunc returns an AnonymousClass that checks whether the current rune of
// the parser matches the given validator.
func CheckRuneFunc(f func(r rune) bool) AnonymousClass {
	return func(p *Parser) (*Cursor, bool) {
		return p.Mark(), f(p.Current())
	}
}

// CheckString returns an AnonymousClass that checks whether the current
// sequence runes of the parser matches the given string. The same result can be
// achieved by using p.Expect(s). Where 'p' is a reference to the parser an 's'
// a string value.
func CheckString(s string) AnonymousClass {
	return func(p *Parser) (*Cursor, bool) {
		var last *Cursor
		for _, r := range []rune(s) {
			if p.Current() != r {
				return last, false
			}
			last = p.Mark()
			p.Next()
		}
		return last, true
	}
}

// CheckStringCI returns an AnonymousClass that checks whether the current
// (lower cased) sequence runes of the parser matches the given (lower cased)
// string. The given string does not need to be lower case.
func CheckStringCI(s string) AnonymousClass {
	s = strings.ToLower(s)
	return func(p *Parser) (*Cursor, bool) {
		var last *Cursor
		for _, r := range []rune(s) {
			if unicode.ToLower(p.Current()) != r {
				return last, false
			}
			last = p.Mark()
			p.Next()
		}
		return last, true
	}
}

// Class provides an interface for checking classes.
type Class interface {
	// Check should return the last p.Mark() that matches the class. It should
	// also return whether it was able to check the whole class.
	//
	// e.g. if the class is defined as follows: '<=' / '=>'. Then a parser that
	// only contains '=' will not match this class and return 'nil, false'.
	Check(p *Parser) (*Cursor, bool)
}
