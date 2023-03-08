package bg

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

// AppReady is called once the frontend has reported it is ready to serve
// requests. It contains tasks related to Sourcegraph App (single binary).
func AppReady(logger log.Logger) {
	if !deploy.IsDeployTypeSingleProgram(deploy.Type()) {
		return
	}

	externalURL := globals.ExternalURL().String()
	printRunningAddress(externalURL)
	if err := browser.OpenURL(externalURL); err != nil {
		logger.Error("failed to open browser", log.String("url", externalURL), log.Error(err))
	}
}

func printRunningAddress(externalURL string) {
	pad := func(s string, n int) string {
		spaces := n - len(s)
		if spaces < 0 {
			spaces = 0
		}
		return s + strings.Repeat(" ", spaces)
	}

	newLine := "\033[0m\n"
	output := color.New(color.FgBlack, color.BgGreen, color.Bold)
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
	output.Fprintf(os.Stderr, "| %s |"+newLine, pad("", 76))
	output.Fprintf(os.Stderr, "| %s |"+newLine, pad("Sourcegraph is now available on "+externalURL, 76))
	output.Fprintf(os.Stderr, "| %s |"+newLine, pad("", 76))
	output.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
}
