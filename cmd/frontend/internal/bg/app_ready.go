package bg

import (
	"context"
	"net/url"
	"os"
	"strings"

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

	externalURL := appSignInURL()

	if err := userpasswd.AppSiteInit(ctx, logger, db); err != nil {
		logger.Error("failed to initialize app user account", log.Error(err))
	}

	printExternalURL(externalURL)
	if err := browser.OpenURL(externalURL); err != nil {
		logger.Error("failed to open browser", log.String("url", externalURL), log.Error(err))
	}
}

func appSignInURL() string {
	externalURL := globals.ExternalURL().String()
	u, err := url.Parse(externalURL)
	if err != nil {
		return externalURL
	}
	nonce, err := userpasswd.AppNonce.Value()
	if err != nil {
		return externalURL
	}
	u.Path = "/sign-in"
	query := u.Query()
	query.Set("nonce", nonce)
	u.RawQuery = query.Encode()
	return u.String()
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
