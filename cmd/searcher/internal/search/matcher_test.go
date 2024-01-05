package search

import (
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"testing"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt/query"
	"github.com/stretchr/testify/require"
)

func TestLongestLiteral(t *testing.T) {
	cases := map[string]string{
		"foo":       "foo",
		"FoO":       "FoO",
		"(?m:^foo)": "foo",
		"(?m:^FoO)": "FoO",
		"[Z]":       "Z",

		`\wddSuballocation\(dump`:    "ddSuballocation(dump",
		`\wfoo(\dlongest\wbam)\dbar`: "longest",

		`(foo\dlongest\dbar)`:  "longest",
		`(foo\dlongest\dbar)+`: "longest",
		`(foo\dlongest\dbar)*`: "",

		"(foo|bar)":     "",
		"[A-Z]":         "",
		"[^A-Z]":        "",
		"[abB-Z]":       "",
		"([abB-Z]|FoO)": "",
		`[@-\[]`:        "",
		`\S`:            "",
	}

	metaLiteral := "AddSuballocation(dump->guid(), system_allocator_name)"
	cases[regexp.QuoteMeta(metaLiteral)] = metaLiteral

	for expr, want := range cases {
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			t.Fatal(expr, err)
		}
		re = re.Simplify()
		got := longestLiteral(re)
		if want != got {
			t.Errorf("longestLiteral(%q) == %q != %q", expr, got, want)
		}
	}
}

func TestToZoektQuery(t *testing.T) {
	m := &andMatcher{
		children: []matcher{
			&orMatcher{
				children: []matcher{
					&regexMatcher{
						re:        regexp.MustCompile("aaaaa"),
						isNegated: true,
					},
					&regexMatcher{
						re: regexp.MustCompile("bbbb*"),
					},
				},
			},
			&regexMatcher{
				re:         regexp.MustCompile("cccc?"),
				ignoreCase: true,
			},
		},
	}

	cases := []struct {
		name         string
		matchContent bool
		matchPath    bool
		want         string
	}{{
		name:         "matches content only",
		matchContent: true,
		matchPath:    false,
		want:         `(and (or (not case_regex:"aaaaa") case_regex:"bbbb*") regex:"cccc?")`,
	},{
		name:         "matches path only",
		matchContent: false,
		matchPath:    true,
		want:         `(and (or (not case_file_regex:"aaaaa") case_file_regex:"bbbb*") file_regex:"cccc?")`,
	},
	{
		name:         "matches content and path",
		matchContent: true,
		matchPath:    true,
		want:         `(and (or (not case_regex:"aaaaa") (not case_file_regex:"aaaaa") case_regex:"bbbb*" case_file_regex:"bbbb*") (or regex:"cccc?" file_regex:"cccc?"))`,
	},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := m.ToZoektQuery(c.matchContent, c.matchPath)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, c.want, query.Simplify(got).String())
		})
	}
}

func TestMatchesString(t *testing.T) {
	m := &orMatcher{
		children: []matcher{
			&andMatcher{
				children: []matcher{
					&regexMatcher{
						re:        regexp.MustCompile("aaa"),
						isNegated: true,
					},
					&regexMatcher{
						re: regexp.MustCompile("bbb*"),
					},
					&regexMatcher{
						re: regexp.MustCompile("ccc*"),
					},
				},
			},
			&regexMatcher{
				re:         regexp.MustCompile("ddd?"),
				ignoreCase: true,
			},
		},
	}

	cases := []struct {
		name      string
		test      string
		wantMatch bool
	}{
		{
			name:      "first 'or' clause matches",
			test:      "bbbbbccc",
			wantMatch: true,
		},
		{
			name:      "negated 'and' clause does not match",
			test:      "aaabbbccc",
			wantMatch: false,
		},
		{
			name:      "case insensitive 'or' clause matches",
			test:      "zzzzzDDD",
			wantMatch: true,
		},
	}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			match := m.MatchesString(c.test)
			require.Equal(t, c.wantMatch, match)
		})
	}
}

func TestMatchesFile(t *testing.T) {
	m := &orMatcher{
		children: []matcher{
			&andMatcher{
				children: []matcher{
					&regexMatcher{
						re:        regexp.MustCompile("and"),
						isNegated: true,
					},
					&regexMatcher{
						re: regexp.MustCompile("file"),
					},
					&regexMatcher{
						re: regexp.MustCompile("the"),
					},
				},
			},
			&regexMatcher{
				re:         regexp.MustCompile("here"),
				ignoreCase: true,
			},
		},
	}

	cases := []struct {
		m matcher
		file        string
		wantMatch   bool
		wantMatches int
	}{
		{
			m: m,
			file:      "this is the first file",
			wantMatch: true,
			wantMatches: 2,
		},
		{
			m: m,
			file:      "here is the second fileee",
			wantMatch: true,
			wantMatches: 3,
		},
		{
			m: m,
			file:      "... and another file!",
			wantMatch: false,
			wantMatches: 0,
		},
		{
			m: &regexMatcher{
				re: regexp.MustCompile("excluded"),
				isNegated: true,
			},
			file:      "here's a file",
			// this matches, but produces no matched ranges
			wantMatch: true,
			wantMatches: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			match, matches := c.m.MatchesFile([]byte(c.file), 1000)
			require.Equal(t, c.wantMatch, match)
			require.Equal(t, c.wantMatches, len(matches))
		})
	}
}

func TestMatchesFileLimits(t *testing.T) {
	file := []byte("the file that mentions file a lot ... the file file")

	cases := []struct {
		name        string
		m           matcher
		limit       int
		wantMatches int
	}{
		{
			name: "'or' matcher with limit",
			m: &orMatcher{
				children: []matcher{
					&regexMatcher{
						re: regexp.MustCompile("file"),
					},
					&regexMatcher{
						re: regexp.MustCompile("the"),
					},
				},
			},
			limit:       3,
			wantMatches: 3,
		},
		{
			name: "'or' matcher without limit",
			m: &orMatcher{
				children: []matcher{
					&regexMatcher{
						re: regexp.MustCompile("file"),
					},
					&regexMatcher{
						re: regexp.MustCompile("the"),
					},
				},
			},
			limit:       10,
			wantMatches: 6,
		},
		{
			name: "'and' matcher with limit",
			m: &andMatcher{
				children: []matcher{
					&regexMatcher{
						re: regexp.MustCompile("file"),
					},
					&regexMatcher{
						re: regexp.MustCompile("the"),
					},
				},
			},
			limit:       3,
			wantMatches: 3,
		},
		{
			name: "'and' matcher with negation",
			m: &andMatcher{
				children: []matcher{
					&regexMatcher{
						re: regexp.MustCompile("excluded"),
						isNegated: true,
					},
					&regexMatcher{
						re: regexp.MustCompile("file"),
					},
					&regexMatcher{
						re: regexp.MustCompile("the"),
					},
				},
			},
			limit:       3,
			wantMatches: 3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, matches := c.m.MatchesFile(file, c.limit)
			require.Equal(t, c.wantMatches, len(matches))
		})
	}
}
