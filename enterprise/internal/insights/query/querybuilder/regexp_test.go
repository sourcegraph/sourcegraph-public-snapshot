package querybuilder

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

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

func TestThing(t *testing.T) {
	pattern := `name:\((.*)\)(.*) [(] asdf`

	text := `name:(test1)bob ( asdf`

	reg, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	matches := reg.FindStringSubmatch(text)
	println(len(matches))
	for _, match := range matches {
		fmt.Println(match)
	}

	groups := findGroups(pattern)

	var replacements []group
	for i := range groups {
		if !groups[i].capturing {
			continue
		}
		groups[i].value = matches[groups[i].number]
		replacements = append(replacements, groups[i])
	}

	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].number < replacements[j].number
	})

	fmt.Println(groups)
	fmt.Println(fmt.Sprintf("old_pattern: %s", pattern))
	fmt.Println(fmt.Sprintf("document: %s", text))
	fmt.Println(fmt.Sprintf("new_pattern: %s", replaceRange(pattern, replacements)))
}
