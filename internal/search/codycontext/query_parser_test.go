package codycontext

import (
	"reflect"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestTransformPattern(t *testing.T) {
	patterns := []string{
		"compute",
		"K",     // very short terms should be removed
		"Means", // stop words should be removed
		"Clustering",
		"implement", // common code-related terms should be removed
		"int",
		"to",
		"string",
		"finding",
		"\"time",    // leading punctuation should be removed
		"elapsed\"", // trailing punctuation should be removed
		"using",
		"a",
		"timer",
		"computing",
		"own", // key terms should not be removed, even if they are common
		"!?",  // punctuation-only token should be removed
		"grf::causal_forest",
		"indexData.scoreFile",
		"resource->GetServer()",
		"foo-bar",
		"bas>quz",
	}
	wantPatterns := []string{
		"comput",
		"cluster",
		"int",
		"string",
		"elaps",
		"timer",
		"own",
		"grf",
		"causal_forest",
		"indexdata",
		"scorefil",
		"resourc",
		"getserv",
		"foo-bar",
		"bas>quz",
	}

	gotPatterns := transformPatterns(patterns)
	autogold.Expect(wantPatterns).Equal(t, gotPatterns)
}

func TestQueryStringToKeywordQuery(t *testing.T) {
	tests := []struct {
		query        string
		wantQuery    autogold.Value
		wantPatterns autogold.Value
	}{
		{
			query:        "context:global abc",
			wantQuery:    autogold.Expect("context:global abc"),
			wantPatterns: autogold.Expect([]string{"abc"}),
		},
		{
			query:        "abc def",
			wantQuery:    autogold.Expect("(abc OR def)"),
			wantPatterns: autogold.Expect([]string{"abc", "def"}),
		},
		{
			query:        "context:global lang:Go how to unzip file",
			wantQuery:    autogold.Expect("context:global lang:Go (unzip OR file)"),
			wantPatterns: autogold.Expect([]string{"unzip", "file"}),
		},
		{
			query:        "K MEANS CLUSTERING in python",
			wantQuery:    autogold.Expect("(cluster OR python)"),
			wantPatterns: autogold.Expect([]string{"cluster", "python"}),
		},
		{
			query:        "context:global the who",
			wantQuery:    autogold.Expect("context:global"),
			wantPatterns: autogold.Expect([]string{}),
		},
		{
			query:        "context:global grf::causal_forest",
			wantQuery:    autogold.Expect("context:global (grf OR causal_forest)"),
			wantPatterns: autogold.Expect([]string{"grf", "causal_forest"}),
		},
		{
			query:     `outer content:"inner {with} (special) ^characters$ and keywords like file or repo"`,
			wantQuery: autogold.Expect("(special OR ^characters$ OR keyword OR file OR repo OR outer)"),
			wantPatterns: autogold.Expect([]string{
				"special", "^characters$", "keyword", "file",
				"repo",
				"outer",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			q, err := parseQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if q == nil {
				t.Fatal("q == nil")
			}

			tt.wantPatterns.Equal(t, q.patterns)
			tt.wantQuery.Equal(t, query.StringHuman(q.keywordQuery.ToParseTree()))
		})
	}
}

func TestFindSymbols(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{
			name:     "simple patterns",
			patterns: []string{"fooBar", "baz", "foo_bar"},
			want:     []string{"fooBar", "foo_bar"},
		},
		{
			name:     "dotted patterns",
			patterns: []string{"foo.bar_baz", "baz_baz.quux", "foo.BarBaz", "end.start"},
			want:     []string{"bar_baz", "baz_baz", "BarBaz"},
		},
		{
			name:     "namespaced patterns",
			patterns: []string{"ns::foo", "bar::baz", "baz_quux", "some->field", "other>field"},
			want:     []string{"ns", "foo", "bar", "baz", "baz_quux", "some", "field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findSymbols(tt.patterns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandQuery(t *testing.T) {
	cases := []struct {
		queryString string
		want        []string
	}{
		{
			queryString: "What does this repo do?",
			want:        []string{kwReadme},
		},
		{
			queryString: "Describe my code",
			want:        []string{kwReadme},
		},
		{
			queryString: "Explain what this project is about",
			want:        []string{kwReadme},
		},
		// Negative tests
		{
			queryString: "what does cmd/update.sh do?",
			want:        nil,
		},
		{
			queryString: "what does update.sh do?",
			want:        nil,
		},
		{
			queryString: "explain update.sh",
			want:        nil,
		},
		{
			queryString: "What is the meaning of life?",
			want:        nil,
		},
		{
			queryString: "Where is my code?",
			want:        nil,
		},
	}

	for _, c := range cases {
		t.Run(c.queryString, func(t *testing.T) {
			got := expandQuery(c.queryString)
			require.Equal(t, c.want, got)
		})
	}
}
