package compute

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestOfLineMatches(t *testing.T) {
	data := &result.FileMatch{
		File: result.File{Path: "bedge"},
		LineMatches: []*result.LineMatch{
			{
				Preview:    "abcdefgh",
				LineNumber: 1,
			},
		},
	}

	test := func(input string) string {
		r, _ := regexp.Compile(input)
		result := ofFileMatches(data, r)
		v, _ := json.MarshalIndent(result, "", "  ")
		return string(v)
	}

	autogold.Want("compute regexp submatch environment", `{
  "matches": [
    {
      "value": "abcdefgh",
      "range": {
        "start": {
          "offset": -1,
          "line": 1,
          "column": 0
        },
        "end": {
          "offset": -1,
          "line": 1,
          "column": 8
        }
      },
      "environment": {
        "$1": {
          "value": "bc",
          "range": {
            "start": {
              "offset": -1,
              "line": 1,
              "column": 1
            },
            "end": {
              "offset": -1,
              "line": 1,
              "column": 3
            }
          }
        },
        "$2": {
          "value": "c",
          "range": {
            "start": {
              "offset": -1,
              "line": 1,
              "column": 2
            },
            "end": {
              "offset": -1,
              "line": 1,
              "column": 3
            }
          }
        },
        "$3": {
          "value": "de",
          "range": {
            "start": {
              "offset": -1,
              "line": 1,
              "column": 3
            },
            "end": {
              "offset": -1,
              "line": 1,
              "column": 5
            }
          }
        },
        "$4": {
          "value": "g",
          "range": {
            "start": {
              "offset": -1,
              "line": 1,
              "column": 6
            },
            "end": {
              "offset": -1,
              "line": 1,
              "column": 7
            }
          }
        }
      }
    }
  ],
  "path": "bedge"
}`).Equal(t, test("a(b(c))(de)f(g)h"))
}
