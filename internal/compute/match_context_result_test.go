package compute

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type serializer func(*MatchContext) interface{}

func match(r *MatchContext) interface{} {
	return r
}

func environment(r *MatchContext) interface{} {
	env := make(map[string]string)
	for _, m := range r.Matches {
		for k, v := range m.Environment {
			env[k] = v.Value
		}
	}
	return env
}

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

	test := func(input string, serialize serializer) string {
		r, _ := regexp.Compile(input)
		result := FromFileMatch(data, r)
		v, _ := json.MarshalIndent(serialize(result), "", "  ")
		return string(v)
	}

	autogold.Want("compute regexp submatch empty environment", `{
  "matches": [],
  "path": "bedge"
}`).Equal(t, test("nothing", match))

	autogold.Want("compute named regexp submatch", `{
  "$1": "a",
  "ThisIsNamed": "b"
}`).Equal(t, test("(a)(?P<ThisIsNamed>b)", environment))

	autogold.Want("compute regexp submatch nonempty environment", `{
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
}`).Equal(t, test("a(b(c))(de)f(g)h", match))
}
