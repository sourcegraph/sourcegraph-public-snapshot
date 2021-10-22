package compute

import (
	"regexp"
	"testing"

	"github.com/hexops/autogold"
)

func Test_doReplaceInPlace(t *testing.T) {
	test := func(input string, op *ReplaceInPlace) string {
		result, err := doReplaceInPlace([]byte(input), op)
		if err != nil {
			return err.Error()
		}
		return result.Value
	}

	autogold.Want(
		"regexp search replace",
		"needs a bit more queryrunner").
		Equal(t, test("needs more queryrunner", &ReplaceInPlace{
			MatchPattern:   &Regexp{Value: regexp.MustCompile(`more (\w+)`)},
			ReplacePattern: "a bit more $1",
		}))
}
