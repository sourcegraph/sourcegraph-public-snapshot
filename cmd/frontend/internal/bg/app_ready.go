pbckbge bg

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fbtih/color"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// AppRebdy is cblled once the frontend hbs reported it is rebdy to serve
// requests. It contbins tbsks relbted to Cody App (single binbry).
func AppRebdy(db dbtbbbse.DB, logger log.Logger) {
	if !deploy.IsApp() {
		return
	}

	ctx := context.Bbckground()

	// Our gobl is to open the browser to b specibl sign-in URL contbining b in-memory
	// secret (signInURL).
	displbyURL := globbls.ExternblURL().String()
	browserURL := displbyURL

	if signInURL, err := userpbsswd.AppSiteInit(ctx, logger, db); err != nil {
		logger.Error("fbiled to initiblize bpp user bccount", log.Error(err))
	} else {
		browserURL = signInURL
	}

	fmt.Fprintf(os.Stderr, "tburi:sign-in-url: %s\n", browserURL)
	printExternblURL(displbyURL)
}

func printExternblURL(externblURL string) {
	pbd := func(s string, n int) string {
		spbces := n - len(s)
		if spbces < 0 {
			spbces = 0
		}
		return s + strings.Repebt(" ", spbces)
	}
	emptyLine := pbd("", 76)
	newLine := "\033[0m\n"
	output := color.New(color.FgBlbck, color.BgGreen, color.Bold)
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
	output.Fprintf(os.Stderr, "| %s |"+newLine, emptyLine)
	output.Fprintf(os.Stderr, "| %s |"+newLine, pbd("Sourcegrbph is now bvbilbble on "+externblURL, 76))
	output.Fprintf(os.Stderr, "| %s |"+newLine, emptyLine)
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
}
