package keyword

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestTransformPattern(t *testing.T) {
	patterns := []string{
		"K",
		"Means",
		"Clustering",
		"convert",
		"int",
		"to",
		"string",
		"finding",
		"time",
		"elapsed",
		"using",
		"a",
		"timer",
	}
	wantPatterns := []string{
		"k",
		"mean",
		"cluster",
		"convert",
		"int",
		"string",
		"find",
		"time",
		"elaps",
		"using",
		"timer",
	}

	gotPatterns := transformPatterns(patterns)
	autogold.Want("transform keyword patterns", wantPatterns).Equal(t, gotPatterns)
}

func TestQueryStringToKeywordQuery(t *testing.T) {
	tests := []struct {
		query        string
		wantQuery    autogold.Value
		wantPatterns autogold.Value
	}{
		{
			query:        "context:global abc",
			wantQuery:    autogold.Want("one pattern query with global context", "count:99999999 type:file context:global abc"),
			wantPatterns: autogold.Want("patterns for one pattern query with global context", []string{"abc"}),
		},
		{
			query:        "abc def",
			wantQuery:    autogold.Want("two pattern query", "count:99999999 type:file (abc OR def)"),
			wantPatterns: autogold.Want("patterns for two pattern query", []string{"abc", "def"}),
		},
		{
			query:        "context:global lang:Go how to unzip file",
			wantQuery:    autogold.Want("query with existing filters", "count:99999999 type:file context:global lang:Go (unzip OR file)"),
			wantPatterns: autogold.Want("patterns for query with existing filters", []string{"unzip", "file"}),
		},
		{
			query:        "K MEANS CLUSTERING in python",
			wantQuery:    autogold.Want("query with language and uppercase patterns", "count:99999999 type:file lang:Python (k OR mean OR cluster)"),
			wantPatterns: autogold.Want("patterns for query with language and uppercase patterns", []string{"k", "mean", "cluster"}),
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
