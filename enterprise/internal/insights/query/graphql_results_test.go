package query

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCommitSearchResultMatchCount(t *testing.T) {
	t.Run("with matches", func(t *testing.T) {
		results := commitSearchResult{
			Matches: []struct{ Highlights []struct{ Line int } }{
				{Highlights: []struct{ Line int }{{Line: 1}, {Line: 2}}}, // this match has 2 highlights and should contribute 2
				{Highlights: []struct{ Line int }{}},                     // this match has no highlights and should contribute 1
			},
		}
		want := 3
		got := results.MatchCount()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected CommitSearchMatchCount (want/got): %v", diff)
		}
	})
	t.Run("with no matches", func(t *testing.T) {
		results := commitSearchResult{}
		want := 0
		got := results.MatchCount()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected CommitSearchMatchCount (want/got): %v", diff)
		}
	})
}

func TestCommitSearchResultFromJson(t *testing.T) {
	var res *GqlSearchResponse
	if err := json.NewDecoder(strings.NewReader(realCommitSearch)).Decode(&res); err != nil {
		t.Error(err)
	}
	for _, raw := range res.Data.Search.Results.Results {
		var commitSearchResult commitSearchResult
		if err := json.Unmarshal(raw, &commitSearchResult); err != nil {
			t.Error(err)
		}
		got := commitSearchResult.MatchCount()
		want := 1
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (want/got): %v", diff)
		}
	}
}

const realCommitSearch = `
{
  "data": {
    "search": {
      "results": {
        "limitHit": false,
        "cloning": [],
        "missing": [],
        "timedout": [],
        "matchCount": 2,
        "results": [
          {
            "__typename": "CommitSearchResult",
            "matches": [
              {
                "highlights": [
                  {
                    "line": 1
                  }
                ]
              }
            ],
            "commit": {
              "repository": {
                "id": "UmVwb3NpdG9yeToxODM3Ng==",
                "name": "codehost.org/sourcegraph/somerepo"
              }
            }
          },
          {
            "__typename": "CommitSearchResult",
            "matches": [
              {
                "highlights": [
                  {
                    "line": 1
                  }
                ]
              }
            ],
            "commit": {
              "repository": {
                "id": "UmVwb3NpdG9yeToxODM3Ng==",
                "name": "codehost.org/sourcegraph/somerepo"
              }
            }
          }
        ],
        "alert": null
      }
    }
  }
}
`
