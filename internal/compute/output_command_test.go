package compute

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
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

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Want(
		"structural search output",
		`train(regional, intercity)
train(commuter, lightrail)`).
		Equal(t, test("Im a train. train(intercity, regional). choo choo. train(lightrail, commuter)", &Output{
			MatchPattern:  &Comby{Value: `train(:[x], :[y])`},
			OutputPattern: "train(:[y], :[x])",
		}))
}

func contentAsFileMatch(data string) *result.FileMatch {
	return &result.FileMatch{
		File: result.File{Path: "my/awesome/path"},
	}
}

func TestRun(t *testing.T) {
	test := func(q, content string) string {
		git.Mocks.ReadFile = func(_ api.CommitID, _ string) ([]byte, error) {
			return []byte(content), nil
		}
		defer git.ResetMocks()
		computeQuery, _ := Parse(q)
		res, err := computeQuery.Command.Run(context.Background(), contentAsFileMatch(content))
		if err != nil {
			return err.Error()
		}
		return res.(*Text).Value
	}

	autogold.Want(
		"template substitution regexp",
		"(1)\n(2)\n(3)\n").
		Equal(t, test(`content:output((\d) -> ($1))`, "a 1 b 2 c 3"))

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Want(
		"template substitution structural",
		">bar<").
		Equal(t, test(`content:output.structural(foo(:[arg]) -> >:[arg]<)`, "foo(bar)"))

}
