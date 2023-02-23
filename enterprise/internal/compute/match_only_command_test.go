package compute

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"

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

type want struct {
	Input  string
	Result any
}

func Test_matchOnly(t *testing.T) {
	content := "abcdefgh\n123\n!@#"
	data := &result.FileMatch{
		File: result.File{Path: "bedge", Repo: types.MinimalRepo{
			ID:   5,
			Name: "codehost.com/myorg/myrepo",
		}},
		ChunkMatches: result.ChunkMatches{{
			Content:      content,
			ContentStart: result.Location{Offset: 0, Line: 1, Column: 0},
			Ranges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 1, Column: 0},
				End:   result.Location{Offset: len(content), Line: 1, Column: len(content)},
			}},
		}},
	}

	test := func(input string, serialize serializer) string {
		r, _ := regexp.Compile(input)
		matchContext := matchOnly(data, r)
		w := want{Input: input, Result: serialize(matchContext)}
		v, _ := json.MarshalIndent(w, "", "  ")
		return string(v)
	}

	cases := []struct {
		input      string
		serializer serializer
	}{
		{input: "nothing", serializer: match},
		{input: "(a)(?P<ThisIsNamed>b)", serializer: environment},
		{input: "(lasvegans)|abcdefgh", serializer: environment},
		{input: "a(b(c))(de)f(g)h", serializer: match},
		{input: "([ag])", serializer: match},
		{input: "g(h(?:(?:.|\n)+)@)#", serializer: match},
		{input: "g(h\n1)23\n!@", serializer: match},
	}

	for _, c := range cases {
		t.Run("match_only", func(t *testing.T) {
			autogold.ExpectFile(t, autogold.Raw(test(c.input, c.serializer)))
		})
	}
}
