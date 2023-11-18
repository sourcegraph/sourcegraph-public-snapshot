package bg

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// AppReady is called once the frontend has reported it is ready to serve
// requests. It contains tasks related to Cody App (single binary).
func AppReady(db database.DB, logger log.Logger) {
	if !deploy.IsApp() {
		return
	}

	ctx := context.Background()

	// Our goal is to open the browser to a special sign-in URL containing a in-memory
	// secret (signInURL).
	displayURL := globals.ExternalURL().String()
	browserURL := displayURL

	if signInURL, err := userpasswd.AppSiteInit(ctx, logger, db); err != nil {
		logger.Error("failed to initialize app user account", log.Error(err))
	} else {
		browserURL = signInURL
	}

	fmt.Fprintf(os.Stderr, "tauri:sign-in-url: %s\n", browserURL)
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
