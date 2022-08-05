package querybuilder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/grafana/regexp"
)

func Test_peek(t *testing.T) {
	tests := []struct {
		pattern       string
		index, offset int
		match         byte
	}{
		{
			pattern: "test/a",
			index:   0,
			offset:  1,
			match:   'e',
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%s:%d", t.Name(), i), func(t *testing.T) {
			if peek(test.pattern, test.index, test.offset) != test.match {
				t.Error()
			}
		})
	}
}

func Test_findGroups(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []group
	}{
		{
			name:     "one group",
			pattern:  "te(s)t",
			expected: []group{{start: 2, end: 4, capturing: true, number: 1}},
		},
		{
			name:     "two groups",
			pattern:  "te(s)(t)",
			expected: []group{{start: 2, end: 4, capturing: true, number: 1}, {start: 5, end: 7, capturing: true, number: 2}},
		},
		{
			name:     "two groups with non-capturing group",
			pattern:  "te(s)(t)(?:asdf)",
			expected: []group{{start: 2, end: 4, capturing: true, number: 1}, {start: 5, end: 7, capturing: true, number: 2}, {start: 8, end: 15, capturing: false, number: 0}},
		},
		{
			name:     "two groups with non-capturing group and character class",
			pattern:  "te(s)(t)(?:asdf)[(]",
			expected: []group{{start: 2, end: 4, capturing: true, number: 1}, {start: 5, end: 7, capturing: true, number: 2}, {start: 8, end: 15, capturing: false, number: 0}},
		},
		{
			name:    "two groups with non-capturing group and character class and nested",
			pattern: "te(s)(t)(?:asdf)[(](())",
			expected: []group{
				{start: 2, end: 4, capturing: true, number: 1},
				{start: 5, end: 7, capturing: true, number: 2},
				{start: 8, end: 15, capturing: false, number: 0},
				{start: 20, end: 21, capturing: true, number: 4},
				{start: 19, end: 22, capturing: true, number: 3},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := findGroups(test.pattern)
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("unexpected indices (want/got):\n%v \n%v", test.expected, got)
			}
		})
	}
}

func Test_replaceCaptureGroupsWithString(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		text     string
		expected string
		maxGroup int
	}{
		{
			name:     "1",
			pattern:  `(\w+)-(\w+)`,
			text:     `cat-cow dog-bat`,
			expected: `(?:cat|dog)-(?:cow|bat)`,
			maxGroup: -1,
		},
		{
			name:     "2",
			pattern:  `(\w+)-(\w+)`,
			text:     `cat-cow dog-bat`,
			expected: `(\w+)-(\w+)`,
			maxGroup: 0,
		},
		{
			name:     "3",
			pattern:  `(\w+)-(\w+)`,
			text:     `cat-cow dog-bat`,
			expected: `(?:cat|dog)-(\w+)`,
			maxGroup: 1,
		},
		{
			name:     "4",
			pattern:  `(\w+)-(\w+)`,
			text:     `cat-cow dog-bat`,
			expected: `(?:cat|dog)-(?:cow|bat)`,
			maxGroup: 2,
		},
		{
			name:     "5",
			pattern:  `(\w+)-(?:\w+)-(\w+)`,
			text:     `cat-cow-camel`,
			expected: `(?:cat)-(?:\w+)-(?:camel)`,
			maxGroup: -1,
		},
		{
			name:     "6",
			pattern:  `(\w+)-(?:\w+)-(\w+)`,
			text:     `cat-cow-camel`,
			expected: `(?:cat)-(?:\w+)-(\w+)`,
			maxGroup: 1,
		},
		{
			name:     "7 ensure non-capturing groups don't count towards group numbers",
			pattern:  `(\w+)-(?:\w+)-(\w+)`,
			text:     `cat-cow-camel`,
			expected: `(?:cat)-(?:\w+)-(?:camel)`,
			maxGroup: 2,
		},
		{
			name:     "ensure literal values are escaped in the new pattern",
			pattern:  `(.*)`,
			text:     `\w`,
			expected: `(?:\\w)`,
			maxGroup: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reg, err := regexp.Compile(test.pattern)
			if err != nil {
				return
			}
			matches := reg.FindAllStringSubmatch(test.text, -1)
			groups := findGroups(test.pattern)
			if diff := cmp.Diff(test.expected, replaceCaptureGroupsWithString(test.pattern, groups, matches, test.maxGroup)); diff != "" {
				t.Errorf("unexpected pattern (want/got): %s", diff)
			}
		})
	}
}
