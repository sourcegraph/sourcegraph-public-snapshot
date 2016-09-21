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
	}
	for _, c := range cases {
		got := parseSymbolQuery(c.Query)
		if !reflect.DeepEqual(got, c.Want) {
			t.Errorf("parseSymbolQuery(%+#v): got %+v, want %+v", c.Query, got, c.Want)
		}
	}
}
