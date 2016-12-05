package syntaxhighlight

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

// Matcher is a function that returns array of slices if source matches some conditions
// (an example is regexp.Regexp.FindSubmatchIndex)
// source input byte slice
// returns array of positions (start, end) where first element is position of full match and each next element is a
// position of sub-match or nil if source does not match conditions
type Matcher func(source []byte) []int

// Produces lexer rules by compiling regular expressions using specific flags (such as DOTALL, MULTILINE and so on)
type FlagsRuleMaker struct {
	// flags to be used, for example "ims"
	flags string
}

// Produces new lexer rule that returns single token if source matches given RE at the current position
// - pattern - RE pattern to use (will be combined with flags and will start with ^)
// - ttype - token type to produce
// - state - optional state(s) to apply to lexer's stack
func (self FlagsRuleMaker) Token(pattern string, ttype *TokenType, state ...string) RegexpRule {
	return RegexpRule{matcher: self.makeRegexp(pattern), ttype: ttype, states: state}
}

// Produces new lexer rule that updates state if lookahead using specific regular expression was successful.
// Position does not advance
// - pattern - RE pattern to use (will be combined with flags and will start with ^)
// - state - state(s) to apply to lexer's stack
func (self FlagsRuleMaker) Lookahead(pattern string, state ...string) RegexpRule {
	re := self.makeRegexp(pattern)
	return RegexpRule{matcher: func(source []byte) []int {
		if re(source) != nil {
			return []int{0, 0}
		}
		return nil
	}, states: state}
}

// Produces new lexer rule that consumes given characters and optionally whitespace and advances position.
// Token won't be produced
func (self FlagsRuleMaker) Consume(runes string, spaces bool) RegexpRule {
	return RegexpRule{matcher: func(source []byte) []int {
		pos := 0
		// TODO: Unicode support?
		r := rune(source[pos])
		if spaces && unicode.IsSpace(r) || strings.ContainsRune(runes, r) {
			pos++
		}
		if pos == 0 {
			return nil
		}
		return []int{0, pos}
	}, action: func(lexer Lexer, source []byte, offset int, matches []int) []Token {
		return nil
	}}
}

// Produces new lexer rule that returns token(s) if source matches given RE at the current position.
// Tokens are produced by a given function
// - pattern - RE pattern to use (will be combined with flags and will start with ^)
// - action - function that will return tokens based on current match
// - state - optional state(s) to apply to lexer's stack
func (self FlagsRuleMaker) Action(pattern string, action RuleAction, state ...string) RegexpRule {
	return RegexpRule{matcher: self.makeRegexp(pattern), action: action, states: state}
}

// Produces new lexer rule that returns single token if source matches given matcher object at the current position
// - matcher - matcher object
// - ttype - token type to produce
// - state - optional state(s) to apply to lexer's stack
func (self FlagsRuleMaker) MatcherToken(matcher Matcher, ttype *TokenType, state ...string) RegexpRule {
	return RegexpRule{matcher: matcher, ttype: ttype, states: state}
}

// Produces new lexer rule that returns token(s) if source matches given matcher object at the current position.
// Tokens are produced by a given function
// - matcher - matcher object
// - action - function that will return tokens based on current match
// - state - optional state(s) to apply to lexer's stack
func (self FlagsRuleMaker) MatcherAction(matcher Matcher, action RuleAction, state ...string) RegexpRule {
	return RegexpRule{matcher: matcher, action: action, states: state}
}

// Adjust RE to include flags and force anchor mode
func (self FlagsRuleMaker) makeRegexp(pattern string) Matcher {
	p := pattern
	if self.flags != "" {
		p = `(?` + self.flags + `)` + p
	}
	return regexp.MustCompile(`\A` + p).FindSubmatchIndex
}

var (
	// Predefined rule maker (no flags)
	F = FlagsRuleMaker{""}
	// Predefined rule maker (multiline, case-insensitive, dotall)
	MSI = FlagsRuleMaker{"msi"}
	// Predefined rule maker (multiline, dotall)
	MS = FlagsRuleMaker{"ms"}
	// Predefined rule maker (multiline)
	M = FlagsRuleMaker{"m"}
)

// Function that may produce zero or more tokens based on a given matcher
// - lexer current lexer object
// - source source code slice starting from 'offset' position
// - offset current offset
// - matches array of spans first pair is a span (start, end + 1) of full match in current source code slice and
// each next pair is span of matcher's group #X
type RuleAction func(lexer Lexer, source []byte, offset int, matches []int) []Token

