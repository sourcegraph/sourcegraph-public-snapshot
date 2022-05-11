package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/download"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var updateCommand = &cli.Command{
	Name:  "update",
	Usage: "Update local sg installation",
	Description: `Update local sg installation with the latest changes. To see what's new, run:

  sg version changelog -next

Requires a local copy of the 'sourcegraph/sourcegraph' codebase.`,
	Category: CategoryUtil,
	Action: func(cmd *cli.Context) error {
		if _, err := updateToPrebuiltSG(cmd.Context); err != nil {
			return err
		}
		std.Out.WriteSuccessf("sg has been updated!")
		std.Out.Write("To see what's new, run 'sg version changelog'.")
		return nil
	},
}

// updateToPrebuiltSG downloads the latest release of sg prebuilt binaries and install it.
func updateToPrebuiltSG(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://github.com/sourcegraph/sg/releases/latest", nil)
	if err != nil {
		return "", err
	}
	// We use the RountTripper to make an HTTP request without having to deal
	// with redirections.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("location")
	if location == "" {
		return "", errors.New("GitHub latest release: empty location")
	}
	location = strings.ReplaceAll(location, "/tag/", "/download/")
	downloadURL := fmt.Sprintf("%s/sg_%s_%s", location, runtime.GOOS, runtime.GOARCH)

	currentExecPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if err := download.Exeuctable(downloadURL, currentExecPath); err != nil {
		return "", err
	}
	return currentExecPath, nil
}
