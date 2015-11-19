package search

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		q       string
		want    []sourcegraph.Token
		wantErr error
	}{
		{q: "a", want: []sourcegraph.Token{sourcegraph.AnyToken("a")}},
		{q: "a b", want: []sourcegraph.Token{sourcegraph.AnyToken("a"), sourcegraph.AnyToken("b")}},
		{q: "o/r", want: []sourcegraph.Token{sourcegraph.RepoToken{URI: "o/r"}}},
		{q: "o/", want: []sourcegraph.Token{sourcegraph.RepoToken{URI: "o/"}}},
		{q: ":v", want: []sourcegraph.Token{sourcegraph.RevToken{Rev: "v"}}},
		{q: ":cccccccccccccccccccccccccccccccccccccccc", want: []sourcegraph.Token{sourcegraph.RevToken{Rev: "cccccccccccccccccccccccccccccccccccccccc"}}},
		{q: "~", want: []sourcegraph.Token{sourcegraph.UnitToken{}}},
		{q: "~u", want: []sourcegraph.Token{sourcegraph.UnitToken{Name: "u"}}},
		{q: "~u@t", want: []sourcegraph.Token{sourcegraph.UnitToken{Name: "u", UnitType: "t"}}},
		{q: "~@t", want: []sourcegraph.Token{sourcegraph.UnitToken{UnitType: "t"}}},
		{q: "@u", want: []sourcegraph.Token{sourcegraph.UserToken{Login: "u"}}},
		{q: "@", want: []sourcegraph.Token{sourcegraph.UserToken{Login: ""}}},
		{
			q: `o/r @u :b a0 a1 t`,
			want: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "o/r"},
				sourcegraph.UserToken{Login: "u"},
				sourcegraph.RevToken{Rev: "b"},
				sourcegraph.AnyToken("a0"),
				sourcegraph.AnyToken("a1"),
				sourcegraph.AnyToken("t"),
			},
		},
	}
	for _, test := range tests {
		label := fmt.Sprintf("<< %v >>", test.q)
		tokens, _, err := tokenize(test.q)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: tokenize: %s", label, err)
			} else {
				t.Errorf("%s: tokenize: got error %q, want %q", label, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(tokens, test.want) {
			t.Errorf("%s: got tokens %v, want %v", label, debugFormatTokens(tokens), debugFormatTokens(test.want))
		}
	}
}

func TestScan_tokenPos(t *testing.T) {
	tests := []struct {
		q    string
		want tokenPos
	}{
		{q: "a", want: tokenPos{[2]int{0, 1}}},
		{q: "a b", want: tokenPos{[2]int{0, 1}, [2]int{2, 3}}},
		{q: `abbbb`, want: tokenPos{[2]int{0, 5}}},
	}
	for _, test := range tests {
		label := fmt.Sprintf("<< %v >>", test.q)
		_, tpos, err := scan(test.q)
		if err != nil {
			t.Errorf("%s: Scan: %s", label, err)
			continue
		}
		if !reflect.DeepEqual(tpos, test.want) {
			t.Errorf("%s: got token positions %v, want %v", label, tpos, test.want)
		}
	}
}

func debugFormatTokens(toks []sourcegraph.Token) string {
	tokStrs := make([]string, len(toks))
	for i, tok := range toks {
		tokStrs[i] = strings.TrimPrefix(strings.TrimPrefix(fmt.Sprintf("%#v", tok), "*"), "query.")
	}
	return "[" + strings.Join(tokStrs, ", ") + "]"
}
