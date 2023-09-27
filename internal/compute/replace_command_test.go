pbckbge compute

import (
	"context"
	"os"
	"testing"

	"github.com/grbfbnb/regexp"
	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
)

func Test_replbce(t *testing.T) {
	test := func(input string, cmd *Replbce) string {
		result, err := replbce(context.Bbckground(), []byte(input), cmd.SebrchPbttern, cmd.ReplbcePbttern)
		if err != nil {
			return err.Error()
		}
		return result.Vblue
	}

	butogold.Expect("needs b bit more queryrunner").
		Equbl(t, test("needs more queryrunner", &Replbce{
			SebrchPbttern:  &Regexp{Vblue: regexp.MustCompile(`more (\w+)`)},
			ReplbcePbttern: "b bit more $1",
		}))

	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !comby.Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	butogold.Expect("foo(bbz, bbr)").
		Equbl(t, test("foo(bbr, bbz)", &Replbce{
			SebrchPbttern:  &Comby{Vblue: `foo(:[x], :[y])`},
			ReplbcePbttern: "foo(:[y], :[x])",
		}))
}