// Defines lexer rule
type RegexpRule struct {
	// matcher object that will identify if rule matches source at the current position
	matcher Matcher
	// if rule produces fixed token, holds token type to produce
	ttype *TokenType
	// optional list of states to be applied to lexer's stack
	states []string
	// function to produce tokens instead of producing single one
	action RuleAction
	// state to include
	include string
}

// Includes rules defined in a given state to current set of rules
func Include(include string) RegexpRule {
	return RegexpRule{include: include}
}

// Indicates a state or state action (e.g. #pop) to apply. Does not produce tokens
func Default(states ...string) RegexpRule {
	return RegexpRule{matcher: func(source []byte) []int {
		return []int{0, 0}
	}, states: states}
}

// Either produces single token per each submatch (if argument is a TokenType) or calls custom functions to
// produce tokens from each sub-match (if argument is a function)
func ByGroups(args ...interface{}) RuleAction {
	return func(lexer Lexer, source []byte, offset int, matches []int) []Token {
		l := len(args)
		ret := make([]Token, 0, l)
		for i := 0; i < l; i++ {
			var t Token
			arg := args[i]
			vf := reflect.ValueOf(arg)
			ftype := vf.Type()
			spanIndex := 2 * (i + 1)
			if ftype.Kind() == reflect.Func {
				ret = append(ret,
					arg.(RuleAction)(lexer,
						source[matches[spanIndex]:matches[spanIndex+1]],
						offset+matches[spanIndex],
						[]int{0, matches[spanIndex+1] - matches[spanIndex]})...)
			} else {
				t = NewToken(source[matches[spanIndex]:matches[spanIndex+1]],
					arg.(*TokenType),
					offset+matches[spanIndex])
				ret = append(ret, t)
			}
		}
		return ret
	}
}

// Produces tokens using current lexer from a substring of source code
func UsingThis() RuleAction {
	return func(lexer Lexer, source []byte, offset int, matches []int) []Token {
		l := &RegexpLexer{rules: lexer.(*RegexpLexer).rules}
		tokens := GetTokens(l, source[matches[0]:matches[1]])
		for i := range tokens {
			tokens[i].Offset += offset
		}
		return tokens
	}
}

// Produces matcher that matches any of the words specified
func Words(words ...string) Matcher {
	return WordsWithBoundary(true, words...)
}

// Produces matcher that matches any of the words specified optionally followed by word boundary
// - boundary - identifies if word should be followed by word boundary
// - words - dictionary
//
// This method should only be called via an init() function at the
// top-level, because it will panic on failure. Otherwise, there may
// be unforeseen runtime failures.
func WordsWithBoundary(boundary bool, words ...string) Matcher {
	t := newTrie()
	for _, w := range words {
		err := t.insert(w)
		if err != nil {
			panic(err)
		}
	}
	return func(source []byte) []int {
		ret := t.lookup(source, func(len int) bool {
			return !boundary || isEndOfWord(source, len)
		})
		if ret > 0 {
			return []int{0, ret}
		}
		return nil
	}
}

// Produces matcher that matches the word specified
func Word(word string) Matcher {
	return WordWithBoundary(word, true)
}

// Produces matcher that matches the word specified optionally followed by word boundary
// - boundary - identifies if word should be followed by word boundary
// - word - word (prefix) to look for
func WordWithBoundary(word string, boundary bool) Matcher {
	p := []byte(word)
	l := len(word)
	return func(source []byte) []int {
		if bytes.HasPrefix(source, p) {
			if !boundary || isEndOfWord(source, l) {
				return []int{0, l}
			}
			return nil
		}
		return nil
	}
}

// Combines given Unicode classes into RE (\p{A}\p{B}...\p{Z}
func UnicodeClasses(classes ...string) string {
	ret := ""
	for _, class := range classes {
		ret += `\p{` + class + `}`
	}
	return ret
}

// Returns true if source at a given offset denotes a word boundary.
// Please note that we only taking into account subset of RE \b rules:
// word boundary is anything except A-Za-z0-9 or end-of-text
func isEndOfWord(source []byte, offset int) bool {
	if len(source) <= offset {
		return true
	}
	c := source[offset]
	if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' {
		return false
	}
	// TODO: add Unicode support
	return true
}
