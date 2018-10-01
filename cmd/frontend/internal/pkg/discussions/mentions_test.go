package discussions

import (
	"reflect"
	"testing"
)

func TestParseMentions(t *testing.T) {
	tests := []struct {
		name, input string
		want        []string
	}{
		{
			name:  "basic",
			input: "@bob",
			want:  []string{"bob"},
		},
		{
			name:  "complex",
			input: "hello @sally world\t@bob-1233%#!%#$1\n@kim",
			want:  []string{"sally", "bob-1233%#!%#$1", "kim"},
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := parseMentions(tst.input)
			if !reflect.DeepEqual(got, tst.want) {
				t.Fatalf("got %q want %q", got, tst.want)
			}
		})
	}
}
