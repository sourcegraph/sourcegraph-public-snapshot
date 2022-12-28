package querybuilder

import (
	"testing"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold"
)

func FuzzTest_replaceCaptureGroupsWithString(f *testing.F) {
	tests := []struct {
		pattern string
		text    string
		want    autogold.Value
	}{
		{
			pattern: `(\w+)-(\w+)`,
			text:    `cat-cow dog-bat`,
		},
		{
			pattern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cat-cow-camel`,
		},
		{
			pattern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cat-cow-camel`,
		},
		{
			pattern: `(.*)`,
			text:    `\w`,
		},
		{
			pattern: `\w{3}(.{3})\w{3}`,
			text:    `foobardog`,
		},
	}
	for _, test := range tests {
		f.Add(test.pattern)
	}
	f.Fuzz(func(t *testing.T, pattern string) {
		reg, err := regexp.Compile(pattern)
		matches := reg.FindStringSubmatch("sometextwith need to match one because you know why")

		if len(matches) > 1 {
			groups := findGroups(pattern)
			_ = replaceCaptureGroupsWithString(pattern, groups, matches[1])
		}
	})
}
