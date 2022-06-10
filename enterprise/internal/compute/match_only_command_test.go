package compute

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type serializer func(*MatchContext) any

func match(r *MatchContext) any {
	return r
}

func environment(r *MatchContext) any {
	env := make(map[string]string)
	for _, m := range r.Matches {
		for k, v := range m.Environment {
			env[k] = v.Value
		}
	}
	return env
}

func Test_matchOnly(t *testing.T) {
	data := &result.FileMatch{
		File: result.File{Path: "bedge", Repo: types.MinimalRepo{
			ID:   5,
			Name: "codehost.com/myorg/myrepo",
		}},
		ChunkMatches: result.ChunkMatches{{
			Content:      "abcdefgh",
			ContentStart: result.Location{Line: 1},
			Ranges: result.Ranges{{
				Start: result.Location{Line: 1},
				End:   result.Location{Line: 1},
			}},
		}},
	}

	test := func(input string, serialize serializer) string {
		r, _ := regexp.Compile(input)
		result := matchOnly(data, r)
		v, _ := json.MarshalIndent(serialize(result), "", "  ")
		return string(v)
	}

	autogold.Want("compute regexp submatch empty environment", `{
  "matches": [],
  "path": "bedge",
  "repositoryID": 5,
  "repository": "codehost.com/myorg/myrepo"
}`).Equal(t, test("nothing", match))

	autogold.Want("compute named regexp submatch", `{
  "1": "a",
  "ThisIsNamed": "b"
}`).Equal(t, test("(a)(?P<ThisIsNamed>b)", environment))

	autogold.Want("no slice out of bounds access on capture group", "{}").Equal(t, test("(lasvegans)|abcdefgh", environment))

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
        "1": {
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
        "2": {
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
        "3": {
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
        "4": {
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
  "path": "bedge",
  "repositoryID": 5,
  "repository": "codehost.com/myorg/myrepo"
}`).Equal(t, test("a(b(c))(de)f(g)h", match))

	autogold.Want("compute regexp submatch includes all matches on line", `{
  "matches": [
    {
      "value": "a",
      "range": {
        "start": {
          "offset": -1,
          "line": 1,
          "column": 0
        },
        "end": {
          "offset": -1,
          "line": 1,
          "column": 1
        }
      },
      "environment": {
        "1": {
          "value": "a",
          "range": {
            "start": {
              "offset": -1,
              "line": 1,
              "column": 0
            },
            "end": {
              "offset": -1,
              "line": 1,
              "column": 1
            }
          }
        }
      }
    },
    {
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
      },
      "environment": {
        "1": {
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
  "path": "bedge",
  "repositoryID": 5,
  "repository": "codehost.com/myorg/myrepo"
}`).Equal(t, test("([ag])", match))

}
