package localcodehost

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []*schema.LocalRepoPattern
	}{
		{
			name:  "Empty input",
			input: "",
			want:  []*schema.LocalRepoPattern{},
		},
		{
			name:  "Single entry w/ group",
			input: "first/pattern group",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: "group"},
			},
		},
		{
			name:  "Single entry w/o group",
			input: "first/pattern",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: ""},
			},
		},
		{
			name:  "Multiple entries",
			input: "first/pattern group\nsecond/pattern\nthird/pattern group2",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: "group"},
				{Pattern: "second/pattern", Group: ""},
				{Pattern: "third/pattern", Group: "group2"},
			},
		},
		{
			name:  "Trims whitespaces around pattern and groups",
			input: "first/pattern group    \n   second/pattern \n     third/pattern group2\t",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: "group"},
				{Pattern: "second/pattern", Group: ""},
				{Pattern: "third/pattern", Group: "group2"},
			},
		},
		{
			name:  "Ignores empty lines",
			input: "first/pattern group\n\n  \nsecond/pattern group2",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: "group"},
				{Pattern: "second/pattern", Group: "group2"},
			},
		},
		{
			name:  "Ignores comment lines",
			input: "first/pattern group\n# This is a comment\nsecond/pattern group2",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern", Group: "group"},
				{Pattern: "second/pattern", Group: "group2"},
			},
		},
		{
			name:  "Properly handles escaped spaces in pattern",
			input: "\"first/pattern with/space\" group\n\"second/pattern with/space\"",
			want: []*schema.LocalRepoPattern{
				{Pattern: "first/pattern with/space", Group: "group"},
				{Pattern: "second/pattern with/space", Group: ""},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConfig(strings.NewReader(tt.input))
			if len(result) != len(tt.want) {
				t.Fatalf("expected %d entries, got %d entries", len(tt.want), len(result))
			}

			for i, entry := range tt.want {
				if entry.Pattern != result[i].Pattern {
					t.Errorf("expected pattern `%s`, got `%s`", entry.Pattern, result[i].Pattern)
				}
				if entry.Group != result[i].Group {
					t.Errorf("expected group `%s`, got `%s`", entry.Group, result[i].Group)
				}
			}
		})
	}
}
