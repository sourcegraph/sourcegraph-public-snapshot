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
			nil,
		},
		{
			"exported",
			&symbolQuery{
				Type:     "exported",
				Packages: []string{},
			},
		},
		{
			"external foo/bar",
			&symbolQuery{
				Type:     "external",
				Packages: []string{"foo/bar"},
			},
		},
		{
			"exported foo/bar foo/baz",
			&symbolQuery{
				Type:     "exported",
				Packages: []string{"foo/bar", "foo/baz"},
			},
		},
	}
	for _, c := range cases {
		got, err := parseSymbolQuery(c.Query)
		if c.Want == nil {
			if err == nil {
				t.Errorf("expected error for query %+#v", c.Query)
			}
			continue
		}
		if err != nil {
			t.Errorf("did not expect error for query %+#v: %s", c.Query, err)
		}
		if !reflect.DeepEqual(got, c.Want) {
			t.Errorf("parseSymbolQuery(%+#v): got %+v, want %+v", c.Query, got, c.Want)
		}
	}
}
