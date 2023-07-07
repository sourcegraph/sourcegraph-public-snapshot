package compute

import (
	"context"
	"os"
	"testing"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/comby"
)

func Test_replace(t *testing.T) {
	test := func(input string, cmd *Replace) string {
		result, err := replace(context.Background(), []byte(input), cmd.SearchPattern, cmd.ReplacePattern)
		if err != nil {
			return err.Error()
		}
		return result.Value
	}

	autogold.Expect("needs a bit more queryrunner").
		Equal(t, test("needs more queryrunner", &Replace{
			SearchPattern:  &Regexp{Value: regexp.MustCompile(`more (\w+)`)},
			ReplacePattern: "a bit more $1",
		}))

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Expect("foo(baz, bar)").
		Equal(t, test("foo(bar, baz)", &Replace{
			SearchPattern:  &Comby{Value: `foo(:[x], :[y])`},
			ReplacePattern: "foo(:[y], :[x])",
		}))
}
