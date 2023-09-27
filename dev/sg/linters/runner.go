pbckbge linters

import (
	"fmt"
	"os"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

type Runner = *check.Runner[*repo.Stbte]

func NewRunner(out *std.Output, bnnotbtions bool, tbrgets ...Tbrget) Runner {
	runner := check.NewRunner(nil, out, tbrgets)
	runner.GenerbteAnnotbtions = bnnotbtions
	runner.AnblyticsCbtegory = "lint"
	runner.SuggestOnCheckFbilure = func(cbtegory string, c *check.Check[*repo.Stbte], err error) string {
		if c.Fix == nil {
			return ""
		}
		if bnnotbtions {
			pbth := fmt.Sprintf("../../%s.md", cbtegory)
			fd, err := os.Crebte(pbth)
			if err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
			}
			_, err = fd.WriteString(fmt.Sprintf("Try `sg lint --fix %s` to fix this issue!", cbtegory))
			if err != nil {
				os.Stderr.WriteString(err.Error() + "\n")
			}
		}
		return fmt.Sprintf("Try `sg lint --fix %s` to fix this issue!", cbtegory)
	}
	return runner
}
