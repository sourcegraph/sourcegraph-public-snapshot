package golang

import (
	"reflect"
	"testing"
)

func TestParseSymbolQuery(t *testing.T) {
	cases := []struct {
		Query string
		Want  *symbolQuery
	}{
		{
			"",
			&symbolQuery{
				Type:   queryTypeAll,
				Tokens: []string{},
			},
		},
		{
			"foo/bar",
			&symbolQuery{
				Type:   queryTypeAll,
				Tokens: []string{"foo/bar"},
			},
		},
		{
			"is:exported",
			&symbolQuery{
				Type:   queryTypeExported,
				Tokens: []string{},
			},
		},
		{
			"is:exported foo/bar",
			&symbolQuery{
				Type:   queryTypeExported,
				Tokens: []string{"foo/bar"},
			},
		},
		{
			"is:external-ref foo/bar",
			&symbolQuery{
				Type:   queryTypeExternalRef,
				Tokens: []string{"foo/bar"},
			},
		},
	}
	for _, c := range cases {
		got := parseSymbolQuery(c.Query)
		if !reflect.DeepEqual(got, c.Want) {
			t.Errorf("parseSymbolQuery(%+#v): got %+v, want %+v", c.Query, got, c.Want)
		}
	}
}

func TestPathHasPrefix(t *testing.T) {
	cases := []struct {
		Path, Prefix string
		Want         bool
	}{
		{Path: "github.com/golang/go-tools", Prefix: "github.com/golang/go", Want: false},
		{Path: "github.com/golang/go/tools", Prefix: "github.com/golang/go", Want: true},
	}
	for _, c := range cases {
		got := pathHasPrefix(c.Path, c.Prefix)
		if got != c.Want {
			t.Errorf("pathHasPrefix(%q, %q): got %v want %v", c.Path, c.Prefix, got, c.Want)
		}
	}
}
