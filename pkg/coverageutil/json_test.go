package coverageutil

import "testing"

type testJSON struct {
	descr    string
	text     string
	expected []Token
}

var tests = []testJSON{
	testJSON{
		descr: "test array literal",
		text:  `[3, "three", true, null]`,
		expected: []Token{
			Token{1, 1, `3`}, Token{5, 1, `three`}, Token{13, 1, `true`}, Token{19, 1, `null`}}},
	testJSON{
		descr: "test object",
		text:  `{"obj": [{"a": 1}]}`,
		expected: []Token{
			Token{2, 1, `obj`}, Token{11, 1, `a`}, Token{15, 1, `1`}}},

	testJSON{
		descr: "test line break",
		text:  `{"obj": [{"a":` + "\n" + ` 1}]}`,
		expected: []Token{
			Token{2, 1, `obj`}, Token{11, 1, `a`}, Token{16, 2, `1`}}},
}

func TestJSON(t *testing.T) {

	for _, test := range tests {
		tokenizer := jsonTokenizer{}
		tokenizer.Init([]byte(test.text))

		var output []*Token
		for {
			token := tokenizer.Next()
			if token == nil {
				break
			}
			output = append(output, token)
		}

		if len(output) != len(test.expected) {
			t.Fatalf("%s: Token lengths differ Expected: %v, %v Actual: %v, %v",
				test.text, len(test.expected), test.expected, len(output), output)
		}

		for i, elem := range output {
			if elem.Offset != test.expected[i].Offset || elem.Text != test.expected[i].Text {
				t.Fatalf("%s: Token Output Incorrect - Expected: %v+ Actual: %v+", test.descr, test.expected[i], elem)
			}
		}

	}

}

func TestJSONError(t *testing.T) {
	broken := `{}}`
	tokenizer := jsonTokenizer{}
	tokenizer.Init([]byte(broken))

	if tokenizer.Next() != nil {
		t.Fatalf("Expected an error when trying to parse unbalanced JSON: %s", broken)
	}

	if len(tokenizer.Errors()) == 0 {
		t.Fatalf("Expected a non-zero number of errors returned when trying to parse unbalanced JSON: %s", broken)
	}

}
