package searchquery

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantRemaining  string
		wantOperations [][2]string
	}{
		{
			name:           "basic",
			input:          `fuzzy title search file:mux.go involves:"bill jack sarah" order:oldest -repo:foo`,
			wantRemaining:  "fuzzy title search",
			wantOperations: [][2]string{{"file", "mux.go"}, {"involves", "bill jack sarah"}, {"order", "oldest"}, {"-repo", "foo"}},
		},
		{
			name:          "string_escape",
			input:         `foo:"bar 123\tefg 1:\"2:3"`,
			wantRemaining: "",
			wantOperations: [][2]string{{"foo", `bar 123	efg 1:"2:3`}},
		},
		{
			name:           "remaining_left_and_right",
			input:          `abc efg foo:bar 123 baz:"bam"456`,
			wantRemaining:  "abc efg 123 456",
			wantOperations: [][2]string{{"foo", "bar"}, {"baz", "bam"}},
		},
		{
			name:           "empty_operation",
			input:          `fuzzytitleprefixmatch: foo:bar`,
			wantRemaining:  "fuzzytitleprefixmatch:",
			wantOperations: [][2]string{{"foo", "bar"}},
		},
		{
			name:           "escaped_operation",
			input:          `fuzzytitleprefixmatch: foo\:bar`,
			wantRemaining:  "fuzzytitleprefixmatch: foo:bar",
			wantOperations: nil,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			remaining, operations := Parse(tst.input)
			if remaining != tst.wantRemaining {
				t.Logf("want remaining: %q", tst.wantRemaining)
				t.Logf("got  remaining: %q", remaining)
				t.Fail()
			}
			if !reflect.DeepEqual(operations, tst.wantOperations) {
				t.Logf("want operations: %q", tst.wantOperations)
				t.Logf("got  operations: %q", operations)
				t.Fail()
			}
		})
	}
}
