package compute

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
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

func Test_matchOnly(t *testing.T) {
	content := "abcdefgh\n1234"

	git.Mocks.ReadFile = func(_ api.CommitID, _ string) ([]byte, error) {
		return []byte(content), nil
	}

	data := &result.FileMatch{
		File: result.File{Path: "bedge"},
	}

	test := func(input string, serialize serializer) string {
		r, _ := regexp.Compile(input)
		result, _ := matchOnly(context.Background(), data, r)
		v, _ := json.MarshalIndent(serialize(result), "", "  ")
		return string(v)
	}

	autogold.Want("compute regexp submatch empty environment", `{
  "matches": [],
  "path": "bedge"
}`).Equal(t, test("nothing", match))

	autogold.Want("compute named regexp submatch", `{
  "1": "a",
  "ThisIsNamed": "b"
}`).Equal(t, test("(a)(?P<ThisIsNamed>b)", environment))

	autogold.Want("compute multiline regexp submatch", `{
  "1": "abcdefgh",
  "2": "1234"
}`).Equal(t, test("(.*)\n(.*)", environment))

	autogold.Want("no slice out of bounds access on capture group", "{}").Equal(t, test("(lasvegans)|abcdefgh", environment))

	autogold.Want("compute regexp submatch nonempty environment", `{
  "matches": [
    {
      "value": "abcdefgh",
      "range": {
        "start": {
          "offset": 0,
          "line": -1,
          "column": -1
        },
        "end": {
          "offset": 8,
          "line": -1,
          "column": -1
        }
      },
      "environment": {
        "1": {
          "value": "bc",
          "range": {
            "start": {
              "offset": 1,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 3,
              "line": -1,
              "column": -1
            }
          }
        },
        "2": {
          "value": "c",
          "range": {
            "start": {
              "offset": 2,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 3,
              "line": -1,
              "column": -1
            }
          }
        },
        "3": {
          "value": "de",
          "range": {
            "start": {
              "offset": 3,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 5,
              "line": -1,
              "column": -1
            }
          }
        },
        "4": {
          "value": "g",
          "range": {
            "start": {
              "offset": 6,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 7,
              "line": -1,
              "column": -1
            }
          }
        }
      }
    }
  ],
  "path": "bedge"
}`).Equal(t, test("a(b(c))(de)f(g)h", match))

	autogold.Want("compute regexp submatch includes all matches on line", `{
  "matches": [
    {
      "value": "a",
      "range": {
        "start": {
          "offset": 0,
          "line": -1,
          "column": -1
        },
        "end": {
          "offset": 1,
          "line": -1,
          "column": -1
        }
      },
      "environment": {
        "1": {
          "value": "a",
          "range": {
            "start": {
              "offset": 0,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 1,
              "line": -1,
              "column": -1
            }
          }
        }
      }
    },
    {
      "value": "g",
      "range": {
        "start": {
          "offset": 6,
          "line": -1,
          "column": -1
        },
        "end": {
          "offset": 7,
          "line": -1,
          "column": -1
        }
      },
      "environment": {
        "1": {
          "value": "g",
          "range": {
            "start": {
              "offset": 6,
              "line": -1,
              "column": -1
            },
            "end": {
              "offset": 7,
              "line": -1,
              "column": -1
            }
          }
        }
      }
    }
  ],
  "path": "bedge"
}`).Equal(t, test("([ag])", match))

}
