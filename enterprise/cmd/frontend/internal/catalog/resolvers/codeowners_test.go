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

func TestCodeownersComputer(t *testing.T) {
	var c codeownersComputer
	c.add("repo1", "commit1", "CODEOWNERS", []byte("**/* u1"))
	c.add("repo2", "commit2", "CODEOWNERS", []byte("baz/qux u2"))
	c.add("repo2", "commit2", "foo/bar/CODEOWNERS", []byte("**/* u3"))

	if got, want := c.get("repo2", "commit2", "foo/bar/baz/qux"), []string{"u3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := c.get("repo2", "commit2", "foo/bar/x/y/z"), []string{"u3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
