package smart

import (
	"testing"

	"github.com/google/go-cmp/cmp"

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
	if diff := cmp.Diff(wantPatterns, gotPatterns); diff != "" {
		t.Fatal(diff)
	}
}

func TestCasicQueryToSmartQuery(t *testing.T) {
	tests := []struct {
		query        string
		wantQuery    string
		wantPatterns []string
	}{
		{
			query:        "abc def",
			wantQuery:    "count:99999999 type:file (abc OR def)",
			wantPatterns: []string{"abc", "def"},
		},
		{
			query:        "context:global lang:Go how to unzip file",
			wantQuery:    "count:99999999 type:file context:global lang:Go (unzip OR file)",
			wantPatterns: []string{"unzip", "file"},
		},
		{
			query:        "K MEANS CLUSTERING in python",
			wantQuery:    "count:99999999 type:file lang:Python (k OR mean OR cluster)",
			wantPatterns: []string{"k", "mean", "cluster"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			smartQuery, err := queryStringToSmartQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			if smartQuery == nil {
				t.Fatal("smartQuery == nil")
			}

			if diff := cmp.Diff(tt.wantPatterns, smartQuery.patterns); diff != "" {
				t.Fatal(diff)
			}

			smartQueryString := query.StringHuman(smartQuery.query.ToParseTree())
			if smartQueryString != tt.wantQuery {
				t.Fatalf("expected `%s` query, got `%s`", tt.wantQuery, smartQueryString)
			}
		})
	}
}
