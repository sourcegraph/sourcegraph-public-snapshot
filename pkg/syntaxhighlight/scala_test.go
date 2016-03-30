package syntaxhighlight

import "testing"

func TestScala_RuleActions(t *testing.T) {
	source := `private[this] class A[B](private[this] var p: C[A]) extends A {`

	scalaLexer := NewLexerByExtension(".scala")
	tokens := GetTokens(scalaLexer, []byte(source))

	checkTokens(tokens, []Token{
		{`private`, Keyword, 0}, {`[`, Operator, 7}, {`this`, Keyword_Type, 8}, {`]`, Operator, 12}, {`class`, Keyword, 14}, {`A`, Name_Class, 20}, {`[`, Operator, 21}, {`B`, Keyword_Type, 22}, {`]`, Operator, 23}, {`(`, Operator, 24}, {`private`, Keyword, 25}, {`[`, Operator, 32}, {`this`, Keyword_Type, 33}, {`]`, Operator, 37}, {`var`, Keyword, 39}, {`p`, Name, 43}, {`:`, Operator, 44}, {`C`, Name, 46}, {`[`, Operator, 47}, {`A`, Keyword_Type, 48}, {`]`, Operator, 49}, {`)`, Operator, 50}, {`extends`, Keyword, 52}, {`A`, Name, 60}, {`{`, Operator, 62},
	}, false, t)
}
