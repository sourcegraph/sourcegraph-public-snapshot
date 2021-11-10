package compute

import (
	"context"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_output(t *testing.T) {
	test := func(input string, cmd *Output) string {
		result, err := output(context.Background(), input, cmd.MatchPattern, cmd.OutputPattern, cmd.Separator)
		if err != nil {
			return err.Error()
		}
		return result.Value
	}

	autogold.Want(
		"regexp search outputs only digits",
		"(1)~(2)~(3)~").
		Equal(t, test("a 1 b 2 c 3", &Output{
			MatchPattern:  &Regexp{Value: regexp.MustCompile(`(\d)`)},
			OutputPattern: "($1)",
			Separator:     "~",
		}))
}

func contentAsFileMatch(data string) *result.FileMatch {
	return &result.FileMatch{
		File: result.File{Path: "my/awesome/path"},
		LineMatches: []*result.LineMatch{
			{
				Preview: data,
			},
		},
	}
}

func TestRun(t *testing.T) {
	test := func(q, content string) string {
		computeQuery, _ := Parse(q)
		res, err := computeQuery.Command.Run(context.Background(), contentAsFileMatch(content))
		if err != nil {
			return err.Error()
		}
		return res.(*Text).Value
	}

	autogold.Want(
		"template substitution",
		"(1)\n(2)\n(3)\n").
		Equal(t, test(`content:output((\d) -> ($1))`, "a 1 b 2 c 3"))
}
