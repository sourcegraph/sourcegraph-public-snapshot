package syntaxhighlight

import (
	"fmt"
	"testing"
)

var goLexerInstance Lexer

func init() {
	goLexerInstance = NewLexerByExtension(`.go`)
}

func TestGoOffsets(t *testing.T) {

	source := "package a\nfunc init() {x := `ы`}"

	tokens := GetTokens(goLexerInstance, []byte(source))

	checkTokens(tokens, []Token{
		{"package", Keyword_Namespace, 0},
		{"a", Name_Other, 8},
		{"func", Keyword_Declaration, 10},
		{"init", Name_Attribute, 15},
		{"(", Punctuation, 19},
		{")", Punctuation, 20},
		{"{", Punctuation, 22},
		{"x", Name_Other, 23},
		{":=", Operator, 25},
		{"`ы`", String, 28},
		{"}", Punctuation, 32},
	}, true, t)
}

func TestGoInts(t *testing.T) {

	source := `
0x7A
0XFF
007
0
10
    `

	tokens := GetTokens(goLexerInstance, []byte(source))

	checkTokens(tokens, []Token{
		{"0x7A", Number_Hex, 0},
		{"0XFF", Number_Hex, 0},
		{"007", Number_Oct, 0},
		{"0", Number_Integer, 0},
		{"10", Number_Integer, 0},
	}, false, t)
}

func TestGoBuiltins(t *testing.T) {

	source := `string(foo) var bar string panic(baz)`

	tokens := GetTokens(goLexerInstance, []byte(source))

	checkTokens(tokens, []Token{
		{"string", Name_Builtin, 0},
		{"(", Punctuation, 0},
		{"foo", Name_Other, 0},
		{")", Punctuation, 0},
		{"var", Keyword_Declaration, 0},
		{"bar", Name_Other, 0},
		{"string", Keyword_Type, 0},
		{"panic", Name_Builtin, 0},
		{"(", Punctuation, 0},
		{"baz", Name_Other, 0},
		{")", Punctuation, 0},
	}, false, t)
}

func TestGoComments(t *testing.T) {

	source := `
/* this is 
multiline comment */

// this is single line comment
    `

	tokens := GetTokens(goLexerInstance, []byte(source))

	checkTokens(tokens, []Token{
		{"/* this is \nmultiline comment */", Comment_Multiline, 0},
		{"// this is single line comment", Comment_Single, 0},
	}, false, t)
}

func checkTokens(actual []Token, expected []Token, checkOffset bool, t *testing.T) {
	if len(actual) != len(expected) {
		fmt.Printf("%v\n", actual)
		t.Fatal(fmt.Sprintf("Expected %d tokens, got %d", len(expected), len(actual)))
	}
	for i := range actual {
		checkToken(actual[i], expected[i], checkOffset, i, t)
	}
}

func checkToken(actual Token, expected Token, checkOffset bool, pos int, t *testing.T) {
	if actual.Type != expected.Type ||
		actual.Text != expected.Text ||
		(checkOffset && (actual.Offset != expected.Offset)) {
		t.Fatal(fmt.Sprintf("Expected %s, got %s at %d", expected, actual, pos))
	}
}
