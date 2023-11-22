package keyword

import (
	"testing"

	"github.com/hexops/autogold/v2"

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
		"!?", // punctuation-only token should be removed
	}
	wantPatterns := []string{
		"comput",
		"cluster",
		"int",
		"string",
		"elaps",
		"timer",
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
			wantQuery:    autogold.Expect("type:file context:global abc"),
			wantPatterns: autogold.Expect([]string{"abc"}),
		},
		{
			query:        "abc def",
			wantQuery:    autogold.Expect("type:file (abc OR def)"),
			wantPatterns: autogold.Expect([]string{"abc", "def"}),
		},
		{
			query:        "context:global lang:Go how to unzip file",
			wantQuery:    autogold.Expect("type:file context:global lang:Go (unzip OR file)"),
			wantPatterns: autogold.Expect([]string{"unzip", "file"}),
		},
		{
			query:        "K MEANS CLUSTERING in python",
			wantQuery:    autogold.Expect("type:file (cluster OR python)"),
			wantPatterns: autogold.Expect([]string{"cluster", "python"}),
		},
		{
			query:     `outer content:"inner {with} (special) ^characters$ and keywords like file or repo"`,
			wantQuery: autogold.Expect("type:file (special OR ^characters$ OR keyword OR file OR repo OR outer)"),
			wantPatterns: autogold.Expect([]string{
				"special", "^characters$", "keyword", "file",
				"repo",
				"outer",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			keywordQuery, err := queryStringToKeywordQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if keywordQuery == nil {
				t.Fatal("keywordQuery == nil")
			}

			tt.wantPatterns.Equal(t, keywordQuery.patterns)
			tt.wantQuery.Equal(t, query.StringHuman(keywordQuery.query.ToParseTree()))
		})
	}
}
