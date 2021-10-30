package compute

import (
	"context"
	"regexp"
	"testing"

	"github.com/hexops/autogold"
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
