package querybuilder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"

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
			name:     "no groups in pattern",
			pattern:  `\w*\s`,
			expected: nil,
		},
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
		pattern string
		text    string
		want    autogold.Value
	}{
		{
			pattern: `(\w+)-(\w+)`,
			text:    `cat-cow dog-bat`,
			want:    autogold.Want("1", "(?:cat)-(\\w+)"),
		},
		{
			pattern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cat-cow-camel`,
			want:    autogold.Want("middle non-capturing group", "(?:cat)-(?:\\w+)-(\\w+)"),
		},
		{
			pattern: `(\w+)-(?:\w+)-(\w+)`,
			text:    `cat-cow-camel`,
			want:    autogold.Want("ensure non-capturing groups don't count towards group numbers", "(?:cat)-(?:\\w+)-(\\w+)"),
		},
		{
			pattern: `(.*)`,
			text:    `\w`,
			want:    autogold.Want("ensure literal values are escaped in the new pattern", "(?:\\\\w)"),
		},
		{
			pattern: `\w{3}(.{3})\w{3}`,
			text:    `foobardog`,
			want:    autogold.Want("fixed repeat pattern", "\\w{3}(?:bar)\\w{3}"),
		},
	}
	for _, test := range tests {
		t.Run(test.want.Name(), func(t *testing.T) {
			reg, err := regexp.Compile(test.pattern)
			if err != nil {
				return
			}
			matches := reg.FindStringSubmatch(test.text)
			value := matches[1]

			groups := findGroups(test.pattern)
			got := replaceCaptureGroupsWithString(test.pattern, groups, value)
			test.want.Equal(t, got)
		})
	}

	t.Run("test explicitly a regexp with no groups", func(t *testing.T) {
		pattern := `replaceme`
		got := replaceCaptureGroupsWithString(pattern, nil, "no")
		require.Equal(t, pattern, got)
	})

	t.Run("regexp with no capturing groups", func(t *testing.T) {
		pattern := `(?:hello)(?:friend)`
		got := replaceCaptureGroupsWithString(pattern, findGroups(pattern), "no")
		require.Equal(t, pattern, got)
	})
}

func TestReplace_Valid(t *testing.T) {
	tests := []struct {
		query       string
		replacement string
		want        autogold.Value
		searchType  query.SearchType
	}{
		{
			query:       "/replaceme/",
			replacement: "replace",
			want:        autogold.Want("replace_1", BasicQuery("/replace/")),
			searchType:  query.SearchTypeStandard,
		},
		{
			query:       "/replace(me)/",
			replacement: "you",
			want:        autogold.Want("replace_2", BasicQuery("/replace(?:you)/")),
			searchType:  query.SearchTypeStandard,
		},
		{
			query:       "/replaceme/",
			replacement: "replace",
			want:        autogold.Want("replace_3", BasicQuery("/replace/")),
			searchType:  query.SearchTypeLucky,
		},
		{
			query:       "/replace(me)/",
			replacement: "you",
			want:        autogold.Want("replace_4", BasicQuery("/replace(?:you)/")),
			searchType:  query.SearchTypeLucky,
		},
		{
			query:       "/b(u)tt(er)/",
			replacement: "e",
			want:        autogold.Want("ensure only one group is replaced", BasicQuery("/b(?:e)tt(er)/")),
			searchType:  query.SearchTypeStandard,
		},
		{
			query:       "/b(?:u)(tt)(er)/",
			replacement: "dd",
			want:        autogold.Want("ensure only one group is replaced after non-capturing group", BasicQuery("/b(?:u)(?:dd)(er)/")),
			searchType:  query.SearchTypeStandard,
		},
		{
			query:       "replaceme",
			replacement: "replace",
			want:        autogold.Want("regexp_type_1", BasicQuery("/replace/")),
			searchType:  query.SearchTypeRegex,
		},
		{
			query:       "replace(me)",
			replacement: "you",
			want:        autogold.Want("regexp_type_2", BasicQuery("/replace(?:you)/")),
			searchType:  query.SearchTypeRegex,
		},
		{
			query:       `\/insight[s]\/`,
			replacement: "you",
			want:        autogold.Want("escaped slashes in regexp without group", BasicQuery("/you/")),
			searchType:  query.SearchTypeRegex,
		},
		{
			query:       `\/insi(g)ht[s]\/`,
			replacement: "ggg",
			want:        autogold.Want("escaped slashes in regexp with group", BasicQuery("/\\/insi(?:ggg)ht[s]\\//")),
			searchType:  query.SearchTypeRegex,
		},
	}
	for _, test := range tests {
		t.Run(test.want.Name(), func(t *testing.T) {
			replacer, err := NewPatternReplacer(BasicQuery(test.query), test.searchType)
			require.NoError(t, err)

			got, err := replacer.Replace(test.replacement)
			test.want.Equal(t, got)
		})
	}
}

func TestReplace_Invalid(t *testing.T) {
	t.Run("multiple patterns", func(t *testing.T) {
		_, err := NewPatternReplacer("/replace(me)/ or asdf", query.SearchTypeStandard)
		require.ErrorIs(t, err, MultiplePatternErr)
	})
	t.Run("literal pattern", func(t *testing.T) {
		_, err := NewPatternReplacer("asdf", query.SearchTypeStandard)
		require.ErrorIs(t, err, UnsupportedPatternTypeErr)
	})
	t.Run("no pattern", func(t *testing.T) {
		_, err := NewPatternReplacer("", query.SearchTypeRegex)
		require.ErrorIs(t, err, UnsupportedPatternTypeErr)
	})
	t.Run("filters with no pattern", func(t *testing.T) {
		_, err := NewPatternReplacer("repo:repoA rev:3.40.0", query.SearchTypeStandard)
		require.ErrorIs(t, err, UnsupportedPatternTypeErr)
	})
}
