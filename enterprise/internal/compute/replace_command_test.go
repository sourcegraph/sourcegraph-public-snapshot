package compute

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold"

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

	autogold.Want(
		"regexp search replace",
		"needs a bit more queryrunner").
		Equal(t, test("needs more queryrunner", &Replace{
			SearchPattern:  &Regexp{Value: regexp.MustCompile(`more (\w+)`)},
			ReplacePattern: "a bit more $1",
		}))

	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	autogold.Want(
		"structural search replace",
		"foo(baz, bar)").
		Equal(t, test("foo(bar, baz)", &Replace{
			SearchPattern:  &Comby{Value: `foo(:[x], :[y])`},
			ReplacePattern: "foo(:[y], :[x])",
		}))
}

func TestRunReplaceRepoMetadata(t *testing.T) {
	test := func(q string, m result.Match) *Text {
		defer gitserver.ResetMocks()
		computeQuery, _ := Parse(q)
		res, err := computeQuery.Command.Run(context.Background(), database.NewMockDB(), m)
		if err != nil {
			t.Error(err)
		}
		return res.(*Text)
	}

	autogold.Want(
		"verify repo metadata exists",
		&Text{
			Value: "abc", Kind: "replace-in-place", RepositoryID: 11,
			Repository: "my/awesome/repo",
		}).
		Equal(t, test(`content:replace(anything -> abc)`, fileMatch("anything")))
}
