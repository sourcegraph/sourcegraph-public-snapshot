package coverageutil

import "testing"

type test struct {
	descr    string
	text     string
	expected []Token
}

var tests = []test{
	test{
		descr: "test array literal",
		text:  `[3, "three", true, null]`,
		expected: []Token{
			Token{1, `3`}, Token{5, `three`}, Token{13, `true`}, Token{19, `null`}}},
	test{
		descr: "test object",
		text:  `{"obj": [{"a": 1}]}`,
		expected: []Token{
			Token{2, `obj`}, Token{11, `a`}, Token{15, `1`}}},
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
