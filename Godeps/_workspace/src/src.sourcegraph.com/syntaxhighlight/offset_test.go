package syntaxhighlight

import (
	"testing"
)

func TestOffsetByGroups(t *testing.T) {

	var f RuleAction
	f = func(lexer Lexer, source []byte, offset int, matches []int) []Token {
		tok := NewToken(source[matches[0]:matches[1]], Error, offset+matches[0])
		return []Token{tok}
	}

	lexer := &RegexpLexer{rules: map[string][]RegexpRule{
		`root`: {
			MS.Action(`(foo) (bar)`, ByGroups(Name_Variable, String)),
			MS.Action(`(john) (snow)`, ByGroups(f, f)),
		},
	}}
	source := "hello foo bar goodbye john snow ..."

	tokens := GetTokens(lexer, []byte(source))

	checkTokens(tokens, []Token{
		{"foo", Name_Variable, 6},
		{"bar", String, 10},
		{"john", Error, 22},
		{"snow", Error, 27},
	}, true, t)
}
