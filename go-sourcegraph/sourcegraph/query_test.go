package sourcegraph

import "testing"

func TestJoin(t *testing.T) {
	tests := []struct {
		tokens []Token
		want   RawQuery
	}{
		{
			tokens: []Token{},
			want:   RawQuery{Text: "", InsertionPoint: 0},
		},
		{
			tokens: []Token{Term("a"), Term("b")},
			want:   RawQuery{Text: "a b", InsertionPoint: 4},
		},
	}
	for _, test := range tests {
		q := Join(test.tokens)
		if q != test.want {
			t.Errorf("%v: got %v, want %v", test.tokens, q, test.want)
		}
	}
}
