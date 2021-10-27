package compute

import (
	"regexp"
	"testing"

	"github.com/hexops/autogold"
)

func Test_replace(t *testing.T) {
	test := func(input string, cmd *Replace) string {
		result, err := replace([]byte(input), cmd.MatchPattern, cmd.ReplacePattern)
		if err != nil {
			return err.Error()
		}
		return result.Value
	}

	autogold.Want(
		"regexp search replace",
		"needs a bit more queryrunner").
		Equal(t, test("needs more queryrunner", &Replace{
			MatchPattern:   &Regexp{Value: regexp.MustCompile(`more (\w+)`)},
			ReplacePattern: "a bit more $1",
		}))
}
