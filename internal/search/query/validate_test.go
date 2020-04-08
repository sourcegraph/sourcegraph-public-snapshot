package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_PartitionSearchPattern(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "x",
			want:  `"x"`,
		},
		{
			input: "x y",
			want:  `(concat "x" "y")`,
		},
		{
			input: "x or y",
			want:  `(or "x" "y")`,
		},
		{
			input: "x and y",
			want:  `(and "x" "y")`,
		},
		{
			input: "file:foo x y",
			want:  `"file:foo" (concat "x" "y")`,
		},
		{
			input: "file:foo (x y)",
			want:  `"file:foo" (concat "(x" "y)")`,
		},
		{
			input: "(file:foo x) y",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
		{
			input: "file:foo (x and y)",
			want:  `"file:foo" (and "x" "y")`,
		},
		{
			input: "file:foo x and y",
			want:  `"file:foo" (and "x" "y")`,
		},
		{
			input: "file:foo (x or y)",
			want:  `"file:foo" (or "x" "y")`,
		},
		{
			input: "file:foo x or y",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
		{
			input: "(file:foo x) or y",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
		{
			input: "file:foo and content:x",
			want:  `"file:foo" "content:x"`,
		},
		{
			input: "repo:foo and file:bar and x",
			want:  `"repo:foo" "file:bar" "x"`,
		},
		{
			input: "repo:foo and (file:bar or file:baz) and x",
			want:  "cannot evaluate: unable to partition pure search pattern",
		},
	}
	for _, tt := range cases {
		t.Run("partition search pattern", func(t *testing.T) {
			q, _ := ParseAndOr(tt.input)
			andOrQuery, _ := q.(*AndOrQuery)
			scopeParameters, pattern, err := PartitionSearchPattern(andOrQuery.Query)
			if err != nil {
				if diff := cmp.Diff(tt.want, err.Error()); diff != "" {
					t.Fatal(diff)
				}
				return
			}
			result := append(scopeParameters, pattern)
			var resultStr []string
			for _, node := range result {
				resultStr = append(resultStr, node.String())
			}
			got := strings.Join(resultStr, " ")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
