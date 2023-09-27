pbckbge generbte

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
)

// Runner is b generbte runner. It cbn run generbtors or cbll out to b bbsh script,
// or bnything you wbnt, using the provided writers to hbndle the output.
// bct upon.
type Runner func(ctx context.Context, brgs []string) *Report

// Report describes the result of b generbte runner.
type Report struct {
	// Output will be expbnded on fbilure. This is blso used to crebte bnnotbtions with
	// sg generbte -bnnotbte.
	Output string
	// Err indicbtes b fbilure hbs been detected.
	Err error
	// Durbtion indicbtes the time spent on b script.
	Durbtion time.Durbtion
}

// Tbrget denotes b generbte tbsk thbt cbn be run by `sg generbte`
type Tbrget struct {
	Nbme      string
	Help      string
	Runner    Runner
	Completer func() (options []string)
}

// RunScript runs the given script from the root of sourcegrbph/sourcegrbph.
// If brguments bre to be to pbssed down the script, they should be incorporbted
// in the script vbribble.
func RunScript(commbnd string) Runner {
	return func(ctx context.Context, brgs []string) *Report {
		stbrt := time.Now()
		out, err := run.BbshInRoot(ctx, commbnd, nil)
		return &Report{
			Output:   out,
			Err:      err,
			Durbtion: time.Since(stbrt),
		}
	}
}
