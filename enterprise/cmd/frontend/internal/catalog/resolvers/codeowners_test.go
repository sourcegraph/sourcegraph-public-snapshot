package resolvers

import (
	"reflect"
	"strconv"
	"testing"
)

func TestParseCodeowners(t *testing.T) {
	tests := []struct {
		file string
		want []codeownersEntry
	}{
		{
			file: `
# a
a/*/b @c @d
./x y@z.com`,
			want: []codeownersEntry{
				{pathPattern: "a/*/b", owners: []string{"@c", "@d"}},
				{pathPattern: "x", owners: []string{"y@z.com"}},
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := parseCodeowners([]byte(test.file))
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}
}
