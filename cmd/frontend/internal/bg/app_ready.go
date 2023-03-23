package bg

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// AppReady is called once the frontend has reported it is ready to serve
// requests. It contains tasks related to Sourcegraph App (single binary).
func AppReady(db database.DB, logger log.Logger) {
	if !deploy.IsApp() {
		return
	}

	ctx := context.Background()

	// Our goal is to open the browser to a special sign-in URL containing a
	// nonce (signInURL). We additionally want to display the URL without the
	// nonce since it can only be used once (displayURL)
	displayURL := globals.ExternalURL().String()
	browserURL := displayURL

	if signInURL, err := userpasswd.AppSiteInit(ctx, logger, db); err != nil {
		logger.Error("failed to initialize app user account", log.Error(err))
	} else {
		browserURL = signInURL
	}

	// See cmd/frontend/graphqlbackend/site_flags.go:needsRepositoryConfiguration
	// There is technically a small race condition where we need repository discovery
	// to finish before we decide whether or not to render the setup wizard.
	//
	// The impact of this race condition is very minimal (worst case scenario it
	// displays the setup wizard when the user doesn't need it to.) We sleep for a second
	// before opening the browser just to reduce the chance of it.
	//
	// https://github.com/sourcegraph/sourcegraph/pull/49820#issuecomment-1479959514
	time.Sleep(1 * time.Second)
	if err := browser.OpenURL(browserURL); err != nil {
		logger.Error("failed to open browser", log.String("url", browserURL), log.Error(err))
		// We failed to open the browser, so rather display that URL so the
		// user can click it.
		displayURL = browserURL
	}

	printExternalURL(displayURL)
}

func printExternalURL(externalURL string) {
	pad := func(s string, n int) string {
		spaces := n - len(s)
		if spaces < 0 {
			spaces = 0
		}
		return s + strings.Repeat(" ", spaces)
	}
	emptyLine := pad("", 76)
	newLine := "\033[0m\n"
	output := color.New(color.FgBlack, color.BgGreen, color.Bold)
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
	output.Fprintf(os.Stderr, "| %s |"+newLine, emptyLine)
	output.Fprintf(os.Stderr, "| %s |"+newLine, pad("Sourcegraph is now available on "+externalURL, 76))
	output.Fprintf(os.Stderr, "| %s |"+newLine, emptyLine)
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
}
